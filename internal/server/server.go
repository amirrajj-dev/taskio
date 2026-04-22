package server

import (
	_ "github.com/amirrajj-dev/taskio/docs"
	"github.com/amirrajj-dev/taskio/internal/handlers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/middlewares"
	"github.com/amirrajj-dev/taskio/internal/routes"
	"github.com/amirrajj-dev/taskio/internal/websocket"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewServer(rabbitConn *amqp.Connection) *gin.Engine {
	r := gin.Default()
	healthHandler := handlers.NewHealthHandler()
	r.GET("/api/health", healthHandler.HealthCheck)

	// Initialize WebSocket manager
	wsManager := websocket.NewWebSocketManager()
	websocket.InitPublisher(wsManager)

	// WebSocket endpoint
	r.GET("/ws", wsManager.HandleWebSocket)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	producer := queue.NewActivityProducer(rabbitConn)
	
	r.Use(middlewares.CorsMiddleware())
	
	routes.RegisterAuthRoutes(r, producer)
	routes.RegisterOrgRoutes(r, producer)
	routes.RegisterTeamRoutes(r, producer)
	routes.RegisterProjectRoutes(r, producer)
	routes.RegisterTaskRoutes(r, producer)
	routes.RegisterCommentRoutes(r, producer)
	routes.RegisterActivityRoutes(r)
	return r
}
