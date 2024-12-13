package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("could not subscribe to %s: %v", queueName, err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	messages, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume %s: %v", queueName, err)
	}

	go func() {
		defer ch.Close()
		for message := range messages {
			var data T
			err := json.Unmarshal(message.Body, &data)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			acktype := handler(data)

			if acktype == Ack {
				message.Ack(false)
			} else if acktype == NackRequeue {
				message.Nack(false, true)
			} else {
				message.Nack(false, false)
			}
		}
	}()

	return nil
}
