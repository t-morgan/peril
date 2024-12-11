package main

import (
	"fmt"
	"log"

	"github.com/t-morgan/peril/internal/pubsub"
	"github.com/t-morgan/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const mqConnection = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(mqConnection)
	if err != nil {
		log.Fatalf("RabbitMQ connection failes: %v", err)
	}
	defer conn.Close()
	fmt.Println("RabbitMQ connection established")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Could not create channel: %v", err)
	}
	err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		log.Printf("Could not publish: %v", err)
	}
}
