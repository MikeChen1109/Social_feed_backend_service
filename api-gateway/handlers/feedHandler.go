package handlers

import (
	"api-gateway/utils"

	"github.com/gin-gonic/gin"
)

func ProxyToFeedService(c *gin.Context) {
	utils.ProxyRequest(c, "http://feed-service:3000")
}
