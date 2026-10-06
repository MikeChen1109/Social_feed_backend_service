package main

import (
	"feed-service/initializers"
	"feed-service/models"
	"log"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	db := initializers.ConnectToDatabase()
	// db.Migrator().DropTable(&models.Comment{}, &models.Feed{})
	if err := db.AutoMigrate(&models.Feed{}, &models.Comment{}); err != nil {
		log.Fatal(err)
	}
}
