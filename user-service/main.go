package main

import (
	"user-service/controllers"
	"user-service/initializers"
	"user-service/repositories"
	"user-service/routes"
	"user-service/services"
	"strings"

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
	authService := &services.AuthService{UserRepo: userRepo}
	userController := &controllers.UsersController{AuthService: authService}

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

	routes.RegisterUserRoutes(router, userController)
	router.Run()
}
