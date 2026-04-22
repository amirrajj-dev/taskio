package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/consumers"
	"github.com/amirrajj-dev/taskio/internal/dtos"
	"github.com/amirrajj-dev/taskio/internal/events"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/websocket"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ProjectService interface {
	CreateProject(ctx context.Context, name, description string, teamID uuid.UUID, userID uuid.UUID) (*models.Project, error)
	GetTeamProjects(ctx context.Context, teamID uuid.UUID) ([]*models.ProjectResult, error)
	GetProject(ctx context.Context, projectID uuid.UUID) (*models.Project, error)
	UpdateProject(ctx context.Context, projectID , userID uuid.UUID, name, description string) (*models.Project, error)
	DeleteProject(ctx context.Context, projectID , userID uuid.UUID) (bool, error)
}

type projectService struct {
	projectRepo      repositories.ProjectRepository
	teamRepo         repositories.TeamRepository
	userRepo         repositories.UserRepository
	activityProducer *queue.ActivityProducer
	queueConn     *amqp.Connection
}

var ProjectServicee *projectService

func NewProjectService(producer *queue.ActivityProducer) {
	ProjectServicee = &projectService{
		projectRepo: repositories.ProjectRepo,
		teamRepo:    repositories.TeamRepo,
		userRepo: repositories.UserRepo,
		activityProducer: producer,
		queueConn:     queue.QueueConn,
	}
}

func (s *projectService) CreateProject(ctx context.Context, name, description string, teamID uuid.UUID, userID uuid.UUID) (*models.Project, error) {
	existingTeam, findTeamErr := s.teamRepo.FindTeamByID(ctx, teamID)
	if findTeamErr != nil {
		return nil, findTeamErr
	}
	slog.Info("creating project", "name", name, "team_id", teamID, "user_id", userID)
	if isMember, _ := s.teamRepo.CheckTeamMemberShip(ctx, teamID, userID); !isMember {
		return nil, fmt.Errorf("not a part of this team")
	}
	existingProject, _ := s.projectRepo.FindProjectByName(ctx, strings.TrimSpace(name))
	if existingProject != nil {
		return nil, fmt.Errorf("project with the given name already exists")
	}
	project := models.Project{
		ID:          uuid.New(),
		OrgID:       existingTeam.OrgID,
		TeamID:      existingTeam.ID,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		CreatedBy:   userID,
		CreatedAt:   time.Now().UTC(),
	}
	newProject, createProjectErr := s.projectRepo.CreateProject(ctx, project)
	if createProjectErr != nil {
		slog.Info("failed to create project" , "error" , createProjectErr.Error())
		return nil, fmt.Errorf("failed to create project : %w", createProjectErr)
	}

	slog.Info("project created", "project_id", newProject.ID, "name", name)

	if websocket.WsManager != nil {
		payload := websocket.ProjectCreatedPayload{
			ProjectName: newProject.Name,
			ProjectID: newProject.ID,
			CreatedBy: newProject.CreatedBy,
		}
		websocket.WsManager.BroadcastToRoom("team:"+teamID.String() , websocket.EventProjectCreated , payload)
	}else {
		slog.Info("failed to send event via websocket for create project" , "project_name" , name)
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.CreateProject,
		UserID: &userID,
		OrgID: &newProject.OrgID,
		ProjectID: &newProject.ID,
		Metadata: map[string]any{
			"project_name" : newProject.Name,
		},
	}
	if err := s.activityProducer.Publish(event);err != nil {
		slog.Warn("failed to publish create project event" , "error" , err.Error())
	}
	return newProject, nil
}

func (s *projectService) GetTeamProjects(ctx context.Context, teamID uuid.UUID) ([]*models.ProjectResult, error) {
	slog.Info("fetching team projects", "team_id", teamID)
	projects, err := s.projectRepo.FindTeamProjects(ctx, teamID)
	if err != nil {
		return nil, err
	}
	slog.Info("team projects fetched", "team_id", teamID, "count", len(projects))
	return projects, nil
}

func (s *projectService) GetProject(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	slog.Info("fetching project", "project_id", projectID)
	project, err := s.projectRepo.FindProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project : %w", err)
	}
	slog.Error("failed to fetch project", "project_id", projectID, "error", err)
	return project, err
}

func (s *projectService) UpdateProject(ctx context.Context, projectID , userID uuid.UUID, name, description string) (*models.Project, error) {
	slog.Info("updating project", "project_id", projectID, "user_id", userID)
	existingProject, findProjectErr := s.projectRepo.FindProjectByID(ctx, projectID)
	if findProjectErr != nil {
		return nil, findProjectErr
	}
	updatedProject := models.Project{
		ID:        existingProject.ID,
		OrgID:     existingProject.OrgID,
		TeamID:    existingProject.TeamID,
		CreatedBy: existingProject.CreatedBy,
		CreatedAt: existingProject.CreatedAt,
	}
	var updates dtos.UpdateProjectRequest
	if strings.TrimSpace(name) != "" {
		updates.Name = strings.TrimSpace(name)
	}
	if strings.TrimSpace(description) != "" {
		updates.Description = strings.TrimSpace(description)
	}

	if updates.Name != "" && len(updates.Name) < 3 {
		return nil, fmt.Errorf("name must be atleast 3 characters long")
	}

	if updates.Description != "" && len(updates.Description) < 5 {
		return nil, fmt.Errorf("description must be atleast 5 characters long")
	}

	if updates.Name == "" && updates.Description == "" {
		return nil, fmt.Errorf("no updates")
	}

	if updates.Name == existingProject.Name {
		return nil, fmt.Errorf("please provide a new name or omit the name field input")
	}

	if existingProject.Description != "" {
		if updates.Description == existingProject.Description {
			return nil, fmt.Errorf("failed to update project description")
		}
	}

	if updated, err := s.projectRepo.UpdateProjectByID(ctx, projectID, updates); !updated {
		return nil, err
	}

	slog.Info("project updated", "project_id", projectID)

	if websocket.WsManager != nil {
		payload := websocket.ProjectUpdatedPayload{
			Updates: updates,
		}
		websocket.WsManager.BroadcastToRoom("project:"+projectID.String() , websocket.EventProjectUpdated , payload)
	}else {
		slog.Info("failed to send event via websocket for update project" , "project_id" , projectID)
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.UpdateProject,
		UserID: &userID,
		OrgID: &existingProject.OrgID,
		ProjectID: &existingProject.ID,
		Metadata: map[string]any{
			"updats" : updates,
		},
	}
	if err := s.activityProducer.Publish(event);err != nil {
		slog.Info("failed to publish update project event")
	}
	return &updatedProject, nil
}

func (s *projectService) DeleteProject(ctx context.Context, projectID , userID uuid.UUID) (bool, error) {
	project, findProjectErr := s.projectRepo.FindProjectByID(ctx, projectID)
	if findProjectErr != nil {
		return false, fmt.Errorf("failed to find project : %w", findProjectErr)
	}
	slog.Info("deleting project", "project_id", projectID, "name", project.Name, "deleted_by", userID)

	teamMembers , _ := s.teamRepo.FindTeamMembers(ctx , project.TeamID)
	user , _ := s.userRepo.FindUserByID(ctx , userID) 
	if websocket.Publisher != nil && len(teamMembers) > 0 {
		userIDs := make([]uuid.UUID, len(teamMembers))
		for i, member := range teamMembers {
			userIDs[i] = member.UserID
		}
		payload := websocket.ProjectDeletedPayload{
			ProjectName: project.Name,
			DeletedBy: user.FullName,
		}
		// non blocking batch publish
		go websocket.Publisher.PublishToMany(userIDs , websocket.EventProjectDeleted , payload)
	}else {
		slog.Warn("failed to send event via websocket for delete project" , "len_members" , len(teamMembers))
	}

	if deleted, deleteErr := s.projectRepo.DeleteProjectByID(ctx, project.ID); !deleted {
		slog.Error("failed to delete project", "project_id", projectID, "error", deleteErr)
		return false, deleteErr
	}
	slog.Info("project deleted successfully" , "project_id" , projectID)
	// Publish activity event for logging
	event := models.ActivityEvent{
		EventName: events.ActivityEvents.DeleteProject,
		UserID: &userID,
		OrgID: &project.OrgID,
		ProjectID: &projectID,
		Metadata: map[string]any{
			"project_name" : project.Name,
		},
	}
	if err := s.activityProducer.Publish(event);err != nil {
		slog.Warn("failed to publish delete project event" , "error" , err.Error())
	}
	// delete related project tasks and comments
	// Publish cleanup event
	cleanupEvent := consumers.ProjectCleanupEvent{
		ProjectID: projectID,
		OrgID:     project.OrgID,
		DeletedBy: userID,
		Timestamp: time.Now().UTC(),
	}
	if err := s.publishProjectCleanupEvent(s.queueConn, cleanupEvent); err != nil {
		slog.Warn("WARNING: failed to publish project cleanup event" , "error" , err.Error())
	}
	return true, nil
}


func (s *projectService) publishProjectCleanupEvent(conn *amqp.Connection, event consumers.ProjectCleanupEvent) error {
	slog.Info("publishing project cleanup event", "project_id", event.ProjectID)
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	
	_, err = ch.QueueDeclare(constants.PROJECT_CLEANUP_QUEUE, true, false, false, false, nil)
	if err != nil {
		return err
	}
	
	body, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal event" , "error" , err.Error())
		return err
	}
	slog.Info("publishing cleanup event ...")
	return ch.Publish(
		"",
		constants.PROJECT_CLEANUP_QUEUE,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}