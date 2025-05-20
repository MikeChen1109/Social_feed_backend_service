package controllers

import (
	"net/http"
	appErrors "user-service/common/appErrors"
	"user-service/services"

	"github.com/gin-gonic/gin"
)

type UsersController struct {
	AuthService *services.AuthService
}

func (usersController *UsersController) Signup(c *gin.Context) {
	var body struct {
		UsearName string `json:"username"`
		Password  string `json:"password"`
	}

	if err := c.Bind(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	err := usersController.AuthService.Signup(body.UsearName, body.Password)
	if err != nil {
		errorHandler(c, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User signed up successfully"})
}

func (usersController *UsersController) Login(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	token, refreshToken, err := usersController.AuthService.Login(body.Username, body.Password)
	if err != nil {
		errorHandler(c, err)
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "refreshToken": refreshToken})
}

func (usersController *UsersController) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.Bind(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	token, refreshToken, err := usersController.AuthService.Refresh(body.RefreshToken)
	if err != nil {
		errorHandler(c, err)
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "refreshToken": refreshToken})
}

func (usersController *UsersController) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.Bind(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	err := usersController.AuthService.Logout(body.RefreshToken)
	if err != nil {
		errorHandler(c, err)
	}
}

func errorHandler(c *gin.Context, err error) {
	if appErr, ok := err.(*appErrors.AppError); ok {
		c.AbortWithStatusJSON(appErr.StatusCode, gin.H{"message": appErr.Message})
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
