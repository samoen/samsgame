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
	slashers map[*entityid]*slasher
}

func (s *slashAttackSystem) addSlasher(id *entityid, b *slasher) {
	s.slashers[id] = b
	id.systems = append(id.systems, abilityActivator)
}

func newSlashAttackSystem() slashAttackSystem {
	s := slashAttackSystem{}
	s.slashers = make(map[*entityid]*slasher)
	return s
}

func (s *slashAttackSystem) work() {
	for slasherid, bot := range s.slashers {
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
			wepid := &entityid{}
			ps := newPivotingShape(slasherid, bot.ent.rect, bot.startangle)
			renderingSystem.addShape(ps.pivoterShape, wepid)
			pivotingSystem.addPivoter(wepid, ps)
			bs := playerSprite{bot.ent.rect, swordImage}
			ws := weaponSprite{ps.pivoterShape, &ps.animationCount, bs}
			addWeaponSprite(&ws, wepid)
			slasherid.associates = append(slasherid.associates, wepid)

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
}
