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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Unable to get username: %v", err)
	}

	pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username), routing.PauseKey, pubsub.SimpleQueueTransient)

	gameState := gamelogic.NewGameState(username)

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
				fmt.Printf("Move units to %s completed!\n", move.ToLocation)
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
