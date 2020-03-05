package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten"
)

type slasher struct {
	ent            *playerent
	animating      bool
	slashLine      *shape
	startangle     float64
	animationCount float64
}

var slashSystem = newSlashAttackSystem()

type slashAttackSystem struct {
	slashAnimationTimer <-chan time.Time
	slashers            []*slasher
	slashees            []*playerent
	blockers            []*shape
}

func (s *slashAttackSystem) addBlocker(b *shape) {
	s.blockers = append(s.blockers, b)
}

func newSlashAttackSystem() slashAttackSystem {
	s := slashAttackSystem{}
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

func newLinePolar(loc location, length int, angle float64) line {
	xpos := int(float64(length)*math.Cos(angle)) + loc.x
	ypos := int(float64(length)*math.Sin(angle)) + loc.y
	return line{loc, location{xpos, ypos}}
}

func (s *slashAttackSystem) work() {
	// select {
	// case <-s.events:
	for _, bot := range s.slashers {
		stopslash := func() {
			s.slashAnimationTimer = nil
			bot.animationCount = 0
			bot.animating = false
			renderingSystem.removeShape(bot.slashLine)
		}
		select {
		case <-s.slashAnimationTimer:
			stopslash()
			return
		default:
		}
		keepOnPlayer := func() {
			midPlayer := bot.ent.rectangle.location
			midPlayer.x += bot.ent.rectangle.dimens.width / 2
			midPlayer.y += bot.ent.rectangle.dimens.height / 2
			bot.animationCount -= 0.16
			for i := 0; i < len(bot.slashLine.lines); i++ {
				rotLine := newLinePolar(midPlayer, 50, bot.animationCount+bot.startangle)
				bot.slashLine.lines[i] = rotLine
			}

		foundSlashee:
			for _, slashee := range s.slashees {
				for _, slasheeLine := range slashee.rectangle.shape.lines {
					for _, bladeLine := range bot.slashLine.lines {
						if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
							renderingSystem.removeShape(slashee.rectangle.shape)
							collideSystem.removeMover(slashee)
							s.removeSlashee(slashee)
							break foundSlashee
						}
					}
				}
			}
		foundBlocker:
			for _, slashee := range s.blockers {
				for _, slasheeLine := range slashee.lines {
					for _, bladeLine := range bot.slashLine.lines {
						if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
							stopslash()
							break foundBlocker
						}
					}
				}
			}
		}

		if bot.animating {
			keepOnPlayer()
		}

		if !bot.animating {
			if bot.ent.directions.down ||
				bot.ent.directions.up ||
				bot.ent.directions.right ||
				bot.ent.directions.left {
				hitRange := 1
				moveTipX := 0
				if bot.ent.directions.right {
					moveTipX = hitRange
				} else if bot.ent.directions.left {
					moveTipX = -hitRange
				}
				moveTipY := 0
				if bot.ent.directions.up {
					moveTipY = -hitRange
				} else if bot.ent.directions.down {
					moveTipY = hitRange
				}
				bot.startangle = math.Atan2(float64(moveTipY), float64(moveTipX))
				bot.startangle += 1.6
			}
			if ebiten.IsKeyPressed(ebiten.KeyX) {
				s.slashAnimationTimer = time.NewTicker(310 * time.Millisecond).C
				bot.animating = true
				renderingSystem.addShape(bot.slashLine)
				keepOnPlayer()
			}
		}

	}
	// default:
	// }
}
