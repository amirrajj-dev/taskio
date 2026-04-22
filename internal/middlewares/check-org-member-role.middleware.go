package middlewares

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CheckOrgMemberRole(roles []string) gin.HandlerFunc {
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
		orgUser , getOrgMemberErr := services.OrgService.GetOrganizationMember(c.Request.Context() , orgID , *userID)
		if getOrgMemberErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError , errors.NewBasicError("something goes wrong : " + getOrgMemberErr.Error() , c.Request.URL.Path))
			return 
		}
		if slices.Contains(roles , orgUser.Role) {
			c.Next()
		}else {
			var pre string
			if len(roles) ==  1 {
				pre = "is"
			}else {
				pre = "are"
			}
			msg := fmt.Sprintf("only %s %s allowed" , strings.Join(roles, ", ") , pre)
			c.AbortWithStatusJSON(http.StatusForbidden , errors.NewBasicError(msg , c.Request.URL.Path))
			return 
		}
		c.Next()
	}
}