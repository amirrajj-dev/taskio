package routes

import (
	"github.com/amirrajj-dev/taskio/internal/handlers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/middlewares"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/gin-gonic/gin"
)


func RegisterTaskRoutes(r *gin.Engine , producer *queue.ActivityProducer){
	repositories.NewTaskRepository()
	services.NewTaskService(producer)
	handler := handlers.NewTaskHandler()

	taskGroup := r.Group("/api/")
	taskGroup.Use(middlewares.AuthMiddleware())
	{
		taskGroup.POST("projects/:projectId/tasks" , middlewares.CheckProjectMemberShipAndRole([]string{"owner" , "admin"}) , handler.CreateTask)
		taskGroup.POST("projects/:projectId/tasks/:taskId/subtasks" , middlewares.CheckProjectMemberShipAndRole([]string{"owner" , "admin"}) , handler.CreateSubTask)
		taskGroup.GET("projects/:projectId/tasks" , middlewares.CheckProjectMemberShipAndRole([]string{}) , handler.GetProjectTasks)
		taskGroup.GET("tasks/:taskId" , middlewares.CheckTaskAccess([]string{"owner"}) , handler.GetTask)
		taskGroup.PUT("/tasks/:taskId" , middlewares.CheckTaskAccess([]string{"owner"}) , handler.UpdateTask)
		taskGroup.GET("/tasks/:taskId/subtasks" , middlewares.CheckTaskAccess([]string{"owner" , "admin"}) , handler.GetSubTasks)
		taskGroup.DELETE("/tasks/:taskId" , middlewares.CheckTaskAccess([]string{"owner"}) , handler.DeleteTask)
	}
}