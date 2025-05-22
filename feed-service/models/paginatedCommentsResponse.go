package models

import "gorm.io/gorm"

type PaginatedCommentsResponse struct {
	Data []Comment `json:"data"`
	Meta Meta      `json:"meta"`
}

type CommentRepository struct {
	DB *gorm.DB
}
