package gamecore

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
		if bot.ent.directions.Down ||
			bot.ent.directions.Up ||
			bot.ent.directions.Right ||
			bot.ent.directions.Left {
			hitRange := 1
			moveTipX := 0
			if bot.ent.directions.Right {
				moveTipX = hitRange
			} else if bot.ent.directions.Left {
				moveTipX = -hitRange
			}
			moveTipY := 0
			if bot.ent.directions.Up {
				moveTipY = -hitRange
			} else if bot.ent.directions.Down {
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

			slasherid.linked = append(slasherid.linked, wepid)

			bot.cooldownCount = 60
		}
	}
}
type deathable struct {
	gotHit         bool
	deathableShape *rectangle
	redScale       int
	currentHP      int
	maxHP          int
}

var deathables = make(map[*entityid]*deathable)

func addDeathable(id *entityid, d *deathable) {
	id.systems = append(id.systems, hurtable)
	deathables[id] = d

	hBarEnt := &entityid{}
	healthBarSprite := &healthBarSprite{}
	healthBarSprite.ownerDeathable = d
	id.linked = append(id.linked, hBarEnt)
	addHealthBarSprite(healthBarSprite, hBarEnt)
}

func deathSystemwork() {
	for dID, mDeathable := range deathables {

		if mDeathable.redScale > 0 {
			mDeathable.redScale--
		}
		if mDeathable.gotHit {
			mDeathable.redScale = 10
			mDeathable.gotHit = false
			mDeathable.currentHP--
		}
		if mDeathable.currentHP < 1 {
			eliminate(dID)
		}
	}
}

func eliminate(id *entityid) {

	for _, asc := range id.linked {
		eliminate(asc)
	}

	for _, sys := range id.systems {
		switch sys {
		case spriteRenderable:
			delete(basicSprites, id)
		case healthBarRenderable:
			delete(healthbars, id)
		case hitBoxRenderable:
			delete(hitBoxes, id)
		case moveCollider:
			delete(movers, id)
		case solidCollider:
			delete(solids, id)
		case enemyControlled:
			delete(enemyControllers, id)
		case abilityActivator:
			delete(slashers, id)
		case hurtable:
			delete(deathables, id)
		case pivotingHitbox:
			delete(pivoters, id)
		case rotatingSprite:
			delete(weapons, id)
		case playerControlled:
			delete(playerControllables, id)
		case weaponBlocker:
			delete(wepBlockers, id)
		}
	}
}