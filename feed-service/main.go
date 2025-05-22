package main

import (
	"feed-service/controllers"
	"feed-service/initializers"
	"feed-service/repositories"
	"feed-service/routes"
	"feed-service/services"
	"strings"

	"github.com/gin-contrib/cors"
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
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// 允許任意 localhost 開頭的來源（含動態 port）
			return strings.HasPrefix(origin, "http://localhost")
		},
		AllowOrigins:     []string{"https://feedapi.onrender.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "X-Requested-With", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.RegisterFeedRoutes(router, feedsController)
	routes.RegisterCommentRoutes(router, commentsController)
	router.Run()
}
