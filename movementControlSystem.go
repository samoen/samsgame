package main

import (
	"time"
)

var botsMoveSystem = newPlayerMovementSystem(500)
var playerMoveSystem = newPlayerMovementSystem(50)

type directions struct {
	right, down, left, up bool
}

type moveSpeed struct {
	currentSpeed int
	maxSpeed     int
}

type playerent struct {
	rectangle  rectangle
	moveSpeed  moveSpeed
	directions directions
	getNewDirs func() directions
}

type botMovementSystem struct {
	events <-chan time.Time
	bots   []*playerent
}

func newPlayerMovementSystem(tickSpeed int) botMovementSystem {
	b := botMovementSystem{}
	b.events = time.NewTicker(time.Duration(tickSpeed) * time.Millisecond).C
	return b
}

func (b *botMovementSystem) addBot(m *playerent) {
	b.bots = append(b.bots, m)
}

func (b *botMovementSystem) work() {
	select {
	case <-b.events:
		for _, bot := range b.bots {
			bot.directions = bot.getNewDirs()
		}
	default:
	}
}
