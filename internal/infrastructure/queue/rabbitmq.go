package queue

import (
	"log"

	"github.com/amirrajj-dev/taskio/internal/configs"
	amqp "github.com/rabbitmq/amqp091-go"
)

var QueueConn *amqp.Connection

func ConnectToRabbitMQ() *amqp.Connection {
	conn, err := amqp.Dial(configs.Configs.RabbitMQ.RABBITMQ_URL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ : %v", err)
	}
	QueueConn = conn
	log.Println("Connected to RabbitMq Succesfully")
	return conn
}
