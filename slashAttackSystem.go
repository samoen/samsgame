package main

import (
	"time"

	"github.com/hajimehoshi/ebiten"
)

type slasher struct {
	ent          *playerent
	slashPressed bool
	animating    bool
}

var slashSystem = newSlashAttackSystem()

type slashAttackSystem struct {
	slashAnimationTimer <-chan time.Time
	events              <-chan time.Time
	slashers            []*slasher
	slashees            []*playerent
}

func newSlashAttackSystem() slashAttackSystem {
	s := slashAttackSystem{}
	s.events = time.NewTicker(50 * time.Millisecond).C
	// s.slashAnimationTimer = time.NewTimer(300 * time.Millisecond).C

	return s
}

func (s *slashAttackSystem) work() {
	select {
	case <-s.events:
		for _, bot := range s.slashers {
			if ebiten.IsKeyPressed(ebiten.KeyX) {
				bot.slashPressed = true
			} else {
				bot.slashPressed = false
			}
			select {
			case <-s.slashAnimationTimer:
				bot.animating = false
			default:
				if bot.animating {
					slashLine := line{bot.ent.rectangle.location, location{bot.ent.rectangle.location.x + 30, bot.ent.rectangle.location.y + 30}}
					for _, slashee := range s.slashees {
						for _, slasheeLine := range slashee.rectangle.shape {
							if _, _, intersected := slashLine.intersects(slasheeLine); intersected {
								renderingSystem.removeShape(&slashee.rectangle.shape)
							}
						}
					}
				} else {
					if bot.slashPressed {
						s.slashAnimationTimer = time.NewTimer(300 * time.Millisecond).C
						bot.animating = true
					}
				}

			}

		}
	default:
	}
}
