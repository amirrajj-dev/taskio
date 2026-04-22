package handlers

import (
	"net/http"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ActivityHandler struct {
	activityService services.ActivityService
}

func NewActivityHandler() *ActivityHandler {
	return  &ActivityHandler{
		activityService: services.ActivitiesService,
	}
}



// GetActivities godoc
// @Summary      Get user activities
// @Description  Retrieves all activity logs for the authenticated user
// @Tags         Activities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.SuccessResponseWithTotal{data=[]models.ActivityLog}  "Activities fetched successfully"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /activities [get]
func (h *ActivityHandler) GetActivities(c *gin.Context){
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	activities , err := h.activityService.GetUserActivities(c.Request.Context() , *userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError(err.Error() , c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK , utils.NewSuccessResponseWithTotal("activities fetched successfully" , activities , int64(len(activities)) , c.Request.URL.Path))
}

// DeleteActivity godoc
// @Summary      Delete an activity
// @Description  Deletes a specific activity log by ID
// @Tags         Activities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        activityId  path      string  true  "Activity ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse  "Activity deleted successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid activity ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /activities/{activityId} [delete]
func (h *ActivityHandler) DeleteActivity(c *gin.Context){
	activityID , parseErr := uuid.Parse(strings.TrimSpace(c.Param("activityId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid activity id", c.Request.URL.Path))
		return
	}
	deleted , err := h.activityService.DeleteActivity(c.Request.Context() , activityID);
	if err != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError(err.Error() , c.Request.URL.Path))
		return
	}
	if !deleted {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError("unexpected: failed to delete activity" , c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK , utils.NewSuccessResponse("activity deleted successfully" , nil , c.Request.URL.Path))
}