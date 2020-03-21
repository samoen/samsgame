package main

import (
	"github.com/hajimehoshi/ebiten"
)

type directions struct {
	right, down, left, up bool
}

type moveSpeed struct {
	currentSpeed int
}

var playerControllables []*acceleratingEnt

func addPlayerControlled(m *acceleratingEnt) {
	playerControllables = append(playerControllables, m)
	m.rect.shape.systems = append(m.rect.shape.systems, playerControlled)
}

func updatePlayerControl() {
	// select {
	// case <-b.events:
	for _, bot := range playerControllables {
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
