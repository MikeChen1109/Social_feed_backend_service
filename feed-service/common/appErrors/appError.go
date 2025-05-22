package apperrors

import "net/http"

type AppError struct {
	StatusCode int
	Message    string
}

var (
	ErrUserNotFound              = New(http.StatusInternalServerError, "User not found")
	DatabaseError                = New(http.StatusInternalServerError, "Database error")
	ErrUserInvalidType           = New(http.StatusInternalServerError, "User type assertion failed")
	ErrFeedInvalidContentOrTitle = New(http.StatusBadRequest, "Feed content or title cannot be empty")
	ErrCommentIvalidContentOrFeedId = New(http.StatusBadRequest, "Comment content or feedId cannot be empty")
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
