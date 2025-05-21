package services

import (
	"context"
	"os"
	"testing"
	"time"
	appErrors "user-service/common/appErrors"
	"user-service/models"
	"user-service/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

/* ---------- Mocks ---------- */

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) Create(u *models.User) error {
	args := m.Called(u)
	return args.Error(0)
}
func (m *mockUserRepo) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

type mockTokenRepo struct{ mock.Mock }

func (m *mockTokenRepo) StoreRefreshToken(token string, uid uint, exp time.Duration) error {
	args := m.Called(token, uid, exp)
	return args.Error(0)
}
func (m *mockTokenRepo) GetUserIDByRefreshToken(token string) (uint, error) {
	args := m.Called(token)
	return args.Get(0).(uint), args.Error(1)
}
func (m *mockTokenRepo) DeleteRefreshToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}
func (m *mockTokenRepo) WithContext(_ context.Context) repositories.TokenRepositoryInterface {
	return m
}

/* ---------- helper ---------- */

func mustHash(pw string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(hash)
}

func init() {
	_ = os.Setenv("JWT_SECRET", "test-secret")
}

/* ---------- Tests ---------- */

func TestLoginSuccess(t *testing.T) {
	assert := assert.New(t)

	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	svc := AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}

	u := &models.User{Username: "mike", Password: mustHash("password")}
	userId := uint(1)
	u.ID = userId

	userRepo.On("FindByUsername", "mike").Return(u, nil)
	tokenRepo.On("StoreRefreshToken", mock.AnythingOfType("string"), userId, mock.Anything).Return(nil)

	access, refresh, err := svc.Login("mike", "password")

	assert.Empty(err)
	assert.NotEmpty(access)
	assert.NotEmpty(refresh)

	parsed, _ := jwt.ParseWithClaims(access, &models.Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	claims := parsed.Claims.(*models.Claims)
	assert.Equal(uint(1), claims.UserID)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestLoginWhenWrongPassword(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	svc := &AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}

	u := &models.User{Username: "mike", Password: mustHash("password")}
	userRepo.On("FindByUsername", "mike").Return(u, nil)

	_, _, appErr := svc.Login("mike", "wrong")
	assert.Equal(t, appErrors.ErrInvalidPassword, appErr)
	userRepo.AssertExpectations(t)
}

func TestLoginWhenInvalidUsernameOrPassword(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	svc := &AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}

	_, _, appErr := svc.Login("", "")
	assert.Equal(t, appErrors.ErrInvalidUsernameOrPassword, appErr)
}


func TestLoginWhenUserNotExist(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	svc := &AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}

	userRepo.On("FindByUsername", "mike").Return((*models.User)(nil), appErrors.ErrUserNotFound)

	_, _, appErr := svc.Login("mike", "wrong")
	assert.Equal(t, appErrors.ErrUserNotFound, appErr)
	userRepo.AssertExpectations(t)
}

func TestSignupSuccess(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	svc := &AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}

	userRepo.On("FindByUsername", "new").Return((*models.User)(nil), appErrors.ErrUserNotFound)
	userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	err := svc.Signup("new", "pw")
	assert.Nil(t, err)
	userRepo.AssertExpectations(t)
}

func TestSignupWhenInvalidUsernameOrPassword(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	svc := &AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}

	err := svc.Signup("", "")
	assert.NotNil(t, err)
	assert.Equal(t, appErrors.ErrInvalidUsernameOrPassword, err)
}

func TestLogout(t *testing.T) {
	tokenRepo := new(mockTokenRepo)
	tokenRepo.On("DeleteRefreshToken", "r1").Return(nil)

	svc := &AuthService{TokenRepo: tokenRepo}
	err := svc.Logout("r1")

	assert.Nil(t, err)
	tokenRepo.AssertCalled(t, "DeleteRefreshToken", "r1")
	tokenRepo.AssertExpectations(t)
}

func TestRefreshSuccess(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	svc := &AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}

	oldRefresh := "old-refresh"
	newRefreshMatcher := mock.AnythingOfType("string")

	u := &models.User{Username: "dev", Password: mustHash("x")}
	userId := uint(2)
	u.ID = userId

	tokenRepo.On("GetUserIDByRefreshToken", oldRefresh).Return(userId, nil)
	tokenRepo.On("DeleteRefreshToken", oldRefresh).Return(nil)
	userRepo.On("FindByID", userId).Return(u, nil)
	tokenRepo.On("StoreRefreshToken", newRefreshMatcher, userId, mock.Anything).Return(nil)

	access, newRefresh, err := svc.Refresh(oldRefresh)

	assert.Nil(t, err)
	assert.NotEmpty(t, access)
	assert.NotEqual(t, oldRefresh, newRefresh)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestRefreshWhenUserNotExist(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	svc := &AuthService{UserRepo: userRepo, TokenRepo: tokenRepo}
	
	oldRefresh := "old-refresh"
	userId := uint(2)

	tokenRepo.On("GetUserIDByRefreshToken", oldRefresh).Return(userId, nil)
	tokenRepo.On("DeleteRefreshToken", oldRefresh).Return(nil)
	userRepo.On("FindByID", userId).Return((*models.User)(nil), appErrors.ErrUserNotFound)

	access, newRefresh, err := svc.Refresh(oldRefresh)

	assert.NotNil(t, err)
	assert.Empty(t, access)
	assert.Empty(t, newRefresh)
	assert.Equal(t, err, appErrors.ErrUserNotFound)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}