package main

import (
	"user-service/controllers"
	"user-service/initializers"
	"user-service/middleware"
	"user-service/repositories"
	"user-service/routes"
	"user-service/services"

	"github.com/gin-gonic/gin"
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
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	routes.RegisterUserRoutes(router, userController)
	router.Run()
}
