package routes

import (
	"github.com/amirrajj-dev/taskio/internal/handlers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/middlewares"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, producer *queue.ActivityProducer) {
	repositories.NewUserRepository()
	services.NewAuthService()
	handler := handlers.NewAuthHandler(producer)

	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", middlewares.RegisterLimit(), handler.Register)
		authGroup.POST("/login", middlewares.LoginLimit(), handler.Login)
		authGroup.POST("/refresh", middlewares.AuthMiddleware(), handler.RefreshToken)
		authGroup.POST("/logout", middlewares.AuthMiddleware(), handler.LogOut)
	}
}
