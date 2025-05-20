package controllers

import (
	appErrors "user-service/common/appErrors"
	"user-service/services"
	"net/http"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// err := authService.Signup(body.UsearName, body.Password)
	err := usersController.AuthService.Signup(body.UsearName, body.Password)
	errorHandler(c, err)

	c.JSON(http.StatusOK, gin.H{"message": "User signed up successfully"})
}

func (usersController *UsersController) Login(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	token, err := usersController.AuthService.Login(body.Username, body.Password)
	errorHandler(c, err)

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func errorHandler(c *gin.Context, err error) {
	if err == nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
