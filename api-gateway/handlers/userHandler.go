package handlers

import (
	"api-gateway/utils"

	"github.com/gin-gonic/gin"
)

func ProxyToUserService(c *gin.Context) {
	utils.ProxyRequest(c, "http://user-service:4000")
}
