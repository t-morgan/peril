package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/t-morgan/peril/internal/routing"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	json, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(context.Background(), exchange, key, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        json,
		},
	)
}

func PublishGob[T any](ch *amqp.Channel, username string, val T) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(context.Background(), routing.ExchangePerilTopic, routing.GameLogSlug+"."+username, false, false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buffer.Bytes(),
		},
	)
}
