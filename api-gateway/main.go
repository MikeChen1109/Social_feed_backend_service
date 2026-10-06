package main

import (
	initializers "api-gateway/initalizers"
	"api-gateway/middleware"
	"api-gateway/routes"
	"context"
	"log"
	"net/http"
	"os"
	monitoring "social-feed/observability"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.New()
	metrics := monitoring.New("api-gateway")
	router.Use(metrics.Middleware(), gin.Recovery(), middleware.CORSMiddleware())
	router.GET("/metrics", metrics.Handler)
	routes.RegisterRoutes(router)
	router.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		client := &http.Client{Timeout: 2 * time.Second}
		for _, service := range []struct{ env, fallback string }{{"USER_SERVICE_URL", "http://user-service:4000"}, {"FEED_SERVICE_URL", "http://feed-service:3000"}} {
			base := os.Getenv(service.env)
			if base == "" {
				base = service.fallback
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/readyz", nil)
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: ":" + port, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
