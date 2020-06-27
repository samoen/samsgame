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
}

type localEnt struct {
	lSlasher 		*slasher
	hitsToSend     []string
}

type remotePlayer struct {
	rSlasher 		*slasher
	servId         string
}

var slashers = make(map[*entityid]*localEnt)

var remotePlayers = make(map[string]*remotePlayer)

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
		for _, blocker := range currentTShapes {
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
				newplace = bot.rSlasher.ent.endpoint
			} else {
				diffx := (bot.rSlasher.ent.endpoint.x - bot.rSlasher.ent.baseloc.x) / interpTime
				diffy := (bot.rSlasher.ent.endpoint.y - bot.rSlasher.ent.baseloc.y) / interpTime
				newplace = bot.rSlasher.ent.rect.location
				newplace.x += diffx
				newplace.y += diffy
			}
			checkrect := newRectangle(newplace, bot.rSlasher.ent.rect.dimens)
			if !normalcollides(*checkrect.shape, bot.rSlasher.ent.rect.shape) {
				bot.rSlasher.ent.rect.refreshShape(newplace)
			}
		case deadreckoning:
			bot.rSlasher.ent.moment = calcMomentum(*bot.rSlasher.ent)
			moveCollide(bot.rSlasher.ent)
		case momentumOnly:
			//if receiveCount > pingFrames {
			bot.rSlasher.ent.directions.Down = false
			bot.rSlasher.ent.directions.Left = false
			bot.rSlasher.ent.directions.Right = false
			bot.rSlasher.ent.directions.Up = false
			//}
			bot.rSlasher.ent.moment = calcMomentum(*bot.rSlasher.ent)
			moveCollide(bot.rSlasher.ent)
		}
		handleSwing(bot.rSlasher)
	}
}
func slashersWork() {
	for _, bot := range slashers {
		bot.lSlasher.ent.moment = calcMomentum(*bot.lSlasher.ent)
		moveCollide(bot.lSlasher.ent)

		if !bot.lSlasher.swangin {
			if bot.lSlasher.ent.directions.Down ||
				bot.lSlasher.ent.directions.Up ||
				bot.lSlasher.ent.directions.Right ||
				bot.lSlasher.ent.directions.Left {
				hitRange := 1
				moveTipX := 0
				if bot.lSlasher.ent.directions.Right {
					moveTipX = hitRange
				} else if bot.lSlasher.ent.directions.Left {
					moveTipX = -hitRange
				}
				moveTipY := 0
				if bot.lSlasher.ent.directions.Up {
					moveTipY = -hitRange
				} else if bot.lSlasher.ent.directions.Down {
					moveTipY = hitRange
				}
				bot.lSlasher.startangle = math.Atan2(float64(moveTipY), float64(moveTipX))
			}
		}
		handleSwing(bot.lSlasher)

		if bot.lSlasher.swangin {
			for _, slashee := range remotePlayers {
				if _, ok := bot.lSlasher.pivShape.alreadyHit[slashee.rSlasher.ent.rect.shape]; ok {
					continue
				}
				if slashee.rSlasher.ent.rect.shape.collidesWith(*bot.lSlasher.pivShape.pivoterShape) {
					slashee.rSlasher.deth.redScale = 10
					slashee.rSlasher.deth.hp.CurrentHP -= bot.lSlasher.pivShape.damage
					slashee.rSlasher.deth.skipHpUpdate = 2
					bot.lSlasher.pivShape.alreadyHit[slashee.rSlasher.ent.rect.shape] = true
					bot.hitsToSend = append(bot.hitsToSend, slashee.servId)
				}
			}

			for slasheeid, slashee := range slashers {
				if slashee.lSlasher.ent == bot.lSlasher.ent {
					continue
				}
				if _, ok := bot.lSlasher.pivShape.alreadyHit[slashee.lSlasher.ent.rect.shape]; ok {
					continue
				}
				if slashee.lSlasher.ent.rect.shape.collidesWith(*bot.lSlasher.pivShape.pivoterShape) {
					slashee.lSlasher.deth.redScale = 10
					slashee.lSlasher.deth.hp.CurrentHP -= bot.lSlasher.pivShape.damage
					bot.lSlasher.pivShape.alreadyHit[slashee.lSlasher.ent.rect.shape] = true
					//bot.lSlasher.hitsToSend = append(bot.lSlasher.hitsToSend, slashee.lSlasher.servId)

					if slashee.lSlasher.deth.hp.CurrentHP < 1 {
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
	if mySlasher.lSlasher.deth.hp.CurrentHP > 0 {
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
	mySlasher.lSlasher.ent.directions.Right = ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight)
	mySlasher.lSlasher.ent.directions.Down = ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)
	mySlasher.lSlasher.ent.directions.Left = ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft)
	mySlasher.lSlasher.ent.directions.Up = ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)
	mySlasher.lSlasher.ent.atkButton = ebiten.IsKeyPressed(ebiten.KeyX)
}
