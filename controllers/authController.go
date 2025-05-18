package controllers

import (
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/repositories"
	"myApp/SocialFeed/services"
)

var authService *services.AuthService

func InitAuthController() {
	DB := initializers.DB
	repo := repositories.UserRepository{DB: DB}
	authService = &services.AuthService{UserRepo: &repo}
}
