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
	pivShape       *pivotingShape
	hitsToSend     []string
	servId         string
}

var slashers = make(map[*entityid]*slasher)

var remotePlayers = make(map[string]*slasher)

func handleSwing(bot *slasher) {
	if bot.cooldownCount > 0 {
		bot.cooldownCount--
	}

	if bot.ent.atkButton && bot.cooldownCount < 1 {
		bot.pivShape.bladeLength = 5
		bot.cooldownCount = 60
		bot.pivShape.alreadyHit = make(map[*shape]bool)
		bot.pivShape.animationCount = bot.startangle + 2.1
		bot.swangin = true
		bot.swangSinceSend = true
		bot.pivShape.startCount = bot.pivShape.animationCount
	}
	if bot.swangin {
		bot.pivShape.animationCount -= axeRotateSpeed
		bot.pivShape.makeAxe()

		for _, blocker := range wepBlockers {
			if blocker.collidesWith(*bot.pivShape.pivoterShape) {
				bot.swangin = false
				return
			}
		}

		arcProgress := math.Abs(bot.pivShape.startCount - bot.pivShape.animationCount)

		if arcProgress > axeArc {
			bot.swangin = false
			return
		} else if arcProgress < axeArc*0.3 {
			bot.pivShape.bladeLength += 4
		} else if arcProgress > axeArc*0.8 {
			bot.pivShape.bladeLength -= 3
		} else {
			bot.pivShape.bladeLength = maxAxeLength
		}
	}
}

func remotePlayersWork() {
	rms := interpolating
	if receiveCount > interpTime {
		rms = deadreckoning
	}
	if receiveCount > interpTime+deathreckTime {
		rms = momentumOnly
	}
	for _, bot := range remotePlayers {
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
			if !normalcollides(*checkrect.shape, bot.ent.rect.shape) {
				bot.ent.rect.refreshShape(newplace)
			}
		case deadreckoning:
			bot.ent.moment = calcMomentum(*bot.ent)
			moveCollide(bot.ent)
		case momentumOnly:
			//if receiveCount > pingFrames {
			bot.ent.directions.Down = false
			bot.ent.directions.Left = false
			bot.ent.directions.Right = false
			bot.ent.directions.Up = false
			//}
			bot.ent.moment = calcMomentum(*bot.ent)
			moveCollide(bot.ent)
		}
		handleSwing(bot)
	}
}
func slashersWork() {
	for slasherid, bot := range slashers {
		bot.ent.moment = calcMomentum(*bot.ent)
		moveCollide(bot.ent)

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
			for _, slashee := range remotePlayers {
				if _, ok := bot.pivShape.alreadyHit[slashee.ent.rect.shape]; ok {
					continue
				}
				if slashee.ent.rect.shape.collidesWith(*bot.pivShape.pivoterShape) {
					slashee.deth.redScale = 10
					slashee.deth.hp.CurrentHP -= bot.pivShape.damage
					slashee.deth.skipHpUpdate = 2
					bot.pivShape.alreadyHit[slashee.ent.rect.shape] = true
					bot.hitsToSend = append(bot.hitsToSend, slashee.servId)
				}
			}

			for slasheeid, slashee := range slashers {
				if slasheeid == slasherid {
					continue
				}
				if _, ok := bot.pivShape.alreadyHit[slashee.ent.rect.shape]; ok {
					continue
				}
				if slashee.ent.rect.shape.collidesWith(*bot.pivShape.pivoterShape) {
					slashee.deth.redScale = 10
					slashee.deth.hp.CurrentHP -= bot.pivShape.damage
					bot.pivShape.alreadyHit[slashee.ent.rect.shape] = true
					bot.hitsToSend = append(bot.hitsToSend, slashee.servId)

					if slashee.deth.hp.CurrentHP < 1 {
						delete(slashers, slasheeid)
						delete(enemyControllers, slasheeid)
					}
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
