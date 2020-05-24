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
	s.pivShape.bladeLength = 45
	s.wepid = &entityid{}
	return s
}

var slashers = make(map[*entityid]*slasher)

func addSlasher(id *entityid, b *slasher) {
	slashers[id] = b
	id.systems = append(id.systems, abilityActivator)
}

func slashersWork() {
	center := renderOffset()
	for slasherid, bot := range slashers {
		bot := bot
		if !slasherid.remote {
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
			bot.pivShape.bladeLength = 10
			bot.pivShape.alreadyHit = make(map[*entityid]bool)
			bot.pivShape.pivotPoint = bot.ent.rect
			if slasherid.remote {
				bot.pivShape.animationCount = bot.startangle
			}
			if !slasherid.remote {
				bot.pivShape.animationCount = bot.startangle + 1.2
			}

			bot.pivShape.pivoterShape = newShape()
			addHitbox(bot.pivShape.pivoterShape, bot.wepid)
			bot.swangin = true
			bot.swangSinceSend = true
			bot.pivShape.startCount = bot.pivShape.animationCount
			//bot.pivShape = p
			bs := &baseSprite{}
			bs.bOps = &ebiten.DrawImageOptions{}
			bs.sprite = swordImage
			addBasicSprite(bs, bot.wepid)
			bot.cooldownCount = 60
		}
		bot.ent.ignoreflip = bot.swangin
		if bot.swangin {
			if bot.pivShape.bladeLength < 45 {
				bot.pivShape.bladeLength += 5
			}
			bot.pivShape.animationCount -= 0.12
			bot.pivShape.makeAxe(bot.pivShape.animationCount, *bot.pivShape.pivotPoint)
			blocked := checkBlocker(*bot.pivShape.pivoterShape)
			if !blocked{
				if ok, slashee, slasheeid := checkSlashee(bot.pivShape, slasherid); ok {
					if !slasherid.remote {
						//if !slashee.remote {
						slashee.gotHit = true
						slashee.hp.CurrentHP--
						slashee.skipHpUpdate = 2
						//}
						bot.pivShape.alreadyHit[slasheeid] = true
						bot.hitsToSend = append(bot.hitsToSend, slasheeid)
					}
				}
			}

			if math.Abs(bot.pivShape.startCount-bot.pivShape.animationCount) > 2 {
				bot.swangin = false
				eliminate(bot.wepid)
			}
		}
		if bs, ok := basicSprites[bot.wepid]; ok {
			_, imH := bs.sprite.Size()
			ownerCenter := rectCenterPoint(*bot.ent.rect)
			cameraShift(ownerCenter, center, bs.bOps)
			addOp := ebiten.GeoM{}
			hRatio := float64(bot.pivShape.bladeLength+bot.pivShape.bladeLength/4) / float64(imH)
			addOp.Scale(hRatio, hRatio)
			addOp.Translate(-float64(bot.ent.rect.dimens.width)/2, 0)
			addOp.Rotate(bot.pivShape.animationCount - (math.Pi / 2))
			bs.bOps.GeoM.Add(addOp)
		}
	}
}

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
	addLocalPlayer()
}

func deathSystemwork() {
	center := renderOffset()
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
		if bs, ok := basicSprites[dID]; ok {
			bs.bOps.ColorM.Translate(float64(mDeathable.redScale), 0, 0, 0)
		}
		if bs, ok := basicSprites[mDeathable.hBarid]; ok {
			healthbarlocation := location{mDeathable.deathableShape.location.x, mDeathable.deathableShape.location.y - 10}
			healthbardimenswidth := mDeathable.hp.CurrentHP * mDeathable.deathableShape.dimens.width / mDeathable.hp.MaxHP
			scaleToDimension(dimens{healthbardimenswidth, 5}, emptyImage, bs.bOps)
			cameraShift(healthbarlocation, center, bs.bOps)
		}

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
