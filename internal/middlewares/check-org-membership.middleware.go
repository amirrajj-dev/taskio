package middlewares

import (
	"net/http"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CheckOrganizationMemberShip() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := utils.GetUserIDFromContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
			return
		}
		id := strings.TrimSpace(c.Param("id"))
		if id == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.NewBasicError("id is required", c.Request.URL.Path))
			return
		}
		orgID, parseUuidErr := uuid.Parse(id)
		if parseUuidErr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.NewBasicError("invalid uuid", c.Request.URL.Path))
			return
		}
		isMember , isMemberErr := services.OrgService.CheckMemberShip(c.Request.Context() , orgID , *userID)
		if !isMember {
			c.AbortWithStatusJSON(http.StatusForbidden , errors.NewBasicError(isMemberErr.Error() , c.Request.URL.Path))
		}
		c.Next()
	}
}
