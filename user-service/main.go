package main

import (
	"os"
	"user-service/controllers"
	"user-service/initializers"
	"user-service/middleware"
	"user-service/repositories"
	"user-service/routes"
	"user-service/services"

	_ "user-service/docs"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	// database connection
	redisClient := initializers.ConnectRedis()
	db := initializers.ConnectToDatabase()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	defer redisClient.Close()

	// Initialize repositories, services, controllers, and middleware
	userRepo := &repositories.UserRepository{DB: db}
	tokenRepo := &repositories.TokenRepository{DB: db, Redis: redisClient}
	authService := &services.AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}
	userController := &controllers.UsersController{AuthService: authService}

	// Initialize Gin router and register routes
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.RegisterUserRoutes(router, userController)
	router.Run()
}
