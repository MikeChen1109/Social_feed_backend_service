package models

import "time"

type FeedResponse struct {
	ID            uint      
	Title         string 
	Content       string   
	AuthorID      uint      
	AuthorName    string    
	Upvotes       int      
	CommentsCount int       
	CreatedAt     time.Time 
	UpdatedAt     time.Time 
}
