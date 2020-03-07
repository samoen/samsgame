package main

import (
	"math"

	"github.com/hajimehoshi/ebiten"
)

type slasher struct {
	ent            *playerent
	animating      bool
	slashLine      *shape
	startangle     float64
	animationCount float64
	onCooldown     bool
	cooldownCount  int
}

func newSlasher(p *playerent) *slasher {
	s := &slasher{}
	s.ent = p
	s.slashLine = &shape{[]line{line{}}}
	return s
}

var slashSystem = newSlashAttackSystem()

type slashAttackSystem struct {
	slashers []*slasher
	slashees []*playerent
	blockers []*shape
}

func (s *slashAttackSystem) addBlocker(b *shape) {
	s.blockers = append(s.blockers, b)
}

func newSlashAttackSystem() slashAttackSystem {
	s := slashAttackSystem{}
	return s
}
func (s *slashAttackSystem) removeSlasher(p *playerent) {
	for i, subslasher := range s.slashers {
		if p == subslasher.ent {
			if i < len(s.slashers)-1 {
				copy(s.slashers[i:], s.slashers[i+1:])
			}
			s.slashers[len(s.slashers)-1] = nil
			s.slashers = s.slashers[:len(s.slashers)-1]

			break
		}
	}

	for i, subslasher := range s.slashers {
		if p == subslasher.ent {
			if i < len(s.slashers)-1 {
				copy(s.slashers[i:], s.slashers[i+1:])
			}
			s.slashers[len(s.slashers)-1] = nil
			s.slashers = s.slashers[:len(s.slashers)-1]
			break
		}
	}

}
func (s *slashAttackSystem) removeSlashee(p *playerent) {
	for i, renderable := range s.slashees {
		if p == renderable {
			if i < len(s.slashees)-1 {
				copy(s.slashees[i:], s.slashees[i+1:])
			}
			s.slashees[len(s.slashees)-1] = nil
			s.slashees = s.slashees[:len(s.slashees)-1]
			break
		}
	}
	// for i, subslasher := range s.slashers {
	// 	if p == subslasher.ent {
	// 		if i < len(s.slashers)-1 {
	// 			copy(s.slashers[i:], s.slashers[i+1:])
	// 		}
	// 		s.slashers[len(s.slashers)-1] = nil
	// 		s.slashers = s.slashers[:len(s.slashers)-1]
	// 		break
	// 	}
	// }
}

func newLinePolar(loc location, length int, angle float64) line {
	xpos := int(float64(length)*math.Cos(angle)) + loc.x
	ypos := int(float64(length)*math.Sin(angle)) + loc.y
	return line{loc, location{xpos, ypos}}
}

func (s *slashAttackSystem) work() {
	// select {
	// case <-s.events:
	toRemove := []*playerent{}
	for _, bot := range s.slashers {

		keepOnPlayer := func() bool {
			midPlayer := bot.ent.rectangle.location
			midPlayer.x += bot.ent.rectangle.dimens.width / 2
			midPlayer.y += bot.ent.rectangle.dimens.height / 2

			for i := 0; i < len(bot.slashLine.lines); i++ {
				rotLine := newLinePolar(midPlayer, 50, bot.animationCount+bot.startangle)
				bot.slashLine.lines[i] = rotLine
			}

			for _, blocker := range s.blockers {
				for _, blockerLine := range blocker.lines {
					for _, bladeLine := range bot.slashLine.lines {
						if _, _, intersected := bladeLine.intersects(blockerLine); intersected {
							return false
						}
					}
				}
			}

			return true
		}
		stopSlashing := func() {
			renderingSystem.removeShape(bot.slashLine)
			// renderingSystem.toRemove <- bot.slashLine
			bot.animationCount = 0
			bot.animating = false
			bot.onCooldown = false
		}

		if bot.animating && bot.animationCount < -3 {
			stopSlashing()
			continue
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
			if ebiten.IsKeyPressed(ebiten.KeyX) && !bot.onCooldown {

				notBlocked := keepOnPlayer()
				if notBlocked {
					renderingSystem.addShape(bot.slashLine)
					bot.animating = true
					bot.onCooldown = true
					// animationTimer := time.NewTimer(310 * time.Millisecond).C
					// go func() {
					// 	select {
					// 	case <-animationTimer:
					// 		if bot.animating {
					// 			stopSlashing()
					// 		}
					// 	}
					// }()
				}
				// else {
				// 	bot.onCooldown = false
				// }

				// bot.onCooldown = true
				// coolDownTimer := time.NewTimer(800 * time.Millisecond).C
				// go func() {
				// 	select {
				// 	case <-coolDownTimer:
				// 		bot.onCooldown = false
				// 	}
				// }()
			}
		} else {
			bot.animationCount -= 0.16
			notBlocked := keepOnPlayer()
			if !notBlocked {
				stopSlashing()
				// bot.onCooldown = false
			}
			if bot.animating {
			foundSlashee:
				for _, slashee := range s.slashees {
					if slashee == bot.ent {
						continue foundSlashee
					}
					for _, slasheeLine := range slashee.rectangle.shape.lines {
						for _, bladeLine := range bot.slashLine.lines {
							if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
								renderingSystem.removeShape(slashee.rectangle.shape)
								// renderingSystem.toRemove <- slashee.rectangle.shape
								collideSystem.removeMover(slashee)
								s.removeSlashee(slashee)
								toRemove = append(toRemove, slashee)
								break foundSlashee
							}
						}
					}
				}
			}

		}

	}

	for _, removeMe := range toRemove {
		for _, slshr := range s.slashers {
			if removeMe == slshr.ent {
				renderingSystem.removeShape(slshr.slashLine)
			}
		}
		s.removeSlasher(removeMe)
	}
	// default:
	// }
}
