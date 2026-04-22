package app

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/consumers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/redis"
	"github.com/amirrajj-dev/taskio/internal/server"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/amirrajj-dev/taskio/internal/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	mail "github.com/wneessen/go-mail"
)

type EmailMessage struct {
	Subject string `json:"subject"`
	To      string `json:"to"`
	IsHtml  bool   `json:"isHtml"`
	Body    string `json:"body"`
	LinkUrl string `json:"linkUrl"`
}

func Bootstrap() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	configs.LoadConfig()
	database.ConnectToPostgresDb()
	redis.ConnectToRedis()
	conn := queue.ConnectToRabbitMQ()
	utils.InitValidate()
	router := server.NewServer(conn)
	srv := &http.Server{
		Addr:         configs.Configs.App.PORT,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server is up and running on port %s", configs.Configs.App.PORT)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Background workers for queue.
	go worker()
	go consumers.StartActivityWorker(queue.QueueConn)
	go consumers.StartOrgCleanupConsumer(queue.QueueConn)
	go consumers.StartProjectCleanupConsumer(queue.QueueConn)
	go consumers.StartTeamCleanupConsumer(queue.QueueConn)
	go consumers.StartTaskCleanupConsumer(queue.QueueConn)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Println("shuting down server ...")
	if websocket.WsManager != nil {
		for _ , client := range websocket.WsManager.Hub.Clients {
			client.Conn.Close()
		}
		log.Println("WebSocket connections closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown : %s", err)
	}
	database.DisconnectFromPostgresDb()
	defer queue.QueueConn.Close()
	defer redis.RedisClient.Close()
	log.Println("server exited")
}

// COSUMER
func worker() {
	conn, err := amqp.Dial(configs.Configs.RabbitMQ.RABBITMQ_URL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ : %v", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()
	queue, err := ch.QueueDeclare(constants.EMAIL_QUEUE, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to declare %s queue : %v", constants.EMAIL_QUEUE, err)
	}
	// Ensure the queue is configured for prefetch (process one message at a time)
	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}
	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}
	log.Println(" [*] Waiting for messages. To exit press CTRL+C")
	// Process messages in a loop
	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var emailPayload EmailMessage
			err := json.Unmarshal(d.Body, &emailPayload)
			if err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				// Nack the message with requeue=false if it's a bad message
				d.Nack(false, false)
				continue
			}
			// Send email using go-mail
			client, err := mail.NewClient("localhost",
				mail.WithPort(1025),
				mail.WithTLSPolicy(mail.NoTLS),
			)
			if err != nil {
				d.Nack(false, true)
				log.Printf("failed to create client : %v", err)
				continue
			}
			msg := mail.NewMsg()
			if err := msg.From("worker@example.com"); err != nil {
				log.Fatalf("failed to configure sender : %v", err)
			}
			if err := msg.To(emailPayload.To); err != nil {
				log.Fatalf("failed to configure to : %v", err)
			}
			msg.Subject(emailPayload.Subject)
			if emailPayload.IsHtml {
				msg.SetBodyString(mail.TypeTextHTML, emailPayload.Body)
			} else {
				msg.SetBodyString(mail.TypeTextPlain, emailPayload.Body)
			}
			if err := client.DialAndSend(msg); err != nil {
				log.Printf("failed to send email : %v", err)
				d.Nack(false, true)
			} else {
				log.Printf("Email sent to %s successfully.", emailPayload.To)
				d.Ack(false)
			}
		}
	}()
	log.Println(" [*] Consumer started. To exit press CTRL+C")
	<-forever
}
