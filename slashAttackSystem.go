package main

import (
	"math"
	"time"
)

type slasher struct {
	ent            *acceleratingEnt
	animating      bool
	slashLine      *shape
	startangle     float64
	animationCount float64
	onCooldown     bool
	cooldownCount  int
	doneAnimating  bool
}

func newSlasher(p *acceleratingEnt) *slasher {
	s := &slasher{}
	s.ent = p
	s.slashLine = &shape{[]line{line{}}}
	s.cooldownCount = 100
	return s
}

var slashSystem = newSlashAttackSystem()

type slashAttackSystem struct {
	slashers []*slasher
	slashees []*rectangle
	blockers []*shape
}

func (s *slashAttackSystem) addBlocker(b *shape) {
	s.blockers = append(s.blockers, b)
}

func newSlashAttackSystem() slashAttackSystem {
	s := slashAttackSystem{}
	return s
}

// func (s *slashAttackSystem) purge(p *rectangle) {
// 	s.removeSlashee(p.ent)
// 	// slashSystem.slashees = removeFromSlice(slashSystem.slashees, p.ent)
// 	// s.removeSlashee(p)
// 	s.removeSlasher(p)
// }

func (s *slashAttackSystem) removeSlasher(p *rectangle) {
	for _, slshr := range s.slashers {
		if p == slshr.ent.rect {
			renderingSystem.removeShape(slshr.slashLine)
		}
	}
	for i, subslasher := range s.slashers {
		if p == subslasher.ent.rect {
			if i < len(s.slashers)-1 {
				copy(s.slashers[i:], s.slashers[i+1:])
			}
			s.slashers[len(s.slashers)-1] = nil
			s.slashers = s.slashers[:len(s.slashers)-1]
			break
		}
	}
}
func (s *slashAttackSystem) removeSlashee(p *rectangle) {
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
}

func newLinePolar(loc location, length int, angle float64) line {
	xpos := int(float64(length)*math.Cos(angle)) + loc.x
	ypos := int(float64(length)*math.Sin(angle)) + loc.y
	return line{loc, location{xpos, ypos}}
}

func (s *slashAttackSystem) work() {
	// select {
	// case <-s.events:
	toRemove := []*rectangle{}
	for _, bot := range s.slashers {
		bot := bot

		keepOnPlayer := func() bool {
			midPlayer := bot.ent.rect.location
			midPlayer.x += bot.ent.rect.dimens.width / 2
			midPlayer.y += bot.ent.rect.dimens.height / 2

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
			bot.animationCount = 0
			bot.animating = false
		}

		if bot.doneAnimating {
			bot.doneAnimating = false
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
			if bot.ent.atkButton && !bot.onCooldown {
				notBlocked := keepOnPlayer()
				if notBlocked {
					renderingSystem.addShape(bot.slashLine)
					bot.animating = true
					animChan := time.NewTimer(310 * time.Millisecond).C
					go func() {
						select {
						case <-animChan:
							bot.doneAnimating = true
						}
					}()
				}

				bot.onCooldown = true
				bot.cooldownCount = 0
				coolDownTimer := time.NewTimer(800 * time.Millisecond).C
				go func() {
					select {
					case <-coolDownTimer:
						bot.onCooldown = false
					}
				}()
			}
		} else {
			bot.animationCount -= 0.16
			notBlocked := keepOnPlayer()
			if !notBlocked {
				stopSlashing()
			}
			if bot.animating {
			foundSlashee:
				for _, slashee := range s.slashees {
					if slashee == bot.ent.rect {
						continue foundSlashee
					}
					for _, slasheeLine := range slashee.shape.lines {
						for _, bladeLine := range bot.slashLine.lines {
							if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
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
		removeFromAllSystems(removeMe)
	}
	// default:
	// }
}

func removeFromAllSystems(removeMe *rectangle) {
	renderingSystem.removeShape(removeMe.shape)
	collideSystem.removeMover(removeMe)
	botsMoveSystem.removeEnemyMover(removeMe)
	// botsMoveSystem.movers = removeFromSlice(botsMoveSystem.movers, removeMe)
	// slashSystem.purge(removeMe)
	slashSystem.removeSlasher(removeMe)
	slashSystem.removeSlashee(removeMe)
}

func removeFromSlice(slice []*acceleratingEnt, p *acceleratingEnt) []*acceleratingEnt {
	for i, renderable := range slice {
		if p == renderable {
			if i < len(slice)-1 {
				copy(slice[i:], slice[i+1:])
			}
			slice[len(slice)-1] = nil
			return slice[:len(slice)-1]
		}
	}
	return slice
}
