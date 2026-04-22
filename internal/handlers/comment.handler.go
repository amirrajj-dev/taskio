package handlers

import (
	"net/http"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/dtos"
	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/helpers"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CommentHandler struct {
	commentService services.CommentService
}

func NewCommentHandler() *CommentHandler {
	return &CommentHandler{
		commentService: services.CommentsService,
	}
}

// CreateComment godoc
// @Summary      Create a comment
// @Description  Adds a new comment to a specific task
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskId   path      string                    true  "Task ID (UUID)"
// @Param        request  body      dtos.CreateCommentRequest true  "Comment content"
// @Success      201  {object}  utils.SuccessResponse{data=models.Comment}  "Comment created successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /tasks/{taskId}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	var req dtos.CreateCommentRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}

	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	taskID, _ := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
	createdComment, createErr := h.commentService.CreateComment(c.Request.Context(), taskID, *userID, req.Content)
	if createErr != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(createErr.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusCreated, utils.NewSuccessResponse("comment created successfully", createdComment, c.Request.URL.Path))
}

// GetComments godoc
// @Summary      Get task comments
// @Description  Retrieves all comments for a specific task
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskId   path      string  true  "Task ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse{data=[]models.Comment}  "Comments fetched successfully"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /tasks/{taskId}/comments [get]
func (h *CommentHandler) GetComments(c *gin.Context){
	taskID, _ := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
	comments , err := h.commentService.GetTaskComments(c.Request.Context() , taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError(err.Error() , c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK , utils.NewSuccessResponse("comments fetched successfully" , comments , c.Request.URL.Path))
}