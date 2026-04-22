package routes

import (
	"github.com/amirrajj-dev/taskio/internal/handlers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/middlewares"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/gin-gonic/gin"
)

func RegisterOrgRoutes(r *gin.Engine, producer *queue.ActivityProducer) {
	repositories.NewOrganizationRepository()
	repositories.NewOrganizationInviteRepository()
	services.NewOrganizationService(producer)
	handler := handlers.NewOrganizationHandler()
	organizationGroup := r.Group("/api/orgs")
	organizationGroup.Use(middlewares.AuthMiddleware())
	{
		organizationGroup.POST("", handler.CreateOrg)
		organizationGroup.GET("", handler.GetOrganizations)
		organizationGroup.GET("/:id", middlewares.CheckOrganizationMemberShip(), handler.GetOrganization)
		organizationGroup.GET("/:id/members", middlewares.CheckOrganizationMemberShip(), middlewares.CheckOrganizationMemberShip(), handler.GetOrganizationMembers)
		organizationGroup.PATCH("/:id", middlewares.CheckOrganizationMemberShip(), middlewares.CheckOrgMemberRole([]string{"owner", "admin"}), handler.UpdateOrgName)
		organizationGroup.DELETE("/:id", middlewares.CheckOrganizationMemberShip(), middlewares.CheckOrgMemberRole([]string{"owner"}), handler.DeleteOrganization)
		organizationGroup.PUT(
			"/:id/users/:orgUserId/role",
			middlewares.CheckOrganizationMemberShip(),
			middlewares.CheckOrgMemberRole([]string{"owner", "admin"}),
			handler.UpdateOrgUserRole,
		)
		organizationGroup.POST("/:id/invite", middlewares.CheckOrganizationMemberShip(), middlewares.CheckOrgMemberRole([]string{"owner", "admin"}), handler.InviteToOrganization)
		organizationGroup.POST("/invites/accept", handler.AcceptInvitation)
		organizationGroup.POST("/invites/reject", handler.RejectInvitation)
		organizationGroup.GET("/invites/pending", handler.GetMyInvitations)
	}
}
