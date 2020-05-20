package gamecore

import (
	"github.com/hajimehoshi/ebiten"
	"math"
)

type slasher struct {
	ent           *acceleratingEnt
	startangle    float64
	cooldownCount int
	swangin       bool
	pivShape      *pivotingShape
	remote        bool
	hitsToSend    []*entityid
}

func newSlasher(p *acceleratingEnt) *slasher {
	s := &slasher{}
	s.ent = p
	s.cooldownCount = 0
	s.pivShape = &pivotingShape{}
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
		if !bot.remote {
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
		} else {

		}

		if bot.cooldownCount > 0 {
			bot.cooldownCount--
		}

		if bot.ent.atkButton && bot.cooldownCount < 1 {
			wepid := &entityid{}
			p := &pivotingShape{}
			p.alreadyHit = make(map[*entityid]bool)
			p.wepid = wepid
			p.pivotPoint = bot.ent.rect
			if bot.remote {
				p.animationCount = bot.startangle
			}
			if !bot.remote {
				p.animationCount = bot.startangle + 1.2
			}

			p.pivoterShape = newShape()
			p.pivoterShape.lines = makeAxe(p.animationCount, *bot.ent.rect)
			if !bot.remote {
				for i := 1; i < 20; i++ {
					if !checkBlocker(*p.pivoterShape) {
						break
					} else {
						p.animationCount -= 0.2
						p.pivoterShape.lines = makeAxe(p.animationCount, *p.pivotPoint)
					}
				}
			}

			addHitbox(p.pivoterShape, wepid)
			bot.swangin = true
			ws := weaponSprite{&p.animationCount, bot, swordImage}

			p.startCount = p.animationCount

			bot.pivShape = p

			addWeaponSprite(&ws, wepid)

			slasherid.linked = append(slasherid.linked, wepid)

			bot.cooldownCount = 60
		}
		if bot.swangin {

			bot.pivShape.animationCount -= 0.12
			bot.pivShape.pivoterShape.lines = makeAxe(bot.pivShape.animationCount, *bot.pivShape.pivotPoint)
			blocked := checkBlocker(*bot.pivShape.pivoterShape)

			if ok, slashee, slasheeid := checkSlashee(bot.pivShape, slasherid); ok {
				if !bot.remote {
					if !slashee.remote {
						slashee.gotHit = true
					}
					bot.pivShape.alreadyHit[slasheeid] = true
					bot.hitsToSend = append(bot.hitsToSend, slasheeid)
				}
			}
			if blocked ||
				math.Abs(bot.pivShape.startCount-bot.pivShape.animationCount) > 2 {
				bot.swangin = false
				//bot.pivShape.alreadyHit = make(map[*entityid]bool)
				eliminate(bot.pivShape.wepid)
				continue
			}
		}
	}
}

type deathable struct {
	gotHit         bool
	deathableShape *rectangle
	redScale       int
	hp             Hitpoints
	remote         bool
}

type Hitpoints struct {
	CurrentHP int
	MaxHP     int
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

func respawnsWork(){
	if myDeathable.hp.CurrentHP>0{
		return
	}
	if !ebiten.IsKeyPressed(ebiten.KeyX){
		return
	}
	addLocalPlayer()
}

func deathSystemwork() {
	for dID, mDeathable := range deathables {

		if mDeathable.redScale > 0 {
			mDeathable.redScale--
		}
		if mDeathable.gotHit {
			mDeathable.redScale = 10
			mDeathable.gotHit = false
			mDeathable.hp.CurrentHP--
		}
		if mDeathable.hp.CurrentHP < 1 {
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
		//case pivotingHitbox:
		//	delete(pivoters, id)
		case rotatingSprite:
			delete(weapons, id)
		case playerControlled:
			delete(playerControllables, id)
		case weaponBlocker:
			delete(wepBlockers, id)
		case remoteMover:
			delete(remoteMovers, id)
		}
	}
}
