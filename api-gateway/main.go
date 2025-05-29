package main

import (
	initializers "api-gateway/initalizers"
	"api-gateway/middleware"
	"api-gateway/routes"
	"os"

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
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	routes.RegisterRoutes(router)
	router.Run()
}
