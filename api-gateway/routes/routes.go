package routes

import (
	"api-gateway/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	api.Any("/feed/*proxyPath", handlers.ProxyToFeedService)
	api.Any("/user/*proxyPath", handlers.ProxyToUserService)
}
