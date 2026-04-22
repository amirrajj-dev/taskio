package consumers

import (
	"context"
	"encoding/json"
	"log"
	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	amqp "github.com/rabbitmq/amqp091-go"
)

func StartActivityWorker(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("worker failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(constants.ACTIVITY_QUEUE, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("worker failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("worker failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			var evt models.ActivityEvent
			if err := json.Unmarshal(msg.Body, &evt); err != nil {
				log.Println("bad event:", err)
				msg.Nack(false, false)
				continue
			}

			_ , err := repositories.ActivityRepo.CreateActivityFromEvent(context.Background() , evt)
			if err != nil {
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}()

	log.Println("Activity worker running...")
	select {}
}
