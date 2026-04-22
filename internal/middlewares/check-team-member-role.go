package middlewares

import (
	"net/http"
	"slices"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CheckTeamMemberRole(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := utils.GetUserIDFromContext(c)
		res, exists := c.Get("teamID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError("no team id provided", c.Request.URL.Path))
			return
		}
		teamID, ok := res.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError , errors.NewBasicError("invalid team id" , c.Request.URL.Path))
		}
		teamMember , findTeamMemberErr := repositories.TeamRepo.FindTeamMember(c.Request.Context(), teamID, *userID)
		if findTeamMemberErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError , errors.NewBasicError(findTeamMemberErr.Error() , c.Request.URL.Path))
			return 
		}
		if !slices.Contains(roles , teamMember.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden , errors.NewBasicError("not allowed for this operation" , c.Request.URL.Path))
			return 
		}
		c.Next()
	}
}
