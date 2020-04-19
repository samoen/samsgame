package main

import (
	"math"
)

type slasher struct {
	ent           *acceleratingEnt
	startangle    float64
	cooldownCount int
	swangin       bool
}

func newSlasher(p *acceleratingEnt) *slasher {
	s := &slasher{}
	s.ent = p
	s.cooldownCount = 0
	return s
}

var slashers = make(map[*entityid]*slasher)

func addSlasher(id *entityid, b *slasher) {
	slashers[id] = b
	id.systems = append(id.systems, abilityActivator)
}

func slashersWork() {
	for slasherid, bot := range slashers {
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

		if bot.cooldownCount > 0 {
			bot.cooldownCount--
		}

		if bot.ent.atkButton && bot.cooldownCount < 1 {
			wepid := &entityid{}
			p := &pivotingShape{}
			p.pivotPoint = bot.ent.rect
			p.ownerid = slasherid
			p.animationCount = bot.startangle + 1.2
			p.pivoterShape = newShape()
			p.pivoterShape.lines = makeAxe(p.animationCount, *bot.ent.rect)
			p.alreadyHit = make(map[*entityid]bool)
			p.swangin = &bot.swangin
			bot.swangin = true
			ws := weaponSprite{&p.animationCount, bot, swordImage}

			addHitbox(p.pivoterShape, wepid)
			addPivoter(wepid, p)
			addWeaponSprite(&ws, wepid)

			if d, ok := deathables[slasherid]; ok {
				d.associates = append(d.associates, wepid)
			}

			bot.cooldownCount = 60
		}
	}
}
