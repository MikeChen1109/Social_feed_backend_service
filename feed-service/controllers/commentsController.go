package controllers

import (
	"feed-service/common/helpers"
	"feed-service/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentsController struct {
	CommentsService *services.CommentService
}

func (s *CommentsController) CreateComment(c *gin.Context) {
	var body struct {
		FeedID  uint
		Content string
	}

	if err := c.Bind(&body); err != nil {
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

	comment, apperror := s.CommentsService.CreateComment(body.FeedID, body.Content, claimsModel.UserID, claimsModel.Username)
	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create comment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment created successfully",
		"comment": comment,
	})
}

func (s *CommentsController) PaginatedComments(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	feedIdStr := c.DefaultQuery("id", "0")

	feedID, err := strconv.Atoi(feedIdStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1)
	response, apperror := s.CommentsService.PaginatedComments(offset, limit, uint(feedID))

	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve comments: " + apperror.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}
