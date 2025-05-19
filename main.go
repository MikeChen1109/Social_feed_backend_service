package main

import (
	"myApp/SocialFeed/controllers"
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/middleware"
	"myApp/SocialFeed/repositories"
	"myApp/SocialFeed/routes"
	"myApp/SocialFeed/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
}

func main() {
	// database connection
	db := initializers.ConnectToDatabase()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Initialize repositories, services, controllers, and middleware
	userRepo := &repositories.UserRepository{DB: db}
	feedsRepo := &repositories.FeedRepository{DB: db}
	authService := &services.AuthService{UserRepo: userRepo}
	feedService := &services.FeedService{FeedRepo: feedsRepo}
	userController := &controllers.UsersController{AuthService: authService}
	feedsController := &controllers.FeedsController{FeedsService: feedService}
	authMiddleware := &middleware.AuthMiddleware{AuthService: authService}

	// Initialize Gin router and register routes
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://feedapi.onrender.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.RegisterFeedRoutes(router, authMiddleware, feedsController)
	routes.RegisterUserRoutes(router, userController)
	router.Run()
}
