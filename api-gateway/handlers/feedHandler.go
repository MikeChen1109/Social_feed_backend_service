package handlers

import (
	"api-gateway/utils"
	"os"

	"github.com/gin-gonic/gin"
)

func ProxyToFeedService(c *gin.Context) {
	url := os.Getenv("FEED_SERVICE_URL")
	if url == "" {
		url = "http://feed-service:3000"
	}
	utils.ProxyRequest(c, url)
}
