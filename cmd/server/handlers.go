package main

import (
	"fmt"

	"github.com/t-morgan/peril/internal/gamelogic"
	"github.com/t-morgan/peril/internal/pubsub"
	"github.com/t-morgan/peril/internal/routing"
)

func handlerWriteLog() func(routing.GameLog) pubsub.Acktype {
	return func(gamelog routing.GameLog) pubsub.Acktype {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gamelog)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
