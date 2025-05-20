package services

import (
	"errors"
	appErrors "user-service/common/appErrors"
	"user-service/models"
	"user-service/repositories"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo repositories.UserRepositoryInterface
}

func (s *AuthService) Login(username, password string) (string, *appErrors.AppError) {
	if username == "" || password == "" {
		return "", appErrors.ErrInvalidUsernameOrPassword
	}

	user, err := s.UserRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return "", appErrors.ErrUserNotFound
		}
		return "", appErrors.DatabaseError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", appErrors.ErrInvalidPassword
	}

	token, err := generateJWT(user)
	if err != nil {
		return "", appErrors.ErrTokenGenerationFailed
	}

	return token, nil
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

func (s *AuthService) VerifyToken(tokenStr string) (*models.User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["sub"] == nil {
		return nil, errors.New("invalid claims")
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		return nil, errors.New("token expired")
	}

	user, err := s.UserRepo.FindByID(claims["sub"].(float64))
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, appErrors.DatabaseError
	}

	return user, nil
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
