package main

import (
	"myApp/SocialFeed/controllers"
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/middleware"
	"myApp/SocialFeed/repositories"
	"myApp/SocialFeed/routes"
	"myApp/SocialFeed/services"

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
	routes.RegisterFeedRoutes(router, authMiddleware, feedsController)
	routes.RegisterUserRoutes(router, userController)
	router.Run()
}
