package handlers

import (
	"api-gateway/utils"
	"os"

	"github.com/gin-gonic/gin"
)

func ProxyToUserService(c *gin.Context) {
	if os.Getenv("APP_ENV") == "dev" {
		// utils.ProxyRequest(c, "http://localhost:4000")
		utils.ProxyRequest(c, "http://user-service:4000")
	} else {
		utils.ProxyRequest(c, os.Getenv("FEED_SERVICE_URL"))
	}
}
