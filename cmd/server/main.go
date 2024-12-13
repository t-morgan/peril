package main

import (
	"fmt"
	"log"

	"github.com/t-morgan/peril/internal/gamelogic"
	"github.com/t-morgan/peril/internal/pubsub"
	"github.com/t-morgan/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const mqConnection = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(mqConnection)
	if err != nil {
		log.Fatalf("RabbitMQ connection failed: %v", err)
	}
	defer conn.Close()
	fmt.Println("RabbitMQ connection established")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Could not create channel: %v", err)
	}

	err = pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, routing.GameLogSlug,
		routing.GameLogSlug+".*", pubsub.SimpleQueueDurable, handlerWriteLog())
	if err != nil {
		log.Fatalf("could not subscribe to game log: %v", err)
	}

	gamelogic.PrintServerHelp()

serverLoop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			{
				err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
				if err != nil {
					log.Printf("Could not publish: %v\n", err)
				}
			}
		case "resume":
			{
				err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
				if err != nil {
					log.Printf("Could not publish: %v\n", err)
				}
			}
		case "quit":
			fmt.Println("Exiting Server...")
			break serverLoop
		default:
			fmt.Printf("Unknown command: %s", words[0])
		}
	}
}
