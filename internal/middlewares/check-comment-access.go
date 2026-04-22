package middlewares

import (
	"net/http"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CheckCommentAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := utils.GetUserIDFromContext(c)
		taskID, parseTaskIdErr := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
		if parseTaskIdErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError("invalid task id", c.Request.URL.Path))
			return
		}
		existingTask, findTaskErr := repositories.TaskRepo.FindTaskByID(c.Request.Context(), taskID)
		if findTaskErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError(findTaskErr.Error(), c.Request.URL.Path))
			return
		}
		if existingTask == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError("unexpected: task not found" , c.Request.URL.Path))
			return
		}
		existingProject, findProjectErr := repositories.ProjectRepo.FindProjectByID(c.Request.Context(), existingTask.ProjectID)
		if findProjectErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError(findProjectErr.Error(), c.Request.URL.Path))
			return
		}
		if existingProject == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errors.NewBasicError("project not found", c.Request.URL.Path))
			return
		}
		teamMember, findTeamMemberErr := repositories.TeamRepo.FindTeamMember(c.Request.Context(), existingProject.TeamID, *userID)
		if findTeamMemberErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError(findTeamMemberErr.Error(), c.Request.URL.Path))
			return
		}
		if teamMember == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errors.NewBasicError("unexpected: team member not found", c.Request.URL.Path))
			return
		}
		c.Next()
	}
}
