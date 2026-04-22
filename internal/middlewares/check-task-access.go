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

func CheckTaskAccess(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := utils.GetUserIDFromContext(c)
		taskId, parseTaskIdErr := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
		if parseTaskIdErr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.NewBasicError("invalid task id", c.Request.URL.Path))
			return
		}
		existingTask, findTaskErr := repositories.TaskRepo.FindTaskByID(c.Request.Context(), taskId)
		if findTaskErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError(findTaskErr.Error(), c.Request.URL.Path))
			return
		}
		if existingTask == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errors.NewBasicError("unexpected: task not found", c.Request.URL.Path))
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

		if c.Request.Method == http.MethodPut {
			c.Set("teamID" , teamMember.TeamID)
			if existingTask.CreatedBy == *userID {
				log.Println("task is assigned by this user")
				c.Next()
			}
			if len(roles) >= 1 {
				if !slices.Contains(roles, teamMember.Role) {
					var prefix string
					if len(roles) == 1 {
						prefix = "is"
					} else {
						prefix = "are"
					}
					errMsg := fmt.Sprintf("only %s %s allowed", strings.Join(roles, ", "), prefix)
					c.AbortWithStatusJSON(http.StatusForbidden, errors.NewBasicError(errMsg, c.Request.URL.Path))
					return
				}
				c.Next()
			}
			c.Next()
		}

		if existingTask.AssingedTo == *userID {
			log.Println("task is assigned to this user")
			c.Next()
		}
		if len(roles) >= 1 {
			if !slices.Contains(roles, teamMember.Role) {
				var prefix string
				if len(roles) == 1 {
					prefix = "is"
				} else {
					prefix = "are"
				}
				errMsg := fmt.Sprintf("only %s %s allowed", strings.Join(roles, ", "), prefix)
				c.AbortWithStatusJSON(http.StatusForbidden, errors.NewBasicError(errMsg, c.Request.URL.Path))
				return
			}
			c.Next()
		}
		c.Next()
	}
}
