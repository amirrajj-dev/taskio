package handlers

import (
	"net/http"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/dtos"
	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/helpers"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrganizationHandler struct {
	orgService services.OrganizationService
}

func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{
		orgService: services.OrgService,
	}
}

// CreateOrg godoc
// @Summary      Create organization
// @Description  Creates a new organization (user becomes owner)
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dtos.CreateOrganizationRequest  true  "Organization details"
// @Success      201  {object}  utils.SuccessResponse{data=models.Organization}  "Organization created successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs [post]
func (h *OrganizationHandler) CreateOrg(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	var req dtos.CreateOrganizationRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}

	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	organization, err := h.orgService.CreateOrganization(c.Request.Context(), req, *userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusCreated, utils.NewSuccessResponse("organization created succesfully", organization, c.Request.URL.Path))
}

// GetOrganizations godoc
// @Summary      Get user's organizations
// @Description  Retrieves all organizations where the user is a member
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.SuccessResponse{data=[]models.CurrentUserOrganizationsResponse}  "Organizations fetched successfully"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs [get]
func (h *OrganizationHandler) GetOrganizations(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	organizations, err := h.orgService.GetOrganizations(c.Request.Context(), *userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError("failed to fetch user involved organizations : "+err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("organizations fetched succesfully", organizations, c.Request.URL.Path))
}

// GetOrganization godoc
// @Summary      Get organization by ID
// @Description  Retrieves detailed information about a specific organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Organization ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse{data=models.OrganizationWithMembersCount}  "Organization fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Not a member"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("id is required", c.Request.URL.Path))
		return
	}
	orgID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid uuid", c.Request.URL.Path))
		return
	}
	organization, getOrgErr := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if getOrgErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError(getOrgErr.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("organization fetched successfully", organization, c.Request.URL.Path))
}

// GetOrganizationMembers godoc
// @Summary      Get organization members
// @Description  Retrieves all members of a specific organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Organization ID (UUID)"
// @Success      200  {object}  utils.SuccessResponseWithTotal{data=[]models.OrganizationUserResult}  "Members fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Not a member"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/{id}/members [get]
func (h *OrganizationHandler) GetOrganizationMembers(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("id is required", c.Request.URL.Path))
		return
	}
	orgID, parseErr := uuid.Parse(id)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid uuid", c.Request.URL.Path))
		return
	}
	orgUsers, getOrgUsersErr := h.orgService.GetOrganizationMembers(c.Request.Context(), orgID)
	if getOrgUsersErr != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError("failed to find organization memebers : %w"+getOrgUsersErr.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK , utils.NewSuccessResponseWithTotal("organization members fetched succesfully" , orgUsers , int64(len(orgUsers)) , c.Request.URL.Path))
}

// UpdateOrgName godoc
// @Summary      Update organization name
// @Description  Updates the name of an organization (owner or admin only)
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                         true  "Organization ID (UUID)"
// @Param        request  body      dtos.UpdateOrganizationNameRequest  true  "New organization name"
// @Success      200  {object}  utils.SuccessResponse{data=object}  "Organization updated successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/{id} [patch]
func (h *OrganizationHandler) UpdateOrgName(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("id is required", c.Request.URL.Path))
		return
	}
	orgID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid uuid", c.Request.URL.Path))
		return
	}
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	var req dtos.UpdateOrganizationNameRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}

	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	updated, err := h.orgService.UpdateOrganizationName(c.Request.Context(), orgID, *userID , req.Name)
	if !updated {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("organization updated successfully", map[string]string{
		"name":  req.Name,
		"orgID": orgID.String(),
	}, c.Request.URL.Path))
}

// DeleteOrganization godoc
// @Summary      Delete organization
// @Description  Permanently deletes an organization and all related data (owner only)
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Organization ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse  "Organization deleted successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Only owner can delete"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("id is required", c.Request.URL.Path))
		return
	}
	orgID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid uuid", c.Request.URL.Path))
		return
	}
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	deleted, err := h.orgService.DeleteOrganization(c.Request.Context(), orgID , *userID)
	if !deleted {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("organization deleted successfully", nil, c.Request.URL.Path))
}

// UpdateOrgUserRole godoc
// @Summary      Update member role
// @Description  Changes the role of an organization member (owner or admin only)
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string                       true  "Organization ID (UUID)"
// @Param        orgUserId    path      string                       true  "Organization User ID (UUID)"
// @Param        request      body      dtos.ChangeMemberRoleRequest true  "New role"
// @Success      200  {object}  utils.SuccessResponse{data=models.OrganizationUser}  "Role updated successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/{id}/users/{orgUserId}/role [put]
func (h *OrganizationHandler) UpdateOrgUserRole(c *gin.Context) {
	orgID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid org id", c.Request.URL.Path))
		return
	}

	targetOrgUserID, err := uuid.Parse(strings.TrimSpace(c.Param("orgUserId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid org user id", c.Request.URL.Path))
		return
	}

	currentUserID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unauthorized", c.Request.URL.Path))
		return
	}

	var req dtos.ChangeMemberRoleRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}

	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}

	updated, updateErr := h.orgService.UpdateOrganizationUserRole(
		c.Request.Context(),
		orgID,
		*currentUserID,
		targetOrgUserID,
		req.Role,
	)

	if updateErr != nil {
		switch {
		case strings.Contains(updateErr.Error(), "not found"):
			c.JSON(http.StatusNotFound, errors.NewBasicError(updateErr.Error(), c.Request.URL.Path))
		case strings.Contains(updateErr.Error(), "unauthorized"):
			c.JSON(http.StatusUnauthorized, errors.NewBasicError(updateErr.Error(), c.Request.URL.Path))
		case strings.Contains(updateErr.Error(), "forbidden"):
			c.JSON(http.StatusForbidden, errors.NewBasicError(updateErr.Error(), c.Request.URL.Path))
		case strings.Contains(updateErr.Error(), "invalid"):
			c.JSON(http.StatusBadRequest, errors.NewBasicError(updateErr.Error(), c.Request.URL.Path))
		default:
			c.JSON(http.StatusInternalServerError, errors.NewBasicError(updateErr.Error(), c.Request.URL.Path))
		}
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse("role updated successfully", updated, c.Request.URL.Path))
}

// InviteToOrganization godoc
// @Summary      Invite user to organization
// @Description  Sends an email invitation to join the organization (owner or admin only)
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                         true  "Organization ID (UUID)"
// @Param        request  body      dtos.InviteToOrganizationRequest  true  "Email to invite"
// @Success      200  {object}  utils.SuccessResponse  "Invitation sent successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/{id}/invite [post]
func (h *OrganizationHandler) InviteToOrganization(c *gin.Context) {
	currentUserID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	var req dtos.InviteToOrganizationRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}
	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	orgID, orgIdErr := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if orgIdErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid org id", c.Request.URL.Path))
		return
	}
	success, err := h.orgService.InviteToOrganization(c.Request.Context(), orgID, req.Email, *currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	if success {
		c.JSON(http.StatusOK, utils.NewSuccessResponse("Invitation sent successfully", nil, c.Request.URL.Path))
	} else {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError("Failed to send invitation", c.Request.URL.Path))
	}
}

// AcceptInvitation godoc
// @Summary      Accept organization invitation
// @Description  Accepts a pending invitation to join an organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dtos.AcceptInviteRequest  true  "Invitation token"
// @Success      200  {object}  utils.SuccessResponse{data=models.OrganizationUser}  "Invitation accepted successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid token"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/invites/accept [post]
func (h *OrganizationHandler) AcceptInvitation(c *gin.Context) {
    var req dtos.AcceptInviteRequest
    if !helpers.ShouldBindJSON(c, &req) {
        return
    }
    
    if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
        return
    }
    
    userID, exists := utils.GetUserIDFromContext(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, errors.NewBasicError("unauthorized", c.Request.URL.Path))
        return
    }
    
    orgUser, err := h.orgService.AcceptInvitation(c.Request.Context(), req.Token, *userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, errors.NewBasicError(err.Error(), c.Request.URL.Path))
        return
    }
    
    c.JSON(http.StatusOK, utils.NewSuccessResponse("invitation accepted successfully", orgUser, c.Request.URL.Path))
}

// RejectInvitation godoc
// @Summary      Reject organization invitation
// @Description  Rejects a pending invitation to join an organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dtos.AcceptInviteRequest  true  "Invitation token"
// @Success      200  {object}  utils.SuccessResponse  "Invitation rejected"
// @Failure      400  {object}  errors.BasicError  "Invalid token"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/invites/reject [post]
func (h *OrganizationHandler) RejectInvitation(c *gin.Context) {
    var req dtos.AcceptInviteRequest
    if !helpers.ShouldBindJSON(c, &req) {
        return
    }
    
    userID, exists := utils.GetUserIDFromContext(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, errors.NewBasicError("unauthorized", c.Request.URL.Path))
        return
    }
    
    err := h.orgService.RejectInvitation(c.Request.Context(), req.Token, *userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, errors.NewBasicError(err.Error(), c.Request.URL.Path))
        return
    }
    
    c.JSON(http.StatusOK, utils.NewSuccessResponse("invitation rejected", nil, c.Request.URL.Path))
}

// GetMyInvitations godoc
// @Summary      Get pending invitations
// @Description  Retrieves all pending organization invitations for the authenticated user
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.SuccessResponse{data=[]models.OrganizationInvite}  "Invitations fetched successfully"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/invites/pending [get]
func (h *OrganizationHandler) GetMyInvitations(c *gin.Context) {
    userID, exists := utils.GetUserIDFromContext(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, errors.NewBasicError("unauthorized", c.Request.URL.Path))
        return
    }
    
    invitations, err := h.orgService.GetPendingInvitations(c.Request.Context(), *userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
        return
    }
    
    c.JSON(http.StatusOK, utils.NewSuccessResponse("invitations fetched successfully", invitations, c.Request.URL.Path))
}