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
// @Security     BearerAuth
// @Router       /comment/create [post]
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

	response, apperror := s.CommentsService.CreateComment(c.Request.Context(), req.FeedID, req.Content, claimsModel.UserID, claimsModel.Username)
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
// @Param        cursor query string false "Opaque nextCursor from the previous response"
// @Param        limit  query     int  false  "Items per page"
// @Success      200 {object} models.PaginatedCommentsResponse
// @Failure      400
// @Failure      500
// @Router       /comment/paginated [get]
func (s *CommentsController) PaginatedComments(c *gin.Context) {
	feedID, err := strconv.ParseUint(c.Query("id"), 10, 63)
	if err != nil || feedID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}
	cursor, limit, err := models.ParsePagination(c.Request.URL.Query(), models.CommentCursorScope(uint(feedID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, apperror := s.CommentsService.PaginatedComments(c.Request.Context(), cursor, limit, uint(feedID))

	if apperror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve comments: " + apperror.Message,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
