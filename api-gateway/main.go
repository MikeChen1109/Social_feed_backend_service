package main

import (
	initializers "api-gateway/initalizers"
	"api-gateway/routes"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	router := gin.Default()
	routes.RegisterRoutes(router)
	router.Run()
}
