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
// @Failure      403
// @Failure      404
// @Failure      500
// @Security     BearerAuth
// @Router       /feed/create [post]
func (s *FeedsController) CreateFeed(c *gin.Context) {
	var req models.CreateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

	response, apperror := s.FeedsService.CreateFeed(c.Request.Context(), req.Title, req.Content, claimsModel.UserID, claimsModel.Username)
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
// @Router       /feed/ [get]
func (s *FeedsController) GetFeeds(c *gin.Context) {
	feeds, err := s.FeedsService.GetFeeds(c.Request.Context())
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
// @Param        cursor query string false "Opaque nextCursor from the previous response"
// @Param        limit  query     int  false  "Items per page"
// @Success      200 {object} models.PaginatedFeedsResponse
// @Failure      400
// @Failure      500
// @Router       /feed/paginated [get]
func (s *FeedsController) PaginatedFeeds(c *gin.Context) {
	cursor, limit, err := models.ParsePagination(c.Request.URL.Query(), "feeds")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, apperror := s.FeedsService.PaginatedFeeds(c.Request.Context(), cursor, limit)

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

	feed, apperror := s.FeedsService.GetFeedByID(c.Request.Context(), uint(feedID))
	if apperror != nil {
		c.JSON(apperror.StatusCode, gin.H{
			"error": apperror.Message,
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
// @Failure      403
// @Failure      404
// @Failure      500
// @Security     BearerAuth
// @Router       /feed/{id} [put]
func (s *FeedsController) UpdateFeed(c *gin.Context) {
	claims, claimsErr := helpers.ParseClaims(c)
	if claimsErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var body struct {
		Title   string
		Content string
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	id := c.Param("id")
	feedID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feed ID",
		})
		return
	}

	apperror := s.FeedsService.UpdateFeed(c.Request.Context(), uint(feedID), claims.UserID, body.Title, body.Content)
	if apperror != nil {
		c.JSON(apperror.StatusCode, gin.H{
			"error": apperror.Message,
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
// @Failure      403
// @Failure      404
// @Failure      500
// @Security     BearerAuth
// @Router       /feed/{id} [delete]
func (s *FeedsController) DeleteFeed(c *gin.Context) {
	claims, claimsErr := helpers.ParseClaims(c)
	if claimsErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	id := c.Param("id")
	feedID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feed ID",
		})
		return
	}

	apperror := s.FeedsService.DeleteFeed(c.Request.Context(), uint(feedID), claims.UserID)
	if apperror != nil {
		c.JSON(apperror.StatusCode, gin.H{
			"error": apperror.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deleted feed successfully",
	})
}
