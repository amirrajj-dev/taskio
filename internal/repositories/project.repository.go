package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/amirrajj-dev/taskio/internal/dtos"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project models.Project) (*models.Project, error)
	FindProjectByName(ctx context.Context, name string) (*models.Project, error)
	FindProjectByID(ctx context.Context, projectID uuid.UUID) (*models.Project, error)
	FindTeamProjects(ctx context.Context, teamID uuid.UUID) ([]*models.ProjectResult, error)
	UpdateProjectByID(ctx context.Context, projectID uuid.UUID, data dtos.UpdateProjectRequest) (bool, error)
	DeleteProjectByID(ctx context.Context, projectID uuid.UUID) (bool, error)
}

type projectRepository struct {
	Projects *gorm.DB
}

var ProjectRepo *projectRepository

func NewProjectRepository() {
	ProjectRepo = &projectRepository{
		Projects: database.PG.Model(&models.Project{}),
	}
}

func (r *projectRepository) CreateProject(ctx context.Context, project models.Project) (*models.Project, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Projects.WithContext(opCtx).Create(&project)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to create project : %w", res.Error)
	}
	return &project, nil
}

func (r *projectRepository) FindProjectByName(ctx context.Context, name string) (*models.Project, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var project models.Project
	res := r.Projects.WithContext(opCtx).Where("name = ?", name).First(&project)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to find project : %w", res.Error)
	}
	return &project, nil
}

func (r *projectRepository) FindProjectByID(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var project models.Project
	res := r.Projects.WithContext(opCtx).Where("id = ?", projectID).First(&project)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to find project : %w", res.Error)
	}
	return &project, nil
}

func (r *projectRepository) FindTeamProjects(ctx context.Context, teamID uuid.UUID) ([]*models.ProjectResult, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var projects []*models.ProjectResult
	query := `
	SELECT projects.name , projects.description , projects.created_by , projects.id ,
	projects.created_at FROM projects WHERE team_id = ?;
	`

	res := r.Projects.WithContext(opCtx).Raw(query, teamID).Scan(&projects)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team has no projects")
		}
		return nil, fmt.Errorf("failed to find team projects : %w", res.Error)
	}
	return projects, nil
}

func (r *projectRepository) UpdateProjectByID(ctx context.Context, projectID uuid.UUID, data dtos.UpdateProjectRequest) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if data.Name != "" && data.Description != "" {
		res := r.Projects.WithContext(opCtx).Where("id = ?", projectID).Update("name", data.Name).Update("description", data.Description)
		if res.Error != nil {
			return false, fmt.Errorf("database error updating role: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return false, fmt.Errorf("organization user not found")
		}
	}

	if data.Name != "" && data.Description == "" {
		res := r.Projects.WithContext(opCtx).Where("id = ?", projectID).Update("name", data.Name)
		if res.Error != nil {
			return false, fmt.Errorf("database error updating role: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return false, fmt.Errorf("organization user not found")
		}
	}
	if data.Name == "" && data.Description != "" {
		res := r.Projects.WithContext(opCtx).Where("id = ?", projectID).Update("description", data.Description)
		if res.Error != nil {
			return false, fmt.Errorf("database error updating role: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return false, fmt.Errorf("organization user not found")
		}
	}
	return true, nil
}

func (r *projectRepository) DeleteProjectByID(ctx context.Context, projectID uuid.UUID) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Projects.WithContext(opCtx).Where("id = ?" , projectID).Delete(&models.Project{})

	if res.Error != nil {
		return false , fmt.Errorf("failed to delete project by id : %w" , res.Error)
	}

	if res.RowsAffected == 0 {
		return false , fmt.Errorf("project not found")
	}
	return true , nil
}
