package pubsub

import (
	"bytes"
	"encoding/gob"
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
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		if err != nil {
			return target, err
		}
		return target, nil
	})
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(data []byte) (T, error) {
		buffer := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buffer)
		var target T
		err := dec.Decode(&target)
		if err != nil {
			return target, err
		}
		return target, nil
	})
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("could not subscribe to %s: %v", queueName, err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	ch.Qos(10, 0, false)
	messages, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume %s: %v", queueName, err)
	}

	go func() {
		defer ch.Close()
		for message := range messages {
			data, err := unmarshaller(message.Body)
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
