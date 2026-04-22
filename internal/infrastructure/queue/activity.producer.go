package queue

import (
	"encoding/json"

	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)



type ActivityProducer struct {
	conn *amqp.Connection
}

func NewActivityProducer(conn *amqp.Connection) *ActivityProducer {
	return &ActivityProducer{conn: conn}
}

func (p *ActivityProducer) Publish(event models.ActivityEvent) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(constants.ACTIVITY_QUEUE, true, false, false, false, nil)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(event)

	return ch.Publish(
		"",
		constants.ACTIVITY_QUEUE,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}
