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

type TeamHandler struct {
	teamService services.TeamService
}

func NewTeamHandler() *TeamHandler {
	return &TeamHandler{
		teamService: services.TeamServicee,
	}
}

// CreateTeam godoc
// @Summary      Create team
// @Description  Creates a new team within an organization (owner or admin only)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string               true  "Organization ID (UUID)"
// @Param        request  body      dtos.CreateTeamRequest true  "Team details"
// @Success      201  {object}  utils.SuccessResponse{data=models.Team}  "Team created successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/{id}/teams [post]
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	currentUserID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	orgID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid org id", c.Request.URL.Path))
		return
	}
	var req dtos.CreateTeamRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}

	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	team, err := h.teamService.CreateTeam(c.Request.Context(), orgID, *currentUserID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusCreated, utils.NewSuccessResponse("team created succesfully", team, c.Request.URL.Path))
}

// ListOrganizationTeams godoc
// @Summary      List organization teams
// @Description  Retrieves all teams within an organization
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Organization ID (UUID)"
// @Success      200  {object}  utils.SuccessResponseWithTotal{data=[]models.TeamWithMemberCount}  "Teams fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid organization ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Not a member"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /orgs/{id}/teams [get]
func (h *TeamHandler) ListOrganizationTeams(c *gin.Context) {
	orgID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid org id", c.Request.URL.Path))
		return
	}
	teams, err := h.teamService.ListOrganizationTeams(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponseWithTotal("teams fetched succesfully", teams, int64(len(teams)), c.Request.URL.Path))
}

// UpdateTeamName godoc
// @Summary      Update team name
// @Description  Updates a team's name (owner only)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId    path      string               true  "Team ID (UUID)"
// @Param        request   body      dtos.UpdateTeamRequest true  "New team name"
// @Success      200  {object}  utils.SuccessResponse{data=models.Team}  "Team updated successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Only owner can update"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId} [patch]
func (h *TeamHandler) UpdateTeamName(c *gin.Context) {
	teamID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("teamId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid team id", c.Request.URL.Path))
		return
	}
	var req dtos.UpdateTeamRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}

	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	updateTeam, err := h.teamService.UpdateTeam(c.Request.Context(), teamID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("team updated succesfully", updateTeam, c.Request.URL.Path))
}

// AddMemberToTeam godoc
// @Summary      Add member to team
// @Description  Adds a user to a team (owner or admin only)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId    path      string                    true  "Team ID (UUID)"
// @Param        request   body      dtos.AddMemberToTeamRequest true  "User ID to add"
// @Success      200  {object}  utils.SuccessResponse{data=models.TeamMember}  "Member added successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/members [post]
func (h *TeamHandler) AddMemberToTeam(c *gin.Context) {
	actorID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	teamID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("teamId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid team id", c.Request.URL.Path))
		return
	}
	var req dtos.AddMemberToTeamRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}
	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	userID, parseErr := uuid.Parse(strings.TrimSpace(req.UserID))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid uuid", c.Request.URL.Path))
		return
	}
	teamMember, err := h.teamService.AddMemberToTeam(c.Request.Context(), teamID, userID , *actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("added to team succesfully", teamMember, c.Request.URL.Path))
}

// DeleteMemberFromTeam godoc
// @Summary      Remove member from team
// @Description  Removes a user from a team (owner or admin only)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId    path      string  true  "Team ID (UUID)"
// @Param        userId    path      string  true  "User ID (UUID) to remove"
// @Success      200  {object}  utils.SuccessResponse  "Member deleted successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/members/{userId} [delete]
func (h *TeamHandler) DeleteMemberFromTeam(c *gin.Context) {
	currentUserID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	teamID, parseTeamIDErr := uuid.Parse(strings.TrimSpace(c.Param("teamId")))
	if parseTeamIDErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid team id", c.Request.URL.Path))
		return
	}
	teamMemberID, parseTeamMemberIDErr := uuid.Parse(strings.TrimSpace(c.Param("userId")))
	if parseTeamMemberIDErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid team member id", c.Request.URL.Path))
		return
	}
	if deleted, err := h.teamService.DeleteMemberFromTeam(c.Request.Context(), teamID, *currentUserID, teamMemberID); !deleted {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("team member deleted successfully", nil, c.Request.URL.Path))
}

// GetTeamMembers godoc
// @Summary      Get team members
// @Description  Retrieves all members of a specific team
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId    path      string  true  "Team ID (UUID)"
// @Success      200  {object}  utils.SuccessResponseWithTotal{data=[]models.TeamMemberResult}  "Team members fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid team ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Not a team member"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/members [get]
func (h *TeamHandler) GetTeamMembers(c *gin.Context) {
	teamID, parseTeamIDErr := uuid.Parse(strings.TrimSpace(c.Param("teamId")))
	if parseTeamIDErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid team id", c.Request.URL.Path))
		return
	}
	teamMembers, err := h.teamService.GetTeamMembers(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponseWithTotal("team members fetched succesfully", teamMembers, int64(len(teamMembers)), c.Request.URL.Path))
}

// ChangeTeamMemberRole godoc
// @Summary      Change team member role
// @Description  Changes the role of a team member (owner or admin only)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId    path      string                       true  "Team ID (UUID)"
// @Param        userId    path      string                       true  "User ID (UUID)"
// @Param        request   body      dtos.ChangeTeamMemberRoleRequest true  "New role"
// @Success      200  {object}  utils.SuccessResponse  "Role changed successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/members/{userId} [put]
func (h *TeamHandler) ChangeTeamMemberRole(c *gin.Context) {
	actorID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}

	var req dtos.ChangeTeamMemberRoleRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}
	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}

	teamID, parseTeamIDErr := uuid.Parse(strings.TrimSpace(c.Param("teamId")))
	if parseTeamIDErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid team id", c.Request.URL.Path))
		return
	}

	targetUserID, parseTeamMemberIDErr := uuid.Parse(strings.TrimSpace(c.Param("userId")))
	if parseTeamMemberIDErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid team member id", c.Request.URL.Path))
		return
	}

	updated, err := h.teamService.ChangeTeamMemberRole(c.Request.Context(), req.Role, teamID, *actorID, targetUserID)
	if !updated {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError("failed to update role : "+err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("team member role changed successfully", nil, c.Request.URL.Path))
}
