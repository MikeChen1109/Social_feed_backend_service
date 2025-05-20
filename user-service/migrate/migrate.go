package main

import (
	"user-service/initializers"
	"user-service/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
}

func main() {
	db := initializers.ConnectToDatabase()
	// db.Migrator().DropTable(&models.User{})
	db.AutoMigrate(&models.User{})
}
