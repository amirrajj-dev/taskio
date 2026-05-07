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

type TaskHandler struct {
	taskService services.TaskService
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		taskService: services.TasksService,
	}
}

// CreateTask godoc
// @Summary      Create task
// @Description  Creates a new task within a project (owner or admin only)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectId   path      string                 true  "Project ID (UUID)"
// @Param        request     body      dtos.CreateTaskRequest true  "Task details"
// @Success      201  {object}  utils.SuccessResponse{data=models.Task}  "Task created successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /projects/{projectId}/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	projectID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("projectId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid project id", c.Request.URL.Path))
		return
	}
	userID, _ := utils.GetUserIDFromContext(c)
	var req dtos.CreateTaskRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}
	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	task, createTaskErr := h.taskService.CreateTask(c.Request.Context(), projectID, *userID, req.Title, req.Description, req.Priority, req.Status, req.DueDate, req.AssignedTo, nil)
	if createTaskErr != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError("failed to create task : "+createTaskErr.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusCreated, utils.NewSuccessResponse("task created succesfully", task, c.Request.URL.Path))
}

// CreateSubTask godoc
// @Summary      Create subtask
// @Description  Creates a new subtask under an existing parent task (owner or admin only)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectId   path      string                 true  "Project ID (UUID)"
// @Param        taskId      path      string                 true  "Parent Task ID (UUID)"
// @Param        request     body      dtos.CreateTaskRequest true  "Subtask details"
// @Success      201  {object}  utils.SuccessResponse{data=models.Task}  "Subtask created successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /projects/{projectId}/tasks/{taskId}/subtasks [post]
func (h *TaskHandler) CreateSubTask(c *gin.Context) {
	projectID, _ := uuid.Parse(strings.TrimSpace(c.Param("projectId")))
	userID, _ := utils.GetUserIDFromContext(c)
	parentTaskID, parseTaskIdErr := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
	if parseTaskIdErr != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError("invalid task id", c.Request.URL.Path))
		return
	}
	str := parentTaskID.String()
	parentTaskIDPtr := &str
	var req dtos.CreateTaskRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}
	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	task, createTaskErr := h.taskService.CreateTask(c.Request.Context(), projectID, *userID, req.Title, req.Description, req.Priority, req.Status, req.DueDate, req.AssignedTo, parentTaskIDPtr)
	if createTaskErr != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(createTaskErr.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusCreated, utils.NewSuccessResponse("task created successfully", task, c.Request.URL.Path))
}

// GetProjectTasks godoc
// @Summary      Get project tasks
// @Description  Retrieves all tasks for a specific project
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectId   path      string  true  "Project ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse{data=[]models.Task}  "Tasks fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid project ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Not a project member"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /projects/{projectId}/tasks [get]
func (h *TaskHandler) GetProjectTasks(c *gin.Context) {
	projectID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("projectId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid project id", c.Request.URL.Path))
		return
	}
	tasks, getProjectTasksErr := h.taskService.GetProjectTasks(c.Request.Context(), projectID)
	if getProjectTasksErr != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(getProjectTasksErr.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("tasks fetched successfully", tasks, c.Request.URL.Path))
}

// GetTask godoc
// @Summary      Get task by ID
// @Description  Retrieves detailed information about a specific task
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskId      path      string  true  "Task ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse{data=models.Task}  "Task fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid task ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "No access to task"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /tasks/{taskId} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskId, parseErr := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid task id", c.Request.URL.Path))
		return
	}
	task, err := h.taskService.GetTaskByID(c.Request.Context(), taskId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("task fetched successfully", task, c.Request.URL.Path))
}

// UpdateTask godoc
// @Summary      Update task
// @Description  Updates task details (owner only)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskId      path      string                 true  "Task ID (UUID)"
// @Param        request     body      dtos.UpdateTaskRequest true  "Updated task details"
// @Success      200  {object}  utils.SuccessResponse{data=models.Task}  "Task updated successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Only owner can update"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /tasks/{taskId} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	taskID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid task id", c.Request.URL.Path))
		return
	}
	id, exists := c.Get("teamID")
	if !exists {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError("something goes wrong", c.Request.URL.Path))
		return
	}
	teamID, ok := id.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid team id", c.Request.URL.Path))
	}
	var req dtos.UpdateTaskRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}
	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	updatedTask, updateErr := h.taskService.UpdateTask(c.Request.Context(), taskID, teamID, *userID , &req.Title, &req.Description, &req.Priority, &req.Status, req.AssignedTo, req.DueDate)
	if updateErr != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(updateErr.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("task updated successfully", updatedTask, c.Request.URL.Path))
}

// DeleteTask godoc
// @Summary      Delete task
// @Description  Permanently deletes a task (owner only)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskId      path      string  true  "Task ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse  "Task deleted successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid task ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Only owner can delete"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /tasks/{taskId} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	taskID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid task id", c.Request.URL.Path))
		return
	}
	if deleted, err := h.taskService.DeleteTask(c.Request.Context(), taskID, *userID); !deleted {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("task deleted successfully", nil, c.Request.URL.Path))
}

// GetSubTasks godoc
// @Summary      Get subtasks
// @Description  Retrieves all subtasks for a specific parent task
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskId      path      string  true  "Parent Task ID (UUID)"
// @Success      200  {object}  utils.SuccessResponseWithTotal{data=[]models.Task}  "Subtasks fetched successfully"
// @Failure      400  {object}  errors.BasicError  "Invalid task ID"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      403  {object}  errors.BasicError  "Insufficient permissions"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /tasks/{taskId}/subtasks [get]
func (h *TaskHandler) GetSubTasks(c *gin.Context) {
	taskID, parseErr := uuid.Parse(strings.TrimSpace(c.Param("taskId")))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, errors.NewBasicError("invalid task id", c.Request.URL.Path))
		return
	}
	subTasks, getSubTasksErr := h.taskService.GetSubTasks(c.Request.Context(), taskID)
	if getSubTasksErr != nil {
		c.JSON(http.StatusInternalServerError, errors.NewBasicError(getSubTasksErr.Error(), c.Request.URL.Path))
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponseWithTotal("sub tasks fetched successfully", subTasks, int64(len(subTasks)), c.Request.URL.Path))
}
