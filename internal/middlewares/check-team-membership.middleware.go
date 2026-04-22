package middlewares

import (
	"net/http"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/helpers"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Req struct {
	TeamID string `json:"team_id"`
}

func CheckTeamMemberShip() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := utils.GetUserIDFromContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewBasicError("unauthorized", c.Request.URL.Path))
			return
		}

		var teamID uuid.UUID
		var parseErr error
		paramTeamIDStr := strings.TrimSpace(c.Param("teamId"))
		if paramTeamIDStr != "" {
			teamID, parseErr = uuid.Parse(paramTeamIDStr)
			if parseErr != nil {
				// Log the specific error for debugging
				// logger.Debugf("Failed to parse teamId from param '%s': %v", paramTeamIDStr, parseErr)
				c.AbortWithStatusJSON(http.StatusBadRequest, errors.NewBasicError("invalid team ID format", c.Request.URL.Path))
				return
			}
		} else {
			// 2. If not in param, try to get from request body
			var req Req
			if !helpers.ShouldBindJSON(c, &req) {
				return
			}
			if req.TeamID == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, errors.NewBasicError("team ID is required in request body", c.Request.URL.Path))
				return
			}

			teamID, parseErr = uuid.Parse(req.TeamID)
			if parseErr != nil {
				// Log the specific error for debugging
				// logger.Debugf("Failed to parse teamId from request body '%s': %v", req.TeamID, parseErr)
				c.AbortWithStatusJSON(http.StatusBadRequest, errors.NewBasicError("invalid team ID format in request body", c.Request.URL.Path))
				return
			}
		}

		if teamID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.NewBasicError("team ID could not be determined", c.Request.URL.Path))
			return
		}

		team , _ := repositories.TeamRepo.FindTeamByID(c.Request.Context() , teamID)

		if team == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError , errors.NewBasicError("team not found" , c.Request.URL.Path))
			return 
		}

		isMember, err := repositories.TeamRepo.CheckTeamMemberShip(c.Request.Context(), teamID, *userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.NewBasicError("internal server error", c.Request.URL.Path))
			return
		}

		if !isMember {
			c.AbortWithStatusJSON(http.StatusForbidden, errors.NewBasicError("you are not a member of this team", c.Request.URL.Path)) // Corrected grammar
			return
		}

		c.Set("teamID", teamID)
		c.Next()
	}
}
