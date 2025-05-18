package services

import (
	"errors"
	"myApp/SocialFeed/models"
	"myApp/SocialFeed/repositories"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo repositories.UserRepositoryInterface
}

type AppError struct {
	StatusCode int
	Message    string
}

func NewAppError(status int, msg string) *AppError {
	return &AppError{StatusCode: status, Message: msg}
}

func (s *AuthService) Login(username, password string) (string, *AppError) {
	if username == "" || password == "" {
		return "", NewAppError(http.StatusBadRequest, "Username and password are required")
	}

	user, err := s.UserRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return "", NewAppError(http.StatusNotFound, "User not found")
		}
		return "", NewAppError(http.StatusInternalServerError, "Database error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", NewAppError(http.StatusUnauthorized, "Invalid password")
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		return "", NewAppError(http.StatusInternalServerError, "Token generation failed")
	}

	return token, nil
}

func (s *AuthService) Signup(username, password string) *AppError {
	if username == "" || password == "" {
		return NewAppError(http.StatusBadRequest, "Username and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return NewAppError(http.StatusInternalServerError, "Failed to hash password")
	}

	user := models.User{
		Username: username,
		Password: string(hash),
	}

	existingUser, err := s.UserRepo.FindByUsername(username)
	
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return NewAppError(http.StatusInternalServerError, "Database error")
	}

	if existingUser != nil {
		return NewAppError(http.StatusConflict, "Username already exists")
	}

	if s.UserRepo.Create(&user) != nil {
		return NewAppError(http.StatusInternalServerError, "Failed to create user")
	}

	return nil
}

func generateJWT(userID uint) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	}).SignedString([]byte(os.Getenv("SECRET")))
}
