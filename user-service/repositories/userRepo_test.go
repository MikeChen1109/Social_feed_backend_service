package repositories

import (
	"testing"
	apperrors "user-service/common/appErrors"
	"user-service/models"

	"github.com/stretchr/testify/assert"
)

func setupUserRepoForTest() (*UserRepository, func()) {
	db, dbCleanUp := setupTestDB()
	repo := &UserRepository{DB: db}

	return repo, dbCleanUp
}

func TestCreateUserSuccess(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanup := setupUserRepoForTest()
	defer dbCleanup()
	expectedUser := models.User{Username: "Test123", Password: "test321"}

	err := repo.Create(&expectedUser)
	assert.Nil(err)

	user, err := repo.FindByID(expectedUser.ID)
	assert.Nil(err)
	assert.Equal(user.ID, expectedUser.ID)
	assert.Equal(user.Username, expectedUser.Username)
}

func TestFindByUsernameSuccess(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanup := setupUserRepoForTest()
	defer dbCleanup()
	mockUser := models.User{Username: "Test123", Password: "test321"}

	err := repo.Create(&mockUser)
	assert.Nil(err)

	user, err := repo.FindByUsername(mockUser.Username)
	assert.Nil(err)
	assert.Equal(user.ID, mockUser.ID)
	assert.Equal(user.Username, mockUser.Username)
}

func TestFindByUsernameWhenUserNotExits(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanup := setupUserRepoForTest()
	defer dbCleanup()
	fakeUserName := "fakeName"

	user, err := repo.FindByUsername(fakeUserName)
	assert.Nil(user)
	assert.NotNil(err)
	assert.Equal(err, apperrors.ErrUserNotFound)
}

func TestFindByIdWhenUserNotExits(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanup := setupUserRepoForTest()
	defer dbCleanup()
	fakeUserId := uint(99)

	user, err := repo.FindByID(fakeUserId)
	assert.Nil(user)
	assert.NotNil(err)
	assert.Equal(err, apperrors.ErrUserNotFound)
}

