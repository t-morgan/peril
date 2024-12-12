package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
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
			handler(data)
			message.Ack(false)
		}
	}()

	return nil
}
