package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/t-morgan/peril/internal/gamelogic"
	"github.com/t-morgan/peril/internal/pubsub"
	"github.com/t-morgan/peril/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")

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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Unable to get username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username,
		routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gameState))
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*", pubsub.SimpleQueueTransient, handlerMove(gameState, ch))
	if err != nil {
		log.Fatalf("could not subscribe to army_moves: %v", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*", pubsub.SimpleQueueDurable, handlerWar(gameState))
	if err != nil {
		log.Fatalf("could not subscribe to war: %v", err)
	}

clientLoop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			{
				err := gameState.CommandSpawn(words)
				if err != nil {
					fmt.Printf("Spawn command failed: %v\n", err)
				}
			}
		case "move":
			{
				move, err := gameState.CommandMove(words)
				if err != nil {
					fmt.Printf("Move command failed: %v\n", err)
					break
				}
				err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, move)
				if err != nil {
					log.Printf("Could not publish: %v", err)
					break
				}
				fmt.Println("Move published sucessfully")
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet")
		case "quit":
			gamelogic.PrintQuit()
			break clientLoop
		default:
			fmt.Printf("Unknown command: %s", words[0])
		}
	}
}
