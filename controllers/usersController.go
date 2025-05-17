package controllers

import (
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/models"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	var body struct {
		UsearName string `json:"username"`
		Password  string `json:"password"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	if body.UsearName == "" || body.Password == "" {
		c.JSON(400, gin.H{"error": "Username and password are required"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: body.UsearName,
		Password: string(hash),
	}

	// Check if user already exists
	var existingUser models.User
	result := initializers.DB.Where("username = ?", user.Username).First(&existingUser)
	if result.Error == nil {
		c.JSON(400, gin.H{"error": "Username already exists"})
		return
	} else if result.Error != nil && result.Error.Error() != "record not found" {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	result = initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(200, gin.H{"message": "User signed up successfully"})
}

func Login(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	if body.Username == "" || body.Password == "" {
		c.JSON(400, gin.H{"error": "Username and password are required"})
		return
	}

	var user models.User
	result := initializers.DB.Where("username = ?", body.Username).First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(404, gin.H{"error": "User not found"})
		} else {
			c.JSON(500, gin.H{"error": "Database error"})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid password"})
		return
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix() , // Token valid for 72 hours
	}).SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create token"})
		return
	}

	c.JSON(200, gin.H{"message": "Login successful", "token": token})
}
