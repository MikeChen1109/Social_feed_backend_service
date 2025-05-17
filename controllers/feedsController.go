package controllers

import (
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/models"

	"github.com/gin-gonic/gin"
)

func CreateFeed(c *gin.Context) {
	var body struct {
		Title string
		Content string
	}

	c.Bind(&body)

	if body.Title == "" || body.Content == "" {
		c.JSON(400, gin.H{
			"error": "Title and content are required",
		})
		return
	}
	
	feed := models.Feed{AuthorName: "John Doe", AuthorID: 001, Title: body.Title, Content: body.Content}
	result := initializers.DB.Create(&feed)

	if result.Error != nil {
		c.JSON(500, gin.H{
			"error": "Failed to create feed",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Feed created successfully",
		"feed": feed,
	})
}

func GetFeeds(c *gin.Context) {
	var feeds []models.Feed
	result := initializers.DB.Find(&feeds)

	if result.Error != nil {
		c.JSON(500, gin.H{
			"error": "Failed to retrieve feeds",
		})
		return
	}

	c.JSON(200, gin.H{
		"feeds": feeds,
	})
}

func GetFeedByID(c *gin.Context) {
	var feed models.Feed
	id := c.Param("id")

	result := initializers.DB.First(&feed, id)

	if result.Error != nil {
		c.JSON(404, gin.H{
			"error": "Feed not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"feed": feed,
	})
}

func UpdateFeed(c *gin.Context) {
	var body struct {
		Title   string
		Content string
	}

	c.Bind(&body)

	id := c.Param("id")
	var feed models.Feed

	result := initializers.DB.First(&feed, id)

	if result.Error != nil {
		c.JSON(404, gin.H{
			"error": "Feed not found",
		})
		return
	}

	if body.Title != "" {
		feed.Title = body.Title
	}
	if body.Content != "" {
		feed.Content = body.Content
	}

	initializers.DB.Save(&feed)

	c.JSON(200, gin.H{
		"message": "Feed updated successfully",
		"feed": feed,
	})
}

func DeleteFeed(c *gin.Context) {
	id := c.Param("id")
	var feed models.Feed

	result := initializers.DB.First(&feed, id)

	if result.Error != nil {
		c.JSON(404, gin.H{
			"error": "Feed not found",
		})
		return
	}

	initializers.DB.Delete(&feed)

	c.JSON(200, gin.H{
		"message": "Feed deleted successfully",
	})
}