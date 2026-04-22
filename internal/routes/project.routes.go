package routes

import (
	"github.com/amirrajj-dev/taskio/internal/handlers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/middlewares"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/gin-gonic/gin"
)


func RegisterProjectRoutes(r *gin.Engine , producer *queue.ActivityProducer){
	repositories.NewProjectRepository()
	services.NewProjectService(producer)
	handler := handlers.NewProjectHandler()
	
	projectGroup := r.Group("/api/")
	projectGroup.Use(middlewares.AuthMiddleware())
	{
		projectGroup.POST("teams/:teamId/projects" , middlewares.CheckTeamMemberShip() , middlewares.CheckTeamMemberRole([]string{"owner" , "admin"}) , handler.CreateProject)
		projectGroup.GET("teams/:teamId/projects" , middlewares.CheckTeamMemberShip() , handler.GetTeamProjects)
		projectGroup.GET("teams/:teamId/projects/:projectId" , middlewares.CheckTeamMemberShip() , handler.GetProject)
		projectGroup.PUT("teams/:teamId/projects/:projectId" , middlewares.CheckTeamMemberShip() , handler.UpdateProject)
		projectGroup.DELETE("teams/:teamId/projects/:projectId" , middlewares.CheckTeamMemberShip() , middlewares.CheckTeamMemberRole([]string{"owner"}) , handler.DeleteProject)
	}
}