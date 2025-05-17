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
	initializers.DB.Migrator().DropTable(&models.Comment{}, &models.Feed{}, &models.User{})
	initializers.DB.AutoMigrate(&models.Feed{}, &models.Comment{}, &models.User{})
}
