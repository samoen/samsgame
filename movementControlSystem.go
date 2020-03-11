package main

import (
	"github.com/hajimehoshi/ebiten"
)

var playerMoveSystem = newPlayerMovementSystem()

type directions struct {
	right, down, left, up bool
}

type moveSpeed struct {
	currentSpeed int
	maxSpeed     int
}

// type rectangle struct {
// 	rectangle *rectangle
// }

type playerMovementSystem struct {
	// events <-chan time.Time
	bots []*acceleratingEnt
}

func newPlayerMovementSystem() playerMovementSystem {
	b := playerMovementSystem{}
	// b.events = time.NewTicker(time.Duration(50) * time.Millisecond).C
	return b
}

func (b *playerMovementSystem) addPlayer(m *acceleratingEnt) {
	b.bots = append(b.bots, m)
}

func (b *playerMovementSystem) work() {
	// select {
	// case <-b.events:
	for _, bot := range b.bots {
		bot.directions = directions{
			ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight),
			ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown),
			ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft),
			ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp),
		}
		bot.atkButton = ebiten.IsKeyPressed(ebiten.KeyX)
	}
	// default:
	// }
}
