package routes

import (
	"github.com/amirrajj-dev/taskio/internal/handlers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/middlewares"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/gin-gonic/gin"
)


func RegisterCommentRoutes(r *gin.Engine , producer *queue.ActivityProducer) {
	repositories.NewCommentRepository()
	services.NewCommentService(producer)
	handler := handlers.NewCommentHandler()

	commentsGroup := r.Group("/api/tasks")
	commentsGroup.Use(middlewares.AuthMiddleware())
	{
		commentsGroup.POST("/:taskId/comments" , middlewares.CheckCommentAccess() , handler.CreateComment)
		commentsGroup.GET("/:taskId/comments" , middlewares.CheckCommentAccess() , handler.GetComments)
	}
}