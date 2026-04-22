package routes

import (
	"github.com/amirrajj-dev/taskio/internal/handlers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/middlewares"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/gin-gonic/gin"
)

func RegisterTeamRoutes(r *gin.Engine , producer *queue.ActivityProducer) {
	repositories.NewTeamRepository()
	services.NewTeamService(producer)
	handler := handlers.NewTeamHandler()
	teamGroup := r.Group("/api/")
	teamGroup.Use(middlewares.AuthMiddleware())
	{
		teamGroup.POST("orgs/:id/teams", middlewares.CheckOrganizationMemberShip(), middlewares.CheckOrgMemberRole([]string{"owner", "admin"}), handler.CreateTeam)
		teamGroup.GET("orgs/:id/teams", middlewares.CheckOrganizationMemberShip(), handler.ListOrganizationTeams)
		teamGroup.PATCH("/teams/:teamId", middlewares.CheckTeamMemberShip(), middlewares.CheckTeamMemberRole([]string{"owner"}), handler.UpdateTeamName)
		teamGroup.POST("teams/:teamId/members", middlewares.CheckTeamMemberShip(), middlewares.CheckTeamMemberRole([]string{"owner" , "admin"}), handler.AddMemberToTeam)
		teamGroup.GET("teams/:teamId/members", middlewares.CheckTeamMemberShip() , handler.GetTeamMembers)
		teamGroup.DELETE("teams/:teamId/members/:userId", middlewares.CheckTeamMemberShip() ,  middlewares.CheckTeamMemberRole([]string{"owner" , "admin"}), handler.DeleteMemberFromTeam)
		teamGroup.PUT("teams/:teamId/members/:userId" , middlewares.CheckTeamMemberShip() , middlewares.CheckTeamMemberRole([]string{"owner" , "admin"}), handler.ChangeTeamMemberRole)
	}
}
