package controllers

import (
	"feed-service/common/helpers"
	"feed-service/models"
	"feed-service/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FeedsController struct {
	FeedsService *services.FeedService
}

// CreateFeed godoc
// @Summary      Create a new feed
// @Description  Authenticated user creates a new feed
// @Tags         Feeds
// @Accept       json
// @Produce      json
// @Param        body  body  models.CreateFeedRequest  true  "Feed content"
// @Success      200   {object}  models.FeedResponse
// @Failure      400   
// @Failure      401   
// @Failure      500   
// @Router       /feed [post]
func (s *FeedsController) CreateFeed(c *gin.Context) {
	var req models.CreateFeedRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	claimsModel, err := helpers.ParseClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: " + err.Message,
		})
		return
	}

	response, apperror := s.FeedsService.CreateFeed(req.Title, req.Content, claimsModel.UserID, claimsModel.Username)
	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create feed",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetFeeds godoc
// @Summary      Get all feeds
// @Description  Retrieve all feeds without pagination
// @Tags         Feeds
// @Produce      json
// @Success      200   {object}  []models.FeedResponse
// @Failure      500   
// @Router       /feed/all [get]
func (s *FeedsController) GetFeeds(c *gin.Context) {
	feeds, err := s.FeedsService.GetFeeds()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve feeds: " + err.Message,
		})
		return
	}

	c.JSON(http.StatusOK, feeds)
}

// PaginatedFeeds godoc
// @Summary      Get paginated feeds
// @Description  Retrieve feeds with pagination parameters
// @Tags         Feeds
// @Produce      json
// @Param        page   query     int  false  "Page number"
// @Param        limit  query     int  false  "Items per page"
// @Success      200    
// @Failure      500    
// @Router       /feed/paginated [get]
func (s *FeedsController) PaginatedFeeds(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1)
	response, apperror := s.FeedsService.PaginatedFeeds(offset, limit)

	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve feeds: " + apperror.Message,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetFeedByID godoc
// @Summary      Get feed by ID
// @Description  Retrieve a single feed by its ID
// @Tags         Feeds
// @Produce      json
// @Param        id   path      int  true  "Feed ID"
// @Success      200  {object}  models.FeedResponse
// @Failure      400  
// @Failure      404  
// @Router       /feed/{id} [get]
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

	c.JSON(http.StatusOK, feed)
}

// UpdateFeed godoc
// @Summary      Update a feed
// @Description  Update feed content by ID
// @Tags         Feeds
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "Feed ID"
// @Param        body  body      models.UpdateFeedRequest  true  "Updated content"
// @Success      200   
// @Failure      400   
// @Failure      500   
// @Router       /feed/{id} [put]
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

// DeleteFeed godoc
// @Summary      Delete a feed
// @Description  Delete a feed by ID
// @Tags         Feeds
// @Produce      json
// @Param        id   path      int  true  "Feed ID"
// @Success      200  
// @Failure      400  
// @Failure      500  
// @Router       /feed/{id} [delete]
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
