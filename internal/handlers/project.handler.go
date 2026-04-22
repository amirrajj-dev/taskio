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


type ProjectHandler struct {
	projectService services.ProjectService
}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{
		projectService: services.ProjectServicee,
	}
}

// CreateProject godoc
// @Summary      Create project
// @Description  Creates a new project within a team (owner or admin only)
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId    path      string                    true  "Team ID (UUID)"
// @Param        request   body      dtos.CreateProjectRequest true  "Project details"
// @Success      201  {object}  utils.SuccessResponse{data=models.Project}  "Project created successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context){
	currentUserID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}

	var req dtos.CreateProjectRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}

	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}

	teamID , parseErr := uuid.Parse(strings.TrimSpace(c.Param("teamId")))

	if parseErr != nil {
		c.JSON(http.StatusBadRequest , errors.NewBasicError("team id parse error", c.Request.URL.Path))
		return
	}

	project , err := h.projectService.CreateProject(c.Request.Context() , req.Name , req.Description , teamID , *currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError("project creation failed : " + err.Error() , c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusCreated , utils.NewSuccessResponse("project created succesfully" , project , c.Request.URL.Path))
}

// GetTeamProjects godoc
// @Summary      Get team projects
// @Description  Retrieves all projects for a specific team
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId   path      string  true  "Team ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse{data=[]models.ProjectResult}  "Projects fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid team ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Not a team member"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/projects [get]
func (h *ProjectHandler) GetTeamProjects(c *gin.Context){
	teamID , parseErr := uuid.Parse(strings.TrimSpace(c.Param("teamId")))

	if parseErr != nil {
		c.JSON(http.StatusBadRequest , errors.NewBasicError("team id parse error", c.Request.URL.Path))
		return
	}
	projects , err := h.projectService.GetTeamProjects(c.Request.Context() , teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError("failed to fetch projects : %w" + err.Error() , c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK , utils.NewSuccessResponse("projects fetched successfully" , projects , c.Request.URL.Path))	
}

// GetProject godoc
// @Summary      Get project by ID
// @Description  Retrieves detailed information about a specific project
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId      path      string  true  "Team ID (UUID)"
// @Param        projectId   path      string  true  "Project ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse{data=models.Project}  "Project fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Not a team member"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/projects/{projectId} [get]
func (h *ProjectHandler) GetProject(c *gin.Context){
	projectID , parseProjectIdErr := uuid.Parse(strings.TrimSpace(c.Param("projectId")))

	if parseProjectIdErr != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError("project id parse error" , c.Request.URL.Path))
		return
	}

	project , err := h.projectService.GetProject(c.Request.Context() , projectID)

	if err != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError(err.Error() , c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK , utils.NewSuccessResponse("project fetched successfully" , project , c.Request.URL.Path))
}

// UpdateProject godoc
// @Summary      Update project
// @Description  Updates project name and description
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId      path      string                    true  "Team ID (UUID)"
// @Param        projectId   path      string                    true  "Project ID (UUID)"
// @Param        request     body      dtos.UpdateProjectRequest true  "Updated project details"
// @Success      200  {object}  utils.SuccessResponse{data=models.Project}  "Project updated successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Not a team member"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/projects/{projectId} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	projectID , parseProjectIdErr := uuid.Parse(strings.TrimSpace(c.Param("projectId")))
	if parseProjectIdErr != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError("invalid project id" , c.Request.URL.Path))
		return
	}
	var req dtos.UpdateProjectRequest
	if !helpers.ShouldBindJSON(c , &req) {
		return
	}

	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}

	project , err := h.projectService.UpdateProject(c.Request.Context() , projectID, *userID , req.Name , req.Description);
	if err != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError(err.Error() , c.Request.URL.Path))
		return	
	}

	c.JSON(http.StatusOK , utils.NewSuccessResponse("project updated successfully" , project , c.Request.URL.Path))
}

// DeleteProject godoc
// @Summary      Delete project
// @Description  Permanently deletes a project and all related tasks (owner only)
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        teamId      path      string  true  "Team ID (UUID)"
// @Param        projectId   path      string  true  "Project ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse  "Project deleted successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Only owner can delete"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /teams/{teamId}/projects/{projectId} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	projectID , parseProjectIdErr := uuid.Parse(strings.TrimSpace(c.Param("projectId")))
	if parseProjectIdErr != nil {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError("invalid project id" , c.Request.URL.Path))
		return
	}
	if deleted , err := h.projectService.DeleteProject(c.Request.Context() , projectID , *userID);!deleted {
		c.JSON(http.StatusInternalServerError , errors.NewBasicError(err.Error() , c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK , utils.NewSuccessResponse("project deleted successfully" , nil , c.Request.URL.Path))
}