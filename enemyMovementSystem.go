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

func (b *enemyMovementSystem) addEnemy(m *acceleratingEnt) {
	b.movers = append(b.movers, m)
}

func (b *enemyMovementSystem) work() {
	select {
	case <-b.events:
		for _, bot := range b.movers {
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
func (c *enemyMovementSystem) removeEnemyMover(s *rectangle) {
	for i, renderable := range c.movers {
		if s == renderable.rect {
			if i < len(c.movers)-1 {
				copy(c.movers[i:], c.movers[i+1:])
			}
			c.movers[len(c.movers)-1] = nil
			c.movers = c.movers[:len(c.movers)-1]
			break
		}
	}
}
