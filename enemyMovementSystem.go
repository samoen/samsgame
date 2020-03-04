package main

import (
	"math/rand"
	"time"
)

type enemyMovementSystem struct {
	events <-chan time.Time
	bots   []*playerent
}

var botsMoveSystem = newEnemyMovementSystem()

func newEnemyMovementSystem() enemyMovementSystem {
	b := enemyMovementSystem{}
	b.events = time.NewTicker(time.Duration(500) * time.Millisecond).C
	return b
}

func (b *enemyMovementSystem) addEnemy(m *playerent) {
	b.bots = append(b.bots, m)
}

func (b *enemyMovementSystem) work() {
	select {
	case <-b.events:
		for _, bot := range b.bots {
			bot.directions = directions{
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
			}
		}
	default:
	}
}
