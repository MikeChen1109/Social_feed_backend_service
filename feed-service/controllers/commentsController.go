package controllers

import (
	"feed-service/common/helpers"
	"feed-service/models"
	"feed-service/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentsController struct {
	CommentsService *services.CommentService
}

// CreateComment godoc
// @Summary      Create a new comment
// @Description  Authenticated user creates a comment on a feed
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        body  body  models.CreateCommentRequest  true  "Comment content"
// @Success      200   {object}  models.CommentResponse
// @Failure      400   
// @Failure      401   
// @Failure      500   
// @Router       /comment [post]
func (s *CommentsController) CreateComment(c *gin.Context) {
	var req models.CreateCommentRequest

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

	response, apperror := s.CommentsService.CreateComment(req.FeedID, req.Content, claimsModel.UserID, claimsModel.Username)
	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create comment",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PaginatedComments godoc
// @Summary      Get paginated comments for a feed
// @Description  Retrieve comments for a specific feed with pagination
// @Tags         Comments
// @Produce      json
// @Param        id     query     int  true   "Feed ID"
// @Param        page   query     int  false  "Page number"
// @Param        limit  query     int  false  "Items per page"
// @Success      200    
// @Failure      400    
// @Failure      500    
// @Router       /comment/paginated [get]
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

	c.JSON(http.StatusOK, response)
}
