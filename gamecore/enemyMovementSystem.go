package gamecore

import (
	"math/rand"
)

type enemyController struct {
	aEnt         *acceleratingEnt
	controlCount int
}

var enemyControllers = make(map[*entityid]*enemyController)

func addEnemyController(m *enemyController, id *entityid) {
	enemyControllers[id] = m
	id.systems = append(id.systems, enemyControlled)
}

func enemyControlWork() {
	for _, bot := range enemyControllers {
		bot.controlCount--
		if bot.controlCount < 1 {
			bot.controlCount = rand.Intn(100)
			bot.aEnt.directions = Directions{
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
			}
			bot.aEnt.atkButton = rand.Intn(2) == 0
		}
	}
}
