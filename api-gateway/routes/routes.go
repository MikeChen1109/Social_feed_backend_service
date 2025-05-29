package routes

import (
	"api-gateway/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	api := r.Group("/api")

	api.Any("/feed/*proxyPath", handlers.ProxyToFeedService)
	api.Any("/user/*proxyPath", handlers.ProxyToUserService)
}
