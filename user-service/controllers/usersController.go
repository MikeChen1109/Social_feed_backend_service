package controllers

import (
	"net/http"
	appErrors "user-service/common/appErrors"
	"user-service/models"
	"user-service/services"

	"github.com/gin-gonic/gin"
)

type UsersController struct {
	AuthService *services.AuthService
}

// Signup godoc
// @Summary      User signup
// @Description  Register a new user with username and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  models.SignupRequest  true  "Signup info"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /user/signup [post]
func (usersController *UsersController) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.Bind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	err := usersController.AuthService.Signup(req.Username, req.Password)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User signed up successfully"})
}

// Login godoc
// @Summary      User login
// @Description  Login with username and password, returns access and refresh tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  models.LoginRequest  true  "Login info"
// @Success      200   {object}  models.TokenResponse
// @Failure      400
// @Failure      401
// @Failure      500
// @Router       /user/login [post]
func (usersController *UsersController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	token, refreshToken, err := usersController.AuthService.Login(req.Username, req.Password)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}

// Refresh godoc
// @Summary      Refresh token
// @Description  Refresh access token using refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  models.RefreshRequest  true  "Refresh token"
// @Success      200   {object}  models.TokenResponse
// @Failure      400
// @Failure      401
// @Failure      500
// @Router       /user/refresh [post]
func (usersController *UsersController) Refresh(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.Bind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	token, refreshToken, err := usersController.AuthService.Refresh(req.RefreshToken)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}

// Logout godoc
// @Summary      User logout
// @Description  Invalidate the refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  models.LogoutRequest  true  "Logout request"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /user/logout [post]
func (usersController *UsersController) Logout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.Bind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	usersController.AuthService.Logout(req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
}

func errorHandler(c *gin.Context, err error) {
	if appErr, ok := err.(*appErrors.AppError); ok {
		c.AbortWithStatusJSON(appErr.StatusCode, gin.H{"message": appErr.Message})
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
