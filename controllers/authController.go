package controllers

import (
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/repositories"
	"myApp/SocialFeed/services"
)

var authService *services.AuthService

func InitAuthController() {
	authService = &services.AuthService{
		UserRepo: &repositories.UserRepository{DB: initializers.DB},
	}
}
