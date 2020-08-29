package gamecore

import (
	"github.com/hajimehoshi/ebiten"
	"math"
	"math/rand"
)

type slasher struct {
	bsprit         baseSprite
	wepsprit       baseSprite
	hbarsprit      baseSprite
	ent            acceleratingEnt
	deth           deathable
	startangle     float64
	cooldownCount  int
	swangin        bool
	swangSinceSend bool
	atkButton      bool
	pivShape       pivotingShape
}

func (s *slasher) defaultStats() {
	cId := false
	s.ent.collisionId = &cId
	s.ent.rect.dimens = dimens{20, 40}
	s.ent.rect.refreshShape(location{50, 50})
	s.ent.agility = 4
	s.ent.moveSpeed = 100
	s.cooldownCount = 0
	s.pivShape.damage = 2
	s.deth.hp = Hitpoints{6, 6}
	s.hbarsprit.bOps = &ebiten.DrawImageOptions{}
	s.hbarsprit.sprite = images.empty
	s.bsprit.bOps = &ebiten.DrawImageOptions{}
	s.bsprit.sprite = images.playerStand
	s.wepsprit.bOps = &ebiten.DrawImageOptions{}
	s.wepsprit.sprite = images.sword
}

func (le *localEnt) hitbox(s *ebiten.Image) {
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
	lSlasher   slasher
	hitsToSend []string
}

type localPlayer struct {
	locEnt localEnt
}

func (l *localPlayer)checkHitOthers(){
	if myLocalPlayer.locEnt.lSlasher.swangin {
		myLocalPlayer.locEnt.hitremotes()
		for slashee, _ := range slashers {
			if myLocalPlayer.locEnt.lSlasher.pivShape.checkHitAnimal(&slashee.locEnt.lSlasher){
				slashee.checkRemove()
			}
		}
	}
}

func (l *localPlayer) placePlayer() {
	l.locEnt.lSlasher.ent.rect.refreshShape(location{50, 50})
	l.locEnt.lSlasher.deth.hp = Hitpoints{6, 6}
	l.locEnt.lSlasher.ent.spawnSafe()
}

type localAnimal struct {
	locEnt       localEnt
	controlCount int
}

func (la *localAnimal) checkRemove(){
	if la.locEnt.lSlasher.deth.hp.CurrentHP < 1 {
		delete(slashers,la)
		la.locEnt.lSlasher.addDeathAnim()
	}
}

func (s *slasher)addDeathAnim(){
	bs0 := baseSprite{}
	bs0.sprite = images.playerfall0
	bs0.bOps = &ebiten.DrawImageOptions{}
	bs0.yaxis = rectCenterPoint(s.ent.rect).y
	bs1 := baseSprite{}
	bs1.sprite = images.playerfall1
	bs1.bOps = &ebiten.DrawImageOptions{}
	bs1.yaxis = rectCenterPoint(s.ent.rect).y
	bs2 := baseSprite{}
	bs2.sprite = images.playerfall2
	bs2.bOps = &ebiten.DrawImageOptions{}
	bs2.yaxis = rectCenterPoint(s.ent.rect).y


	da := &deathAnim{}
	da.sprites = append(da.sprites, bs0)
	da.sprites = append(da.sprites, bs1)
	da.sprites = append(da.sprites, bs2)
	da.rect = s.ent.rect
	da.inverted = math.Abs(s.startangle) > math.Pi/2
	deathAnimations[da]=true
}

func (la *localAnimal)checkHitOthers(){
	if la.locEnt.lSlasher.swangin {
		la.locEnt.hitremotes()
		for slashee, _ := range slashers {
			if slashee.locEnt.lSlasher.ent.collisionId == la.locEnt.lSlasher.ent.collisionId {
				continue
			}
			if la.locEnt.lSlasher.pivShape.checkHitAnimal(&slashee.locEnt.lSlasher){
				slashee.checkRemove()
			}
		}
		if la.locEnt.lSlasher.pivShape.checkHitAnimal(&myLocalPlayer.locEnt.lSlasher){
			if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP<1{
				myLocalPlayer.locEnt.lSlasher.addDeathAnim()
			}
		}
	}
}

type remotePlayer struct {
	rSlasher slasher
	servId   string
}

func (bot *remotePlayer)remoteMovement(){
	if receiveCount < interpTime{
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
		checkrect := bot.rSlasher.ent.rect
		checkrect.refreshShape(newplace)
		if !checkrect.shape.normalcollides(bot.rSlasher.ent.collisionId) {
			bot.rSlasher.ent.rect.refreshShape(newplace)
		}
	}else if receiveCount > interpTime+deathreckTime{
		//if receiveCount > pingFrames {
		bot.rSlasher.ent.directions.Down = false
		bot.rSlasher.ent.directions.Left = false
		bot.rSlasher.ent.directions.Right = false
		bot.rSlasher.ent.directions.Up = false
		//}
		bot.rSlasher.ent.moveCollide()
	} else{
		bot.rSlasher.ent.moveCollide()
	}
}

func (bot *slasher) handleSwing() {
	if bot.cooldownCount > 0 {
		bot.cooldownCount--
	}

	if bot.atkButton && bot.cooldownCount < 1 {
		bot.pivShape.bladeLength = 5
		bot.cooldownCount = 60
		bot.pivShape.alreadyHit = make(map[*bool]bool)
		bot.pivShape.animationCount = bot.startangle + 2.1
		bot.swangin = true
		bot.swangSinceSend = true
		bot.pivShape.startCount = bot.pivShape.animationCount
	}
	if bot.swangin {
		bot.pivShape.animationCount -= axeRotateSpeed
		midPlayer := bot.ent.rect.location
		midPlayer.x += bot.ent.rect.dimens.width / 2
		midPlayer.y += bot.ent.rect.dimens.height / 2
		rotLine := line{}
		rotLine.newLinePolar(midPlayer, bot.pivShape.bladeLength, bot.pivShape.animationCount)
		crossLine := line{}
		crossLine.newLinePolar(rotLine.p2, bot.pivShape.bladeLength/3, bot.pivShape.animationCount+math.Pi/2)
		frontCrossLine := line{}
		frontCrossLine.newLinePolar(rotLine.p2, bot.pivShape.bladeLength/3, bot.pivShape.animationCount-math.Pi/2)
		bot.pivShape.pivoterShape.lines = []line{rotLine, crossLine, frontCrossLine}

		for blocker, _ := range wepBlockers {
			if blocker.collidesWith(bot.pivShape.pivoterShape) {
				bot.swangin = false
				return
			}
		}
		for _, blocker := range currentTShapes {
			if blocker.collidesWith(bot.pivShape.pivoterShape) {
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

func (s *slasher) updateAim() {
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

func (bot *localEnt) hitremotes() {
	for _, slashee := range remotePlayers {
		if bot.lSlasher.pivShape.checkHitAnimal(&slashee.rSlasher){
			slashee.rSlasher.deth.skipHpUpdate = 2
			bot.hitsToSend = append(bot.hitsToSend, slashee.servId)
		}
	}
}

func (s *pivotingShape) checkHitAnimal(slashee *slasher)bool{
	if _, ok := s.alreadyHit[slashee.ent.collisionId]; ok {
		return false
	}
	if slashee.ent.rect.shape.collidesWith(s.pivoterShape) {
		slashee.deth.redScale = 10
		slashee.deth.hp.CurrentHP -= s.damage
		s.alreadyHit[slashee.ent.collisionId] = true
		return true
	}
	return false
}

func (bot *localAnimal) AIControl() {
	bot.controlCount--
	if bot.controlCount < 1 {
		bot.controlCount = rand.Intn(100)
		bot.locEnt.lSlasher.ent.directions = Directions{
			rand.Intn(9) == 0,
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
	hp           Hitpoints
	skipHpUpdate int
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
	myLocalPlayer.placePlayer()
}

type Directions struct {
	Right bool
	Down  bool
	Left  bool
	Up    bool
}

func (l *localPlayer) updatePlayerControl() {
	l.locEnt.lSlasher.ent.directions.Right = ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight)
	l.locEnt.lSlasher.ent.directions.Down = ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)
	l.locEnt.lSlasher.ent.directions.Left = ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft)
	l.locEnt.lSlasher.ent.directions.Up = ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)
	l.locEnt.lSlasher.atkButton = ebiten.IsKeyPressed(ebiten.KeyX)
}
