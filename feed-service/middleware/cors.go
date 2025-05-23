package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN") // 讀取前端網址

	config := cors.Config{
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	if os.Getenv("APP_ENV") == "dev" {
		config.AllowOriginFunc = func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost")
		}
		config.MaxAge = 0
	} else {
		config.MaxAge = 12 * time.Hour
	}

	return cors.New(config)
}
