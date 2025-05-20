package services

import (
	"errors"
	"os"
	"time"
	appErrors "user-service/common/appErrors"
	"user-service/models"
	"user-service/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo  repositories.UserRepositoryInterface
	TokenRepo repositories.TokenRepositoryInterface
}

func (s *AuthService) Login(username, password string) (string, string, *appErrors.AppError) {
	if username == "" || password == "" {
		return "", "", appErrors.ErrInvalidUsernameOrPassword
	}

	user, err := s.UserRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return "", "", appErrors.ErrUserNotFound
		}
		return "", "", appErrors.DatabaseError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", appErrors.ErrInvalidPassword
	}

	token, err := generateJWT(user)

	if err != nil {
		return "", "", appErrors.ErrTokenGenerationFailed
	}

	refreshToken := uuid.NewString()
	storeErr := s.TokenRepo.StoreRefreshToken(refreshToken, user.ID)
	if storeErr != nil {
		return "", "", appErrors.ErrRefreshTokenStoreFailed
	}

	return token, refreshToken, nil
}

func (s *AuthService) Signup(username, password string) *appErrors.AppError {
	if username == "" || password == "" {
		return appErrors.ErrInvalidUsernameOrPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return appErrors.ErrFailedToHashPassword
	}

	user := models.User{
		Username: username,
		Password: string(hash),
	}

	existingUser, err := s.UserRepo.FindByUsername(username)

	if err != nil && !errors.Is(err, appErrors.ErrUserNotFound) {
		return appErrors.DatabaseError
	}

	if existingUser != nil {
		return appErrors.ErrUsernameAlreadyExists
	}

	if s.UserRepo.Create(&user) != nil {
		return appErrors.ErrFailedToCreateUser
	}

	return nil
}

func (s *AuthService) Refresh(refreshToken string) (string, string, *appErrors.AppError) {
	userId, err := s.TokenRepo.GetUserIDByRefreshToken(refreshToken)
	if err != nil {
		return "", "", appErrors.ErrRefreshTokenExpiredOrNotExists
	}

	s.TokenRepo.DeleteRefreshToken(refreshToken)

	user, err := s.UserRepo.FindByID(userId)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return "", "", appErrors.ErrUserNotFound
		}
		return "", "", appErrors.DatabaseError
	}

	token, err := generateJWT(user)
	if err != nil {
		return "", "", appErrors.ErrTokenGenerationFailed
	}

	newRefreshToken := uuid.NewString()
	storeErr := s.TokenRepo.StoreRefreshToken(newRefreshToken, user.ID)
	if storeErr != nil {
		return "", "", appErrors.ErrRefreshTokenStoreFailed
	}

	return token, newRefreshToken, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.TokenRepo.DeleteRefreshToken(refreshToken)
}

func generateJWT(user *models.User) (string, error) {
	claims := models.Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
