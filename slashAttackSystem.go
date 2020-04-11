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

// var slashSystem = newSlashAttackSystem()

// type slashAttackSystem struct {

// }
var slashers = make(map[*entityid]*slasher)

// func (s *slashAttackSystem) addSlasher(id *entityid, b *slasher) {
// 	s.slashers[id] = b
// 	id.systems = append(id.systems, abilityActivator)
// }

// func newSlashAttackSystem() slashAttackSystem {
// 	s := slashAttackSystem{}
// 	s.slashers = make(map[*entityid]*slasher)
// 	return s
// }

func (bot *slasher) work(parent *entityid) {
	// for slasherid, bot := range s.slashers {
	// bot := bot
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
		ps := newPivotingShape(parent, bot.ent.rect, bot.startangle)
		wepid.realSystems[pivotingHitbox] = ps

		ndeathable := deathable{}
		ndeathable.deathableShape = ps.pivoterShape
		if pdeathable, ok := parent.realSystems[hurtable]; ok {
			if pd, ok := pdeathable.(deathable); ok {
				pd.associates = append(pd.associates, wepid)
			}
		}
		// renderingSystem.addShape(ps.pivoterShape, wepid)
		// pivotingSystem.addPivoter(wepid, ps)
		bs := playerSprite{bot.ent.rect, swordImage}
		ws := weaponSprite{ps.pivoterShape, &ps.animationCount, bs}

		addWeaponSprite(&ws, wepid)

		allEnts = append(allEnts, wepid)

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
	// }
}
