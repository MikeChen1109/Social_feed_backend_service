package main

import (
	"log"
	"user-service/initializers"
	"user-service/models"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	db := initializers.ConnectToDatabase()
	// db.Migrator().DropTable(&models.User{})
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal(err)
	}
}
