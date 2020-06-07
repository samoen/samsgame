package gamecore

import (
	"github.com/hajimehoshi/ebiten"
	"math"
)

type slasher struct {
	bsprit         *baseSprite
	wepsprit       *baseSprite
	hbarsprit      *baseSprite
	ent            *acceleratingEnt
	deth           *deathable
	startangle     float64
	cooldownCount  int
	swangin        bool
	swangSinceSend bool
	wepid          *entityid
	pivShape       *pivotingShape
	hitsToSend     []*entityid
}

var slashers = make(map[*entityid]*slasher)

func addSlasher(id *entityid, b *slasher) {
	slashers[id] = b
	id.systems = append(id.systems, abilityActivator)
}

var remotePlayers = make(map[*entityid]*slasher)

func addRemotePlayer(id *entityid, b *slasher) {
	remotePlayers[id] = b
	id.systems = append(id.systems, remotePlayer)
}

func handleSwing(bot *slasher) bool {
	if bot.cooldownCount > 0 {
		bot.cooldownCount--
	}

	if bot.ent.atkButton && bot.cooldownCount < 1 {
		bot.pivShape.bladeLength = 5
		bot.cooldownCount = 60
		bot.pivShape.alreadyHit = make(map[*entityid]bool)

		bot.pivShape.animationCount = bot.startangle + 2.1

		bot.swangin = true
		bot.swangSinceSend = true
		bot.pivShape.startCount = bot.pivShape.animationCount
		//addHitbox(bot.pivShape.pivoterShape, bot.wepid)
	}
	if bot.swangin {
		bot.pivShape.animationCount -= axeRotateSpeed
		bot.pivShape.makeAxe()

		if _, _, blocked := checkBlocker(*bot.pivShape.pivoterShape); blocked {
			bot.swangin = false
			eliminate(bot.wepid)
			return true
		}

		arcProgress := math.Abs(bot.pivShape.startCount - bot.pivShape.animationCount)

		if arcProgress > axeArc {
			bot.swangin = false
			eliminate(bot.wepid)
			return true
		} else if arcProgress < axeArc*0.3 {
			bot.pivShape.bladeLength += 4
		} else if arcProgress > axeArc*0.8 {
			bot.pivShape.bladeLength -= 3
		} else {
			bot.pivShape.bladeLength = maxAxeLength
		}
	}
	return false
}
func remotePlayersWork() {
	rms := interpolating
	if receiveCount > interpTime {
		rms = deadreckoning
	}
	if receiveCount > interpTime+deathreckTime {
		rms = momentumOnly
	}
	for slasherid, bot := range remotePlayers {
		bot := bot
		switch rms {
		case interpolating:
			var newplace location
			if receiveCount == interpTime {
				newplace = bot.ent.endpoint
			} else {
				diffx := (bot.ent.endpoint.x - bot.ent.baseloc.x) / interpTime
				diffy := (bot.ent.endpoint.y - bot.ent.baseloc.y) / interpTime
				newplace = bot.ent.rect.location
				newplace.x += diffx
				newplace.y += diffy
			}
			checkrect := newRectangle(newplace, bot.ent.rect.dimens)
			if !normalcollides(*checkrect.shape, slasherid) {
				bot.ent.rect.refreshShape(newplace)
			}
		case deadreckoning:
			bot.ent.moment = calcMomentum(*bot.ent)
			moveCollide(bot.ent, slasherid)
		case momentumOnly:
			//if receiveCount > pingFrames {
			bot.ent.directions.Down = false
			bot.ent.directions.Left = false
			bot.ent.directions.Right = false
			bot.ent.directions.Up = false
			//}
			bot.ent.moment = calcMomentum(*bot.ent)
			moveCollide(bot.ent, slasherid)
		}
		handleSwing(bot)
	}
}
func slashersWork() {

	for slasherid, bot := range slashers {
		bot := bot

		bot.ent.moment = calcMomentum(*bot.ent)
		moveCollide(bot.ent, slasherid)

		if !bot.swangin {
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
		handleSwing(bot)

		if bot.swangin {
			if ok, slashee, slasheeid := checkSlashee(bot.pivShape, slasherid, remotePlayers); ok {
				//blockx,blocky,blocked := checkBlocker(*bot.pivShape.pivoterShape)
				//if blocked{
				//	mecenter := rectCenterPoint(*bot.ent.rect)
				//	blockdist := math.Abs(float64(mecenter.x - blockx))+math.Abs(float64(mecenter.y - blocky))
				//	slasheecenter := rectCenterPoint(*slashee.ent.rect)
				//	slasheedist := math.Abs(float64(mecenter.x - slasheecenter.x))+math.Abs(float64(mecenter.y - slasheecenter.y))
				//	if blockdist > slasheedist{
				slashee.deth.redScale = 10
				slashee.deth.hp.CurrentHP -= bot.pivShape.damage
				slashee.deth.skipHpUpdate = 2
				bot.pivShape.alreadyHit[slasheeid] = true
				bot.hitsToSend = append(bot.hitsToSend, slasheeid)
				//}
				//}
			}
			if ok, slashee, slasheeid := checkSlashee(bot.pivShape, slasherid, slashers); ok {
				slashee.deth.redScale = 10
				slashee.deth.hp.CurrentHP -= bot.pivShape.damage
				bot.pivShape.alreadyHit[slasheeid] = true
				bot.hitsToSend = append(bot.hitsToSend, slasheeid)

				if slashee.deth.hp.CurrentHP < 1 {
					eliminate(slasheeid)
				}
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

func respawnsWork() {
	if mySlasher.deth.hp.CurrentHP > 0 {
		return
	}
	if !ebiten.IsKeyPressed(ebiten.KeyX) {
		return
	}
	placePlayer()
}

type Directions struct {
	Right bool
	Down  bool
	Left  bool
	Up    bool
}

func updatePlayerControl() {
	mySlasher.ent.directions.Right = ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight)
	mySlasher.ent.directions.Down = ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)
	mySlasher.ent.directions.Left = ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft)
	mySlasher.ent.directions.Up = ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)
	mySlasher.ent.atkButton = ebiten.IsKeyPressed(ebiten.KeyX)
}

func eliminate(id *entityid) {

	for _, sys := range id.systems {
		switch sys {
		case enemyControlled:
			delete(enemyControllers, id)
		case abilityActivator:
			if d, ok := slashers[id]; ok {
				eliminate(d.wepid)
				eliminate(d.deth.hBarid)
			}
			delete(slashers, id)
		case remotePlayer:
			if d, ok := remotePlayers[id]; ok {
				eliminate(d.wepid)
				eliminate(d.deth.hBarid)
			}
			delete(remotePlayers, id)
		case weaponBlocker:
			delete(wepBlockers, id)
		}
	}
}
