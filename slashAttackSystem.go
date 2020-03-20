package main

import (
	"math"
	"time"
)

type slasher struct {
	ent           *acceleratingEnt
	startangle    float64
	onCooldown    bool
	cooldownCount int
}

func newSlasher(p *acceleratingEnt) *slasher {
	s := &slasher{}
	s.ent = p
	s.cooldownCount = 100
	return s
}

var slashSystem = newSlashAttackSystem()

type slashAttackSystem struct {
	slashers []*slasher
}

func (s *slashAttackSystem) addSlasher(b *slasher) {
	s.slashers = append(s.slashers, b)
	b.ent.rect.shape.removals = append(b.ent.rect.shape.removals, func() {
		s.removeSlasher(b.ent.rect.shape)
	})
}

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
		}

		if bot.ent.atkButton && !bot.onCooldown {
			ps := newPivotingShape(bot.ent.rect, bot.startangle)

			renderingSystem.addShape(ps.pivoterShape)
			pivotingSystem.addPivoter(ps)
			bs := playerSprite{bot.ent.rect, swordImage}
			ws := weaponSprite{ps.pivoterShape, &ps.animationCount, bs}
			weaponRenderingSystem.addWeaponSprite(&ws)

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
