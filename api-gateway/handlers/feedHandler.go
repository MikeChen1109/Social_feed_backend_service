package handlers

import (
	"api-gateway/utils"
	"os"

	"github.com/gin-gonic/gin"
)

func ProxyToFeedService(c *gin.Context) {
	if os.Getenv("APP_ENV") == "dev" {
		utils.ProxyRequest(c, "http://localhost:3000")
	} else {
		utils.ProxyRequest(c, os.Getenv("FEED_SERVICE_URL"))
	}
}
