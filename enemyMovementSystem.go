package main

import (
	"math/rand"
	"time"
)

type enemyMovementSystem struct {
	events <-chan time.Time
	movers []*acceleratingEnt
}

var botsMoveSystem = newEnemyMovementSystem()

func newEnemyMovementSystem() enemyMovementSystem {
	b := enemyMovementSystem{}
	b.events = time.NewTicker(time.Duration(500) * time.Millisecond).C
	return b
}

func (e *enemyMovementSystem) addEnemy(m *acceleratingEnt) {
	e.movers = append(e.movers, m)
	m.rect.shape.removals = append(m.rect.shape.removals, func() {
		e.removeEnemyMover(m.rect.shape)
	})
}

func (e *enemyMovementSystem) work() {
	select {
	case <-e.events:
		for _, bot := range e.movers {
			bot.directions = directions{
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
			}
			bot.atkButton = rand.Intn(2) == 0
		}
	default:
	}
}
func (e *enemyMovementSystem) removeEnemyMover(s *shape) {
	for i, renderable := range e.movers {
		if s == renderable.rect.shape {
			if i < len(e.movers)-1 {
				copy(e.movers[i:], e.movers[i+1:])
			}
			e.movers[len(e.movers)-1] = nil
			e.movers = e.movers[:len(e.movers)-1]
			break
		}
	}
}
