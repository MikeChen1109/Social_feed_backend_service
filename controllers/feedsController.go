package controllers

import (
	"myApp/SocialFeed/common/helpers"
	"myApp/SocialFeed/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FeedsController struct {
	FeedsService *services.FeedService
}

func (s *FeedsController) CreateFeed(c *gin.Context) {
	var body struct {
		Title   string
		Content string
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	userModel, err := helpers.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: " + err.Message,
		})
		return
	}

	feed, apperror := s.FeedsService.CreateFeed(body.Title, body.Content, userModel)
	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create feed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feed created successfully",
		"feed":    feed,
	})
}

func (s *FeedsController) GetFeeds(c *gin.Context) {
	feeds, err := s.FeedsService.GetFeeds()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve feeds: " + err.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feeds": feeds,
	})
}

func (s *FeedsController) GetFeedByID(c *gin.Context) {
	id := c.Param("id")
	feedID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feed ID",
		})
		return
	}

	feed, apperror := s.FeedsService.GetFeedByID(uint(feedID))
	if apperror != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Feed not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feed": feed,
	})
}

func (s *FeedsController) UpdateFeed(c *gin.Context) {
	var body struct {
		Title   string
		Content string
	}

	c.Bind(&body)

	id := c.Param("id")
	feedID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feed ID",
		})
		return
	}

	apperror := s.FeedsService.UpdateFeed(uint(feedID), body.Title, body.Content)
	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update feed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Updated feed successfully",
	})
}

func (s *FeedsController) DeleteFeed(c *gin.Context) {
	id := c.Param("id")
	feedID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feed ID",
		})
		return
	}

	apperror := s.FeedsService.DeleteFeed(uint(feedID))
	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete feed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deleted feed successfully",
	})
}
