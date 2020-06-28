package gamecore

import (
	"github.com/hajimehoshi/ebiten"
	"math"
	"math/rand"
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
	atkButton   bool
	pivShape       *pivotingShape
}

func (bot *slasher)hitPlayer(){
	if _, ok := bot.pivShape.alreadyHit[myLocalPlayer.locEnt.lSlasher.ent.rect.shape]; ok {
		return
	}
	if myLocalPlayer.locEnt.lSlasher.ent.rect.shape.collidesWith(*bot.pivShape.pivoterShape) {
		myLocalPlayer.locEnt.lSlasher.getClapped(bot)
		if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP < 1 {
			myLocalPlayer.dead = true
		}
	}
}

func (le *localEnt)hitbox(s *ebiten.Image){
	for _, l := range le.lSlasher.ent.rect.shape.lines {
		l.samDrawLine(s)
	}
	if le.lSlasher.swangin {
		for _, l := range le.lSlasher.pivShape.pivoterShape.lines {
			l.samDrawLine(s)
		}
	}
}

type localEnt struct {
	lSlasher 		*slasher
	hitsToSend     []string
}

type localPlayer struct{
	locEnt localEnt
	dead bool
}

type localAnimal struct{
	locEnt localEnt
	controlCount int
}

type remotePlayer struct {
	rSlasher 		*slasher
	servId         string
}

var slashers = make(map[*entityid]*localAnimal)

var remotePlayers = make(map[string]*remotePlayer)

func (bot *slasher)handleSwing() {
	if bot.cooldownCount > 0 {
		bot.cooldownCount--
	}

	if bot.atkButton && bot.cooldownCount < 1 {
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
			bot.rSlasher.ent.moveCollide()
		case momentumOnly:
			//if receiveCount > pingFrames {
			bot.rSlasher.ent.directions.Down = false
			bot.rSlasher.ent.directions.Left = false
			bot.rSlasher.ent.directions.Right = false
			bot.rSlasher.ent.directions.Up = false
			//}
			bot.rSlasher.ent.moveCollide()
		}
		bot.rSlasher.handleSwing()
	}
}

func (s *slasher) updateAim(){
	if !s.swangin {
		if s.ent.directions.Down ||
			s.ent.directions.Up ||
			s.ent.directions.Right ||
			s.ent.directions.Left {
			hitRange := 1
			moveTipX := 0
			if s.ent.directions.Right {
				moveTipX = hitRange
			} else if s.ent.directions.Left {
				moveTipX = -hitRange
			}
			moveTipY := 0
			if s.ent.directions.Up {
				moveTipY = -hitRange
			} else if s.ent.directions.Down {
				moveTipY = hitRange
			}
			s.startangle = math.Atan2(float64(moveTipY), float64(moveTipX))
		}
	}
}

func (bot *localEnt) hitremotes(){
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
}

func (bot *localEnt) HitAnimals(){
	for slasheeid, slashee := range slashers {
		if slashee.locEnt.lSlasher.ent == bot.lSlasher.ent {
			return
		}
		if _, ok := bot.lSlasher.pivShape.alreadyHit[slashee.locEnt.lSlasher.ent.rect.shape]; ok {
			return
		}
		if slashee.locEnt.lSlasher.ent.rect.shape.collidesWith(*bot.lSlasher.pivShape.pivoterShape) {
			slashee.locEnt.lSlasher.getClapped(bot.lSlasher)
			if slashee.locEnt.lSlasher.deth.hp.CurrentHP < 1 {
				delete(slashers, slasheeid)
			}
		}
	}
}

func (slashee *slasher)getClapped(bot *slasher){
	slashee.deth.redScale = 10
	slashee.deth.hp.CurrentHP -= bot.pivShape.damage
	bot.pivShape.alreadyHit[slashee.ent.rect.shape] = true
}

func (bot *localAnimal) AIControl(){
	bot.controlCount--
	if bot.controlCount < 1 {
		bot.controlCount = rand.Intn(100)
		bot.locEnt.lSlasher.ent.directions = Directions{
			rand.Intn(2) == 0,
			rand.Intn(2) == 0,
			rand.Intn(2) == 0,
			rand.Intn(2) == 0,
		}
		bot.locEnt.lSlasher.atkButton = rand.Intn(2) == 0
	}
}

func animalsWork() {
	for _, bot := range slashers {
		bot.AIControl()
		bot.locEnt.lSlasher.ent.moveCollide()
		bot.locEnt.lSlasher.updateAim()
		bot.locEnt.lSlasher.handleSwing()
		if bot.locEnt.lSlasher.swangin {
			bot.locEnt.hitremotes()
			bot.locEnt.HitAnimals()
			bot.locEnt.lSlasher.hitPlayer()
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
	if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP > 0 {
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
	myLocalPlayer.locEnt.lSlasher.ent.directions.Right = ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight)
	myLocalPlayer.locEnt.lSlasher.ent.directions.Down = ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)
	myLocalPlayer.locEnt.lSlasher.ent.directions.Left = ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft)
	myLocalPlayer.locEnt.lSlasher.ent.directions.Up = ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)
	myLocalPlayer.locEnt.lSlasher.atkButton = ebiten.IsKeyPressed(ebiten.KeyX)
}
