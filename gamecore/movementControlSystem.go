package gamecore

import (
	"github.com/hajimehoshi/ebiten"
)

//Directions dirs
type Directions struct {
	Right bool `json:"right"`
	Down  bool `json:"down"`
	Left  bool `json:"left"`
	Up    bool `json:"up"`
}

var playerControllables = make(map[*entityid]*acceleratingEnt)

func addPlayerControlled(m *acceleratingEnt, id *entityid) {
	playerControllables[id] = m
	id.systems = append(id.systems, playerControlled)
}

func updatePlayerControl() {
	for _, bot := range playerControllables {
		bot.directions = Directions{
			ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight),
			ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown),
			ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft),
			ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp),
		}
		bot.atkButton = ebiten.IsKeyPressed(ebiten.KeyX)
	}
}
