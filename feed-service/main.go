package main

import (
	"feed-service/controllers"
	"feed-service/initializers"
	"feed-service/middleware"
	"feed-service/repositories"
	"feed-service/routes"
	"feed-service/services"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	// database connection
	db := initializers.ConnectToDatabase()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Initialize repositories, services, controllers, and middleware
	feedsRepo := &repositories.FeedRepository{DB: db}
	commentsRepo := &repositories.CommentRepository{DB: db}
	feedService := &services.FeedService{FeedRepo: feedsRepo}
	commentService := &services.CommentService{CommentRepo: commentsRepo}
	feedsController := &controllers.FeedsController{FeedsService: feedService}
	commentsController := &controllers.CommentsController{CommentsService: commentService}

	// Initialize Gin router and register routes
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	routes.RegisterFeedRoutes(router, feedsController)
	routes.RegisterCommentRoutes(router, commentsController)
	router.Run()
}
