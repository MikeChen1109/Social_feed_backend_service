package main

import (
	"feed-service/controllers"
	"feed-service/initializers"
	"feed-service/middleware"
	"feed-service/repositories"
	"feed-service/routes"
	"feed-service/services"
	"os"

	"github.com/gin-gonic/gin"
	_ "feed-service/docs"
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
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	routes.RegisterFeedRoutes(router, feedsController)
	routes.RegisterCommentRoutes(router, commentsController)
	router.Run()
}
