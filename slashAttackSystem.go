package main

import (
	"math"
	"time"
)

type slasher struct {
	ent       *acceleratingEnt
	animating bool
	// slashLine *shape
	startangle float64
	// animationCount float64
	onCooldown    bool
	cooldownCount int
	// doneAnimating  bool
}

func newSlasher(p *acceleratingEnt) *slasher {
	s := &slasher{}
	s.ent = p
	// s.slashLine = newShape()
	// s.slashLine.lines = []line{line{}}
	s.cooldownCount = 100
	return s
}

var slashSystem = newSlashAttackSystem()

type slashAttackSystem struct {
	slashers []*slasher
	// slashees []*rectangle
	// blockers []*shape
}

func (s *slashAttackSystem) addSlasher(b *slasher) {
	s.slashers = append(s.slashers, b)
	b.ent.rect.shape.removals = append(b.ent.rect.shape.removals, func() {
		s.removeSlasher(b.ent.rect.shape)
		// s.removeSlashee(b.ent.rect.shape)
		// b.slashLine.sysPurge()
	})
}

// func (s *slashAttackSystem) addSlashee(b *rectangle) {
// 	// s.slashees = append(s.slashees, b)
// 	// b.shape.removals = append(b.shape.removals, func() {
// 	// 	// s.removeSlasher(b.ent.rect.shape)
// 	// 	s.removeSlashee(b.shape)

// 	// })
// }

// func (s *slashAttackSystem) addBlocker(b *shape) {
// 	// s.blockers = append(s.blockers, b)
// }

func newSlashAttackSystem() slashAttackSystem {
	s := slashAttackSystem{}
	return s
}

func (s *slashAttackSystem) removeSlasher(p *shape) {
	for i, subslasher := range s.slashers {
		if p == subslasher.ent.rect.shape {
			if i < len(s.slashers)-1 {
				copy(s.slashers[i:], s.slashers[i+1:])
			}
			s.slashers[len(s.slashers)-1] = nil
			s.slashers = s.slashers[:len(s.slashers)-1]
			break
		}
	}
}

func (s *slashAttackSystem) removeSlashee(p *shape) {
	// for i, renderable := range s.slashees {
	// 	if p == renderable.shape {
	// 		if i < len(s.slashees)-1 {
	// 			copy(s.slashees[i:], s.slashees[i+1:])
	// 		}
	// 		s.slashees[len(s.slashees)-1] = nil
	// 		s.slashees = s.slashees[:len(s.slashees)-1]
	// 		break
	// 	}
	// }
}

func newLinePolar(loc location, length int, angle float64) line {
	xpos := int(float64(length)*math.Cos(angle)) + loc.x
	ypos := int(float64(length)*math.Sin(angle)) + loc.y
	return line{loc, location{xpos, ypos}}
}

func (s *shape) sysPurge() {
	for _, rem := range s.removals {
		rem()
	}
}

func (s *slashAttackSystem) work() {
	// select {
	// case <-s.events:
	// toRemove := []*rectangle{}
	for _, bot := range s.slashers {
		bot := bot

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
			// notBlocked := keepOnPlayer()
			// if notBlocked {
			slashLine := newShape()
			slashLine.lines = []line{line{}}
			renderingSystem.addShape(slashLine)
			ps := newPivotingShape(slashLine, bot.ent)
			// ps.startangle = bot.startangle
			ps.animationCount = bot.startangle
			pivotingSystem.addPivoter(ps)

			// }

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

	}

	// for _, removeMe := range toRemove {
	// 	removeMe.shape.sysPurge()
	// }

	// default:
	// }
}

// func removeFromSlice(slice []*interface{}, p *interface{}) []*interface{} {
// 	for i, renderable := range slice {
// 		if p == renderable {
// 			if i < len(slice)-1 {
// 				copy(slice[i:], slice[i+1:])
// 			}
// 			slice[len(slice)-1] = nil
// 			return slice[:len(slice)-1]
// 		}
// 	}
// 	return slice
// }
