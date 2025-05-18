package apperrors

import "net/http"

type AppError struct {
	StatusCode int
	Message    string
}

var (
	ErrUserNotFound              = New(http.StatusInternalServerError, "User not found")
	DatabaseError                = New(http.StatusInternalServerError, "Database error")
	ErrInvalidUsernameOrPassword = New(http.StatusBadRequest, "Username and password are required")
	ErrInvalidPassword           = New(http.StatusUnauthorized, "Invalid password")
	ErrTokenGenerationFailed     = New(http.StatusInternalServerError, "Token generation failed")
	ErrFailedToHashPassword      = New(http.StatusInternalServerError, "Failed to hash password")
	ErrUsernameAlreadyExists     = New(http.StatusConflict, "Username already exists")
	ErrFailedToCreateUser        = New(http.StatusInternalServerError, "Failed to create user")
	ErrUserInvalidType           = New(http.StatusInternalServerError, "User type assertion failed")
	ErrUserNotInContext          = New(http.StatusUnauthorized, "User not found in context")
	ErrFeedNil                   = New(http.StatusBadRequest, "Feed cannot be nil")
	ErrFeedInvalidContentOrTitle = New(http.StatusBadRequest, "Feed content or title cannot be empty")
)

func New(status int, message string) *AppError {
	return &AppError{
		StatusCode: status,
		Message:    message,
	}
}

func (e *AppError) Error() string {
	return e.Message
}
