package main

import (
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
}

func main() {
	db := initializers.ConnectToDatabase()
	// db.Migrator().DropTable(&models.Comment{}, &models.Feed{}, &models.User{})
	db.AutoMigrate(&models.Feed{}, &models.Comment{}, &models.User{})
}
