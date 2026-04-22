package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/consumers"
	"github.com/amirrajj-dev/taskio/internal/events"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/websocket"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type TaskService interface {
	CreateTask(ctx context.Context, projectID, userID uuid.UUID, title, description, priority, status string, dueDate *time.Time, assignedTo, parentTaskID *string) (*models.Task, error)
	GetProjectTasks(ctx context.Context, projectID uuid.UUID) ([]*models.Task, error)
	GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error)
	UpdateTask(ctx context.Context, taskID, teamID, userID uuid.UUID, title, description, priority, status, assignedTo *string, dueDate *time.Time) (*models.Task, error)
	DeleteTask(ctx context.Context, taskID, userID uuid.UUID) (bool, error)
	GetSubTasks(ctx context.Context, taskID uuid.UUID) ([]*models.Task, error)
}

type taskService struct {
	taskRepo         repositories.TaskRepository
	projectRepo      repositories.ProjectRepository
	teamRepo         repositories.TeamRepository
	queueConn        *amqp.Connection
	activityProducer *queue.ActivityProducer
}

var TasksService *taskService

func NewTaskService(producer *queue.ActivityProducer) {
	TasksService = &taskService{
		taskRepo:         repositories.TaskRepo,
		projectRepo:      repositories.ProjectRepo,
		teamRepo:         ProjectServicee.teamRepo,
		queueConn:        queue.QueueConn,
		activityProducer: producer,
	}
}

func (s *taskService) CreateTask(ctx context.Context, projectID, userID uuid.UUID, title, description, priority, status string, dueDate *time.Time, assignedTo, parentTaskID *string) (*models.Task, error) {
	slog.Info("creating task", "title", title, "project_id", projectID, "user_id", userID)
	existingTask, _ := s.taskRepo.FindTaskByTitle(ctx, strings.TrimSpace(title))
	if existingTask.Title != "" {
		return nil, fmt.Errorf("task with this title already exists")
	}

	existingProject, findProjectErr := s.projectRepo.FindProjectByID(ctx, projectID)
	if findProjectErr != nil {
		return nil, findProjectErr
	}

	if existingProject == nil {
		return nil, fmt.Errorf("unexpected: project not found")
	}

	now := time.Now().UTC()
	task := models.Task{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Title:       strings.TrimSpace(title),
		Description: description,
		Status:      status,
		Priority:    priority,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if assignedTo != nil {
		validAsignedToUUID, parseErr := uuid.Parse(strings.TrimSpace(*assignedTo))
		if parseErr != nil {
			return nil, fmt.Errorf("invalid assigned to id")
		}
		existingTeamMember, findTeamMemberErr := s.teamRepo.FindTeamMember(ctx, existingProject.TeamID, validAsignedToUUID)
		if findTeamMemberErr != nil {
			return nil, findTeamMemberErr
		}
		if existingTeamMember == nil {
			return nil, fmt.Errorf("unexpected:team member not found")
		}
		task.AssingedTo = validAsignedToUUID
	}
	if dueDate != nil {
		task.DueDate = dueDate
	}
	if parentTaskID != nil {
		validParentTaskUUID, parseErr := uuid.Parse(strings.TrimSpace(*parentTaskID))
		if parseErr != nil {
			return nil, fmt.Errorf("invalid assigned to id")
		}
		existingTask, findTaskErr := s.taskRepo.FindTaskByID(ctx, validParentTaskUUID)
		if findTaskErr != nil {
			return nil, findTaskErr
		}
		if existingTask == nil {
			return nil, fmt.Errorf("unexpected: task not found")
		}
		validParentTaskUUIDPtr := &validParentTaskUUID
		task.ParentTaskID = validParentTaskUUIDPtr
	} else {
		task.ParentTaskID = nil
	}
	createdTask, createTaskErr := s.taskRepo.CreateTask(ctx, task)
	if createTaskErr != nil {
		return nil, createTaskErr
	}
	slog.Info("task created", "task_id", createdTask.ID, "title", title, "assigned_to", task.AssingedTo)

	// send websocket notification
	if websocket.WsManager != nil {
		payload := websocket.TaskAssignedPayload{
			TaskID:      createdTask.ID,
			Title:       createdTask.Title,
			AssignedBy:  userID,
			AssignedTo:  createdTask.AssingedTo,
			ProjectID:   createdTask.ProjectID,
			ProjectName: existingProject.Name,
		}
		if createdTask.DueDate != nil {
			payload.DueDate = createdTask.DueDate
		}
		websocket.WsManager.BroadcastToUser(createdTask.AssingedTo, websocket.EventTaskAssigned, payload)
		websocket.WsManager.BroadcastToRoom("project:"+createdTask.ProjectID.String(), websocket.EventTaskAssigned, payload)
	} else {
		slog.Warn("failed to send event via websocket for create task", "task_title", title)
	}

	event := models.ActivityEvent{
		UserID:    &userID,
		ProjectID: &projectID,
		TaskID:    &createdTask.ID,
		OrgID:     &existingProject.ID,
		Metadata: map[string]any{
			"task_title":  createdTask.Title,
			"priority":    createdTask.Priority,
			"assigned_to": createdTask.AssingedTo,
		},
	}
	if createdTask.ParentTaskID != nil {
		event.EventName = events.ActivityEvents.CreateSubTask
		event.Metadata["parent_task_id"] = createdTask.ParentTaskID
		if err := s.activityProducer.Publish(event); err != nil {
			slog.Warn("failed to publish create sub task event", "task_id", createdTask.ID, "parent_task_id", createdTask.ParentTaskID)
		}
	} else {
		event.EventName = events.ActivityEvents.CreateTask
		if err := s.activityProducer.Publish(event); err != nil {
			slog.Warn("failed to publish create task event", "task_id", createdTask.ID)
		}
	}
	return createdTask, nil
}

func (s *taskService) GetProjectTasks(ctx context.Context, projectID uuid.UUID) ([]*models.Task, error) {
	slog.Info("fetching project tasks", "project_id", projectID)
	existingProject, findProjectErr := s.projectRepo.FindProjectByID(ctx, projectID)
	if findProjectErr != nil {
		return nil, findProjectErr
	}

	if existingProject == nil {
		return nil, fmt.Errorf("unexpected: project not found")
	}
	tasks, findProjectTasksErr := s.taskRepo.FindProjectTasks(ctx, projectID)

	if findProjectTasksErr != nil {
		return nil, findProjectTasksErr
	}
	slog.Info("project tasks fetched", "project_id", projectID, "count", len(tasks))
	return tasks, nil
}

func (s *taskService) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	slog.Info("fetching task", "task_id", taskID)
	task, err := s.taskRepo.FindTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	slog.Error("failed to fetch task", "task_id", taskID, "error", err)
	return task, nil
}

func (s *taskService) UpdateTask(ctx context.Context, taskID, teamID, userID uuid.UUID,
	title, description, priority, status, assignedTo *string, dueDate *time.Time) (*models.Task, error) {

	slog.Info("updating task", "task_id", taskID, "user_id", userID)

	validPriorities := []string{"low", "medium", "high", "urgent"}
	validStatuses := []string{"todo", "in_progress", "review", "done"}
	existingTask, getTaskErr := s.GetTaskByID(ctx, taskID)
	if getTaskErr != nil {
		return nil, getTaskErr
	}

	if existingTask == nil {
		return nil, fmt.Errorf("task not found")
	}

	updates := &models.Task{}
	hasChanges := false
	if *title != "" {
		trimmedTitle := strings.TrimSpace(*title)
		if trimmedTitle != "" && len(trimmedTitle) >= 4 && existingTask.Title != trimmedTitle {
			updates.Title = trimmedTitle
			hasChanges = true
		}
	}
	if *description != "" {
		trimmedDesc := strings.TrimSpace(*description)
		if existingTask.Description != trimmedDesc {
			updates.Description = trimmedDesc
			hasChanges = true
		}
	}
	if *priority != "" {
		if !slices.Contains(validPriorities, strings.TrimSpace(*priority)) {
			return nil, fmt.Errorf("Priority must be one of low , medium , high , urgent.")
		}
		if existingTask.Priority != *priority {
			updates.Priority = *priority
			hasChanges = true
		}
	}
	if *status != "" {
		if !slices.Contains(validStatuses, strings.TrimSpace(*status)) {
			return nil, fmt.Errorf("Status must be one of todo , in_progress , review , done.")
		}
		if existingTask.Status != *status {
			updates.Status = *status
			hasChanges = true
		}
	}
	if assignedTo != nil {
		validAssignedTo, parseErr := uuid.Parse(strings.TrimSpace(*assignedTo))
		if parseErr != nil {
			return nil, fmt.Errorf("invalid assigned to id")
		}
		if existingTask.AssingedTo != validAssignedTo {
			// Check team membership
			teamMember, findTeamMemberErr := s.teamRepo.FindTeamMember(ctx, teamID, validAssignedTo)
			if findTeamMemberErr != nil {
				return nil, findTeamMemberErr
			}
			if teamMember == nil {
				return nil, fmt.Errorf("team member not found")
			}
			updates.AssingedTo = validAssignedTo
			hasChanges = true
		}
	}
	if dueDate != nil {
		if existingTask.DueDate == nil || !existingTask.DueDate.Equal(*dueDate) {
			updates.DueDate = dueDate
			hasChanges = true
		}
	}
	if !hasChanges {
		fmt.Println("im returning existing task")
		return existingTask, nil
	}

	updatedTask, updateErr := s.taskRepo.UpdateTaskByID(ctx, taskID, *updates)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update: %w", updateErr)
	}
	slog.Info("task updated", "task_id", taskID)

	if websocket.WsManager != nil {
		payload := websocket.TaskUpdatedPayload{
			TaskID:    updatedTask.ID,
			Title:     updatedTask.Title,
			UpdatedBy: userID,
			Changes:   updates,
		}
		websocket.WsManager.BroadcastToUser(updatedTask.AssingedTo, websocket.EventTaskUpdated, payload)
		websocket.WsManager.BroadcastToRoom("project:"+updatedTask.ProjectID.String(), websocket.EventTaskUpdated, payload)
	} else {
		slog.Warn("failed to send event via websocket for update task" , "task_title" , updatedTask.Title)
	}

	event := models.ActivityEvent{
		EventName: events.ActivityEvents.UpdateTask,
		UserID:    &userID,
		ProjectID: &existingTask.ProjectID,
		TaskID:    &taskID,
		TeamID:    &teamID,
		Metadata: map[string]any{
			"updates": updates,
		},
	}

	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish update task event" , "error" , err.Error())
	}

	return updatedTask, nil
}

func (s *taskService) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) (bool, error) {
	existingTask, findTaskErr := s.taskRepo.FindTaskByID(ctx, taskID)
	if findTaskErr != nil {
		return false, findTaskErr
	}
	slog.Info("deleting task", "task_id", taskID, "title", existingTask.Title, "deleted_by", userID)
	if existingTask == nil {
		return false, fmt.Errorf("unexpected: task not found")
	}

	if websocket.WsManager != nil {
		payload := websocket.TaskDeletedPayload{
			TaskID: taskID,
			Title:  existingTask.Title,
		}
		websocket.WsManager.BroadcastToRoom("project:"+existingTask.ProjectID.String(), websocket.EventTaskDeleted, payload)
	} else {
		slog.Warn("failed to send event via websocket for delete task" , "task_title" , existingTask.Title)
	}

	// delete main task
	deleted, deleteErr := s.taskRepo.DeleteTaskByID(ctx, taskID)
	if !deleted {
		slog.Error("failed to delete task", "task_id", taskID, "error", deleteErr)
		return false, fmt.Errorf("failed to delete task %s: %w", existingTask.Title, deleteErr)
	}

	slog.Info("task deleted successfully" , "deleted_by", userID , "task_id" , taskID)

	// Publish cleanup event for subtasks and comments
	cleanupEvent := consumers.TaskCleanupEvent{
		TaskID:    taskID,
		DeletedBy: userID,
		ProjectID: existingTask.ProjectID,
		Timestamp: time.Now().UTC(),
	}

	if err := s.publishTaskCleanupEvent(s.queueConn, cleanupEvent); err != nil {
		slog.Error("WARNING: failed to publish task cleanup event" , "error" , err.Error())
	}

	// Publish activity event for logging
	event := models.ActivityEvent{
		EventName: events.ActivityEvents.DeleteTask,
		UserID:    &userID,
		TaskID:    &taskID,
		ProjectID: &existingTask.ProjectID,
		Metadata: map[string]any{
			"deleted_task_title": existingTask.Title,
		},
	}
	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish delete task event" , "error" , err.Error())
	}
	return true, nil
}

func (s *taskService) GetSubTasks(ctx context.Context, taskID uuid.UUID) ([]*models.Task, error) {
	slog.Info("fetching subtasks", "parent_task_id", taskID)
	existingTask, findTaskErr := s.taskRepo.FindTaskByID(ctx, taskID)
	if findTaskErr != nil {
		return nil, findTaskErr
	}
	if existingTask == nil {
		return nil, fmt.Errorf("unexpected: task not found")
	}
	subTasks, findTaskSubTasksErr := s.taskRepo.FindTaskSubTasks(ctx, taskID)
	if findTaskSubTasksErr != nil {
		return nil, findTaskSubTasksErr
	}
	slog.Info("subtasks fetched", "parent_task_id", taskID, "count", len(subTasks))
	return subTasks, nil
}

func (s *taskService) publishTaskCleanupEvent(conn *amqp.Connection, event consumers.TaskCleanupEvent) error {
	slog.Info("publishing task cleanup event", "task_id", event.TaskID)
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(constants.TASK_CLEANUP_QUEUE, true, false, false, false, nil)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		slog.Info("failed to marshal event" , "error" , err.Error() , "queue" , constants.TASK_CLEANUP_QUEUE)
		return err
	}

	slog.Info("publishing task cleanup event ...")

	return ch.Publish(
		"",
		constants.TASK_CLEANUP_QUEUE,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}
