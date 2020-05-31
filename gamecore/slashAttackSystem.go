package gamecore

import (
	"github.com/hajimehoshi/ebiten"
	"math"
)

type slasher struct {
	ent            *acceleratingEnt
	startangle     float64
	cooldownCount  int
	swangin        bool
	swangSinceSend bool
	wepid          *entityid
	pivShape       *pivotingShape
	hitsToSend     []*entityid
}

func newSlasher(p *acceleratingEnt) *slasher {
	s := &slasher{}
	s.ent = p
	s.cooldownCount = 0
	s.pivShape = &pivotingShape{}
	s.pivShape.damage = 2
	s.pivShape.pivoterShape = newShape()
	s.pivShape.pivotPoint = s.ent.rect
	s.wepid = &entityid{}
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
		if !slasherid.remote && !bot.swangin {
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
		}

		if bot.cooldownCount > 0 {
			bot.cooldownCount--
		}

		if bot.ent.atkButton && bot.cooldownCount < 1 {
			bot.pivShape.bladeLength = 5
			bot.cooldownCount = 60
			bot.pivShape.alreadyHit = make(map[*entityid]bool)

			//if slasherid.remote {
			//	bot.pivShape.animationCount = bot.startangle
			//}
			//if !slasherid.remote {
				bot.pivShape.animationCount = bot.startangle + 2.1
			//}

			bot.swangin = true
			bot.swangSinceSend = true
			bot.pivShape.startCount = bot.pivShape.animationCount
			bs := &baseSprite{}
			bs.layer = 1
			bs.bOps = &ebiten.DrawImageOptions{}
			bs.sprite = images.sword
			addBasicSprite(bs, bot.wepid)
			addHitbox(bot.pivShape.pivoterShape,bot.wepid)
		}
		//bot.ent.ignoreflip = bot.swangin
		if bot.swangin {

			bot.pivShape.animationCount -= axeRotateSpeed
			bot.pivShape.makeAxe(bot.pivShape.animationCount, *bot.pivShape.pivotPoint)
			blocked := checkBlocker(*bot.pivShape.pivoterShape)
			if !blocked {
				if ok, slashee, slasheeid := checkSlashee(bot.pivShape, slasherid); ok {
					if !slasherid.remote {
						//if !slashee.remote {
						slashee.gotHit = true
						//slashee.hp.CurrentHP--
						slashee.hp.CurrentHP -= bot.pivShape.damage
						slashee.skipHpUpdate = 2
						//}
						bot.pivShape.alreadyHit[slasheeid] = true
						bot.hitsToSend = append(bot.hitsToSend, slasheeid)
					}
				}
			}
			arcProgress := math.Abs(bot.pivShape.startCount - bot.pivShape.animationCount)

			if arcProgress > axeArc {
				bot.swangin = false
				eliminate(bot.wepid)
			} else if arcProgress < axeArc * 0.3 {
				bot.pivShape.bladeLength += 4
			} else if arcProgress > axeArc*0.8 {
				bot.pivShape.bladeLength -= 3
			}else{
				bot.pivShape.bladeLength = maxAxeLength
			}
		}

	}
}

const (
	maxAxeLength   = 45
	axeRotateSpeed = 0.12
	axeArc         = 3.9
)

type deathable struct {
	gotHit         bool
	deathableShape *rectangle
	redScale       int
	hp             Hitpoints
	skipHpUpdate   int
	hBarid         *entityid
}

type Hitpoints struct {
	CurrentHP int
	MaxHP     int
}

var deathables = make(map[*entityid]*deathable)

func addDeathable(id *entityid, d *deathable) {
	id.systems = append(id.systems, hurtable)
	deathables[id] = d
}

func respawnsWork() {
	if myDeathable.hp.CurrentHP > 0 {
		return
	}
	if !ebiten.IsKeyPressed(ebiten.KeyX) {
		return
	}
	addPlayerEntity(&entityid{}, location{50, 50}, Hitpoints{6, 6}, true)
}

func deathSystemwork() {
	for _, mDeathable := range deathables {
		if mDeathable.redScale > 0 {
			mDeathable.redScale--
		}
		if mDeathable.gotHit {
			mDeathable.redScale = 10
			mDeathable.gotHit = false
			//mDeathable.hp.CurrentHP--
		}

	}

	for dID, mDeathable := range deathables {
		if mDeathable.hp.CurrentHP < 1 && !dID.remote {
			eliminate(dID)
		}

	}
}

func eliminate(id *entityid) {

	for _, sys := range id.systems {
		switch sys {
		case spriteRenderable:
			delete(basicSprites, id)
		case hitBoxRenderable:
			delete(hitBoxes, id)
		case moveCollider:
			delete(movers, id)
		case solidCollider:
			delete(solids, id)
		case enemyControlled:
			delete(enemyControllers, id)
		case abilityActivator:
			if d, ok := slashers[id]; ok {
				eliminate(d.wepid)
			}
			delete(slashers, id)
		case hurtable:
			if d, ok := deathables[id]; ok {
				eliminate(d.hBarid)
			}
			delete(deathables, id)
		case playerControlled:
			delete(playerControllables, id)
		case weaponBlocker:
			delete(wepBlockers, id)
		}
	}
}
