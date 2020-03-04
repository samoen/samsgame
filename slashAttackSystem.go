package main

import (
	"time"

	"github.com/hajimehoshi/ebiten"
)

type slasher struct {
	ent           *playerent
	slashPressed  bool
	animating     bool
	slashLine     *shape
	lastActiveDir directions
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
	// s.events = time.NewTicker(20 * time.Millisecond).C
	return s
}

func (s *slashAttackSystem) removeSlashee(p *playerent) {
	for i, renderable := range s.slashees {
		if p == renderable {
			if i < len(s.slashees)-1 {
				copy(s.slashees[i:], s.slashees[i+1:])
			}
			s.slashees[len(s.slashees)-1] = nil
			s.slashees = s.slashees[:len(s.slashees)-1]
		}
	}
}

func (s *slashAttackSystem) work() {
	// select {
	// case <-s.events:
	for _, bot := range s.slashers {
		if ebiten.IsKeyPressed(ebiten.KeyX) {
			bot.slashPressed = true
		} else {
			bot.slashPressed = false
		}
		select {
		case <-s.slashAnimationTimer:
			bot.animating = false
			renderingSystem.removeShape(bot.slashLine)
		default:
			keepOnPlayer := func() {

				midPlayer := bot.ent.rectangle.location
				midPlayer.x += bot.ent.rectangle.dimens.width / 2
				midPlayer.y += bot.ent.rectangle.dimens.height / 2
				hitRange := 30
				if (bot.lastActiveDir.left || bot.lastActiveDir.right) && (bot.lastActiveDir.down || bot.lastActiveDir.up) {
					hitRange = int(float64(hitRange) * 0.707)
				}

				moveTipX := 0
				if bot.lastActiveDir.right {
					moveTipX = hitRange
				} else if bot.lastActiveDir.left {
					moveTipX = -hitRange
				}
				hitTipX := midPlayer.x + moveTipX

				moveTipY := 0
				if bot.lastActiveDir.up {
					moveTipY = -hitRange
				} else if bot.lastActiveDir.down {
					moveTipY = hitRange
				}
				hitTipY := midPlayer.y + moveTipY

				for i := 0; i < len(bot.slashLine.lines); i++ {
					bot.slashLine.lines[i].p1 = midPlayer
					bot.slashLine.lines[i].p2 = location{
						hitTipX,
						hitTipY,
					}
				}
			}
			if bot.ent.directions.down ||
				bot.ent.directions.up ||
				bot.ent.directions.right ||
				bot.ent.directions.left {
				bot.lastActiveDir = bot.ent.directions
			}
			if bot.animating {
				keepOnPlayer()

			found:
				for _, slashee := range s.slashees {
					for _, slasheeLine := range slashee.rectangle.shape.lines {
						for _, bladeLine := range bot.slashLine.lines {
							if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
								renderingSystem.removeShape(slashee.rectangle.shape)
								collideSystem.removeMover(slashee)
								s.removeSlashee(slashee)
								break found
							}
						}
					}
				}
			} else {
				if bot.slashPressed {
					s.slashAnimationTimer = time.NewTimer(300 * time.Millisecond).C
					bot.animating = true
					keepOnPlayer()
					renderingSystem.addShape(bot.slashLine)
				}
			}

		}
	}
	// default:
	// }
}
