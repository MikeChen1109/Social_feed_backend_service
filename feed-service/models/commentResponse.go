package models

import "time"

type CommentResponse struct {
	ID            uint       
	Content       string    
	AuthorID      uint      
	AuthorName    string   
	FeedID        uint
	CreatedAt     time.Time 
	UpdatedAt     time.Time 
}