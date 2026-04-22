package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CheckProjectMemberShipAndRole(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := utils.GetUserIDFromContext(c)
		projectID, parseIdErr := uuid.Parse(strings.TrimSpace(c.Param("projectId")))
		if parseIdErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError("invalid project id", c.Request.URL.Path))
			return
		}
		existingProject, findProjectErr := repositories.ProjectRepo.FindProjectByID(c.Request.Context(), projectID)
		if findProjectErr != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errors.NewBasicError("failed to find project "+findProjectErr.Error(), c.Request.URL.Path))
			return
		}

		if existingProject == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError("unexpected: project not found", c.Request.URL.Path))
			return
		}

		teamMember, findTeamMemberErr := repositories.TeamRepo.FindTeamMember(c.Request.Context(), existingProject.TeamID, *userID)
		if findTeamMemberErr != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errors.NewBasicError("team member not found", c.Request.URL.Path))
			return
		}
		if teamMember == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError("unexpected: team member not found", c.Request.URL.Path))
			return
		}

		taskId, parseTaskIdErr := uuid.Parse(strings.TrimSpace(c.Param("taskId")))

		// special for projects/:projectId/tasks/:taskId/subtasks
		if c.Request.Method == http.MethodPost && parseTaskIdErr == nil {
			existingTask, findTaskErr := repositories.TaskRepo.FindTaskByID(c.Request.Context(), taskId)
			if findTaskErr != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError(findTaskErr.Error(), c.Request.URL.Path))
				return
			}
			if existingTask.AssingedTo == *userID {
				log.Println("task is assigned to this user so allowed")
				c.Next()
			}
		}

		if len(roles) >= 1 {
			if !slices.Contains(roles, teamMember.Role) {
				errMsg := fmt.Sprintf("forbidden: only %s can do this", strings.Join(roles, ", "))
				c.AbortWithStatusJSON(http.StatusForbidden, errors.NewBasicError(errMsg, c.Request.URL.Path))
				return
			}
			c.Next()
		}
		c.Next()
	}
}
