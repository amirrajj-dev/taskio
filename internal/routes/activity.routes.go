package routes

import (
	"github.com/amirrajj-dev/taskio/internal/handlers"
	"github.com/amirrajj-dev/taskio/internal/middlewares"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/gin-gonic/gin"
)

func RegisterActivityRoutes(r *gin.Engine){
	repositories.NewActivityRepository()
	services.NewActivityService()
	handler := handlers.NewActivityHandler()
	activityGroup := r.Group("/api/activity")
	activityGroup.Use(middlewares.AuthMiddleware())
	{
		activityGroup.GET("" , handler.GetActivities)
		activityGroup.DELETE("/:activityId" , handler.DeleteActivity)
	}
}