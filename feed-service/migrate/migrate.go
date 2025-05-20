package main

import (
	"feed-service/initializers"
	"feed-service/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
}

func main() {
	db := initializers.ConnectToDatabase()
	// db.Migrator().DropTable(&models.Comment{}, &models.Feed{})
	db.AutoMigrate(&models.Feed{}, &models.Comment{})
}
