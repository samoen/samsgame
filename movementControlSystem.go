package main

import (
	"github.com/hajimehoshi/ebiten"
)

type directions struct {
	right, down, left, up bool
}

var playerControllables = make(map[*entityid]*acceleratingEnt)

func addPlayerControlled(m *acceleratingEnt, id *entityid) {
	playerControllables[id] = m
	id.systems = append(id.systems, playerControlled)
}

func updatePlayerControl() {
	for _, bot := range playerControllables {
		bot.directions = directions{
			ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight),
			ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown),
			ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft),
			ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp),
		}
		bot.atkButton = ebiten.IsKeyPressed(ebiten.KeyX)
	}
}
