package main

import (
	"fmt"
	"time"

	"github.com/t-morgan/peril/internal/gamelogic"
	"github.com/t-morgan/peril/internal/pubsub"
	"github.com/t-morgan/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")

		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")

		moveOutcome := gs.HandleMove(move)
		if moveOutcome == gamelogic.MoveOutcomeMakeWar {
			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.Player.Username, gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.Player,
			})
			if err != nil {
				fmt.Printf("Error publishing war: %v\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		if moveOutcome == gamelogic.MoveOutcomeSafe {
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		if outcome == gamelogic.WarOutcomeNotInvolved {
			return pubsub.NackRequeue
		}
		if outcome == gamelogic.WarOutcomeNoUnits {
			return pubsub.NackDiscard
		}
		switch outcome {
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			{
				err := pubsub.PublishGob(ch, gs.GetUsername(), routing.GameLog{
					CurrentTime: time.Now(),
					Username:    gs.GetUsername(),
					Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
				})
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return pubsub.NackRequeue
				}
			}
		case gamelogic.WarOutcomeDraw:
			{
				err := pubsub.PublishGob(ch, gs.GetUsername(), routing.GameLog{
					CurrentTime: time.Now(),
					Username:    gs.GetUsername(),
					Message:     fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
				})
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return pubsub.NackRequeue
				}
			}
		}
		if outcome == gamelogic.WarOutcomeOpponentWon ||
			outcome == gamelogic.WarOutcomeYouWon ||
			outcome == gamelogic.WarOutcomeDraw {
			return pubsub.Ack
		}
		fmt.Printf("Unknown outcome: %v - %v - %v\n", outcome, winner, loser)
		return pubsub.NackDiscard
	}
}
