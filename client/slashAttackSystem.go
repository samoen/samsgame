package main

import (
	"github.com/hajimehoshi/ebiten"
	"mahgame/gamecore"
	"math"
	"math/rand"
)

type slasher struct {
	bsprit        baseSprite
	wepsprit      baseSprite
	hbarsprit     baseSprite
	collisionId   *bool
	rect          rectangle
	moment        momentum
	agility       float64
	moveSpeed     float64
	directions    directions
	baseloc       location
	endpoint      location
	deth          deathable
	startangle    float64
	cooldownCount int
	swangin       bool
	atkButton     bool
	pivShape      pivotingShape
}

func (s *slasher) defaultStats() {
	cId := false
	s.collisionId = &cId
	s.rect.dimens = dimens{20, 40}
	s.rect.refreshShape(location{50, 50})
	s.agility = 4
	s.moveSpeed = 100
	s.cooldownCount = 0
	s.pivShape.damage = 2
	s.deth.hp = hitpoints{6, 6}
	s.hbarsprit.bOps = &ebiten.DrawImageOptions{}
	s.hbarsprit.sprite = images.empty
	s.bsprit.bOps = &ebiten.DrawImageOptions{}
	s.bsprit.sprite = images.playerStand
	s.wepsprit.bOps = &ebiten.DrawImageOptions{}
	s.wepsprit.sprite = images.sword
}

func (le *slasher) hitbox(s *ebiten.Image) {
	for _, l := range le.rect.shape.lines {
		l.samDrawLine(s)
	}
	if le.swangin {
		for _, l := range le.pivShape.pivoterShape.lines {
			l.samDrawLine(s)
		}
	}
}

type localEnt struct {
	lSlasher       slasher
	hitsToSend     []string
	swangSinceSend bool
	swangAngle     float64
}

func (l *localEnt) toRemoteEnt(pnum string) gamecore.EntityData {
	message := gamecore.EntityData{}
	message.MyPNum = pnum
	message.X = l.lSlasher.rect.location.x
	message.Y = l.lSlasher.rect.location.y
	message.Xaxis = l.lSlasher.moment.Xaxis
	message.Yaxis = l.lSlasher.moment.Yaxis
	message.Up = l.lSlasher.directions.Up
	message.Left = l.lSlasher.directions.Left
	message.Right = l.lSlasher.directions.Right
	message.Down = l.lSlasher.directions.Down
	message.Dmg = l.lSlasher.pivShape.damage
	message.NewSwing = l.swangSinceSend
	message.NewSwingAngle = l.swangAngle
	message.Heading = l.lSlasher.startangle
	message.Swangin = l.lSlasher.swangin
	message.IHit = l.hitsToSend
	message.CurrentHP = l.lSlasher.deth.hp.CurrentHP
	message.MaxHP = l.lSlasher.deth.hp.MaxHP
	l.hitsToSend = nil
	l.swangSinceSend = false
	return message
}

type localPlayer struct {
	locEnt localEnt
}

func (l *localPlayer) checkHitOthers() {
	if myLocalPlayer.locEnt.lSlasher.swangin {
		myLocalPlayer.locEnt.hitremotes()
		for slashee, _ := range localAnimals {
			if myLocalPlayer.locEnt.lSlasher.pivShape.hitConfirm(&slashee.locEnt.lSlasher) {
				slashee.checkRemove()
			}
		}
	}
}

func (l *localPlayer) placePlayer() {
	l.locEnt.lSlasher.rect.refreshShape(location{50, 50})
	l.locEnt.lSlasher.deth.hp = hitpoints{6, 6}
	l.locEnt.lSlasher.spawnSafe()
}

type localAnimal struct {
	locEnt       localEnt
	controlCount int
}

func (la *localAnimal) checkRemove() {
	if la.locEnt.lSlasher.deth.hp.CurrentHP < 1 {
		delete(localAnimals, la)
		la.locEnt.lSlasher.addDeathAnim()
	}
}

func (s *slasher) addDeathAnim() {
	bs0 := baseSprite{}
	bs0.sprite = images.playerfall0
	bs0.bOps = &ebiten.DrawImageOptions{}
	bs0.yaxis = rectCenterPoint(s.rect).y
	bs1 := baseSprite{}
	bs1.sprite = images.playerfall1
	bs1.bOps = &ebiten.DrawImageOptions{}
	bs1.yaxis = rectCenterPoint(s.rect).y
	bs2 := baseSprite{}
	bs2.sprite = images.playerfall2
	bs2.bOps = &ebiten.DrawImageOptions{}
	bs2.yaxis = rectCenterPoint(s.rect).y

	da := &deathAnim{}
	da.sprites = append(da.sprites, bs0)
	da.sprites = append(da.sprites, bs1)
	da.sprites = append(da.sprites, bs2)
	da.rect = s.rect
	da.inverted = math.Abs(s.startangle) > math.Pi/2
	deathAnimations[da] = true
}

func (la *localAnimal) checkHitOthers() {
	if la.locEnt.lSlasher.swangin {
		la.locEnt.hitremotes()
		for slashee, _ := range localAnimals {
			if slashee.locEnt.lSlasher.collisionId == la.locEnt.lSlasher.collisionId {
				continue
			}
			if la.locEnt.lSlasher.pivShape.hitConfirm(&slashee.locEnt.lSlasher) {
				slashee.checkRemove()
			}
		}
		if la.locEnt.lSlasher.pivShape.hitConfirm(&myLocalPlayer.locEnt.lSlasher) {
			if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP < 1 {
				myLocalPlayer.locEnt.lSlasher.addDeathAnim()
			}
		}
	}
}

type remotePlayer struct {
	rSlasher slasher
	servId   string
}

func (bot *remotePlayer) remoteMovement() {
	if receiveCount <= interpTime {
		var newplace location
		if receiveCount == interpTime {
			newplace = bot.rSlasher.endpoint
		} else {
			diffx := (bot.rSlasher.endpoint.x - bot.rSlasher.baseloc.x) / interpTime
			diffy := (bot.rSlasher.endpoint.y - bot.rSlasher.baseloc.y) / interpTime
			newplace = bot.rSlasher.rect.location
			newplace.x += diffx
			newplace.y += diffy
		}
		checkrect := bot.rSlasher.rect
		checkrect.refreshShape(newplace)
		if !checkrect.shape.normalcollides(bot.rSlasher.collisionId) {
			bot.rSlasher.rect.refreshShape(newplace)
		}
	} else {
		deadReckonFrames := interpTime / 2
		if deadReckonFrames < 1 {
			deadReckonFrames = 1
		}
		if receiveCount > interpTime+deadReckonFrames {
			bot.rSlasher.directions.Down = false
			bot.rSlasher.directions.Left = false
			bot.rSlasher.directions.Right = false
			bot.rSlasher.directions.Up = false
		}
		bot.rSlasher.moveCollide()
	}
}

func (s *slasher) startSwing() {
	s.pivShape.bladeLength = 5
	s.cooldownCount = 60
	s.pivShape.alreadyHit = make(map[*bool]bool)
	s.pivShape.animationCount = s.startangle + 2.1
	s.swangin = true
	s.pivShape.startCount = s.pivShape.animationCount
}
func (s *slasher) progressSwing() {
	s.pivShape.animationCount -= axeRotateSpeed
	midPlayer := s.rect.location
	midPlayer.x += s.rect.dimens.width / 2
	midPlayer.y += s.rect.dimens.height / 2
	rotLine := line{}
	rotLine.newLinePolar(midPlayer, s.pivShape.bladeLength, s.pivShape.animationCount)
	crossLine := line{}
	crossLine.newLinePolar(rotLine.p2, s.pivShape.bladeLength/3, s.pivShape.animationCount+math.Pi/2)
	frontCrossLine := line{}
	frontCrossLine.newLinePolar(rotLine.p2, s.pivShape.bladeLength/3, s.pivShape.animationCount-math.Pi/2)
	s.pivShape.pivoterShape.lines = []line{rotLine, crossLine, frontCrossLine}
	
	arcProgress := math.Abs(s.pivShape.startCount - s.pivShape.animationCount)

	if arcProgress > axeArc {
		s.swangin = false
		return
	} else if arcProgress < axeArc*0.3 {
		s.pivShape.bladeLength += 4
	} else if arcProgress > axeArc*0.8 {
		s.pivShape.bladeLength -= 3
	} else {
		s.pivShape.bladeLength = maxAxeLength
	}
}
func (bot *localEnt) handleSwing() {
	if bot.lSlasher.cooldownCount > 0 {
		bot.lSlasher.cooldownCount--
	}
	if bot.lSlasher.atkButton && bot.lSlasher.cooldownCount < 1 {
		bot.lSlasher.startSwing()
		bot.swangSinceSend = true
		bot.swangAngle = bot.lSlasher.startangle
	}
	if bot.lSlasher.swangin {
		bot.lSlasher.progressSwing()
		for blocker, _ := range wepBlockers {
			if blocker.collidesWith(bot.lSlasher.pivShape.pivoterShape) {
				bot.lSlasher.swangin = false
				return
			}
		}
		for _, blocker := range currentTShapes {
			if blocker.collidesWith(bot.lSlasher.pivShape.pivoterShape) {
				bot.lSlasher.swangin = false
				return
			}
		}
	}
}

func (s *slasher) updateAim() {
	if !s.swangin {
		if s.directions.Down ||
			s.directions.Up ||
			s.directions.Right ||
			s.directions.Left {
			hitRange := 1
			moveTipX := 0
			if s.directions.Right {
				moveTipX = hitRange
			} else if s.directions.Left {
				moveTipX = -hitRange
			}
			moveTipY := 0
			if s.directions.Up {
				moveTipY = -hitRange
			} else if s.directions.Down {
				moveTipY = hitRange
			}
			s.startangle = math.Atan2(float64(moveTipY), float64(moveTipX))
		}
	}
}

func (bot *localEnt) hitremotes() {
	for _, slashee := range remotePlayers {
		if bot.lSlasher.pivShape.hitConfirm(&slashee.rSlasher) {
			slashee.rSlasher.deth.skipHpUpdate = 2
			bot.hitsToSend = append(bot.hitsToSend, slashee.servId)
		}
	}
}

func (s *pivotingShape) hitConfirm(slashee *slasher) bool {
	if _, ok := s.alreadyHit[slashee.collisionId]; ok {
		return false
	}
	if slashee.rect.shape.collidesWith(s.pivoterShape) {
		slashee.deth.redScale = 10
		slashee.deth.hp.CurrentHP -= s.damage
		s.alreadyHit[slashee.collisionId] = true
		return true
	}
	return false
}

func (bot *localAnimal) AIControl() {
	bot.controlCount--
	if bot.controlCount < 1 {
		bot.controlCount = rand.Intn(100)
		bot.locEnt.lSlasher.directions = directions{
			rand.Intn(2) == 0,
			rand.Intn(2) == 0,
			rand.Intn(2) == 0,
			rand.Intn(2) == 0,
		}
		bot.locEnt.lSlasher.atkButton = rand.Intn(2) == 0
	}
}

type pivotingShape struct {
	pivoterShape   shape
	animationCount float64
	alreadyHit     map[*bool]bool
	startCount     float64
	bladeLength    int
	damage         int
}

// func rotateAround(center location, point location, angle float64) location {
// 	result := location{}
// 	rotatedX := math.Cos(angle)*float64(point.x-center.x) - math.Sin(angle)*float64(point.y-center.y) + float64(center.x)
// 	rotatedY := math.Sin(angle)*float64(point.x-center.x) + math.Cos(angle)*float64(point.y-center.y) + float64(center.y)
// 	result.x = int(rotatedX)
// 	result.y = int(rotatedY)
// 	return result
// }

type deathable struct {
	redScale     int
	hp           hitpoints
	skipHpUpdate int
}

type hitpoints struct {
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
	myLocalPlayer.placePlayer()
}

type directions struct {
	Right bool
	Down  bool
	Left  bool
	Up    bool
}

func (l *localPlayer) updatePlayerControl() {
	l.locEnt.lSlasher.directions.Right = ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight)
	l.locEnt.lSlasher.directions.Down = ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)
	l.locEnt.lSlasher.directions.Left = ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft)
	l.locEnt.lSlasher.directions.Up = ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)
	l.locEnt.lSlasher.atkButton = ebiten.IsKeyPressed(ebiten.KeyX)
}
