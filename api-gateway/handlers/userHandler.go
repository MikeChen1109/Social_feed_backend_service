package handlers

import (
	"api-gateway/utils"
	"os"

	"github.com/gin-gonic/gin"
)

func ProxyToUserService(c *gin.Context) {
	url := os.Getenv("USER_SERVICE_URL")
	if url == "" {
		url = "http://user-service:4000"
	}
	utils.ProxyRequest(c, url)
}
