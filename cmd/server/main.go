package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting Peril server...")

	const mqConnection = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(mqConnection)
	if err != nil {
		fmt.Errorf("Could not connect to RabbitMQ")
	}
	defer conn.Close()
	fmt.Println("RabbitMQ connection established")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
}
