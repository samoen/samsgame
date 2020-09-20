package main

import (
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"nhooyr.io/websocket"
)

const (
	maxAxeLength   = 45
	axeRotateSpeed = 0.12
	axeArc         = 3.9
	screenWidth    = 1600
	screenHeight   = 1000
	worldWidth     = screenWidth * 2
	bgTileWidth    = screenWidth
)

var interpTime = 1
var receiveCount = 1
var receiveDebug = ""
var receiveChan = make(chan sockSelecter)
var socketConnection *websocket.Conn
var othersock *websocket.Conn
var toRender []baseSprite
var offset location
var myPNum string
var myLocalPlayer localPlayer
var localAnimals = make(map[*localAnimal]bool)
var remotePlayers = make(map[string]*remotePlayer)
var wepBlockers = make(map[*shape]bool)
var deathAnimations = make(map[*deathAnim]bool)
var bgchan = make(chan ttwithIm)
var bgtiles = make(map[location]*bgLoading)
var ttmap = make(map[tileType]*ebiten.Image)
var ttshapes = make(map[tileType]shape)
var currentTShapes = make(map[location]shape)
var mycenterpoint location
var images imagesStruct

type deathAnim struct {
	sprites   []baseSprite
	rect      rectangle
	inverted  bool
	animcount int
}

type SamGame struct{}

func (g *SamGame) Update(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		closeConn()
		return errors.New("SamGame ended by player")
	}

	respawnsWork()
	socketReceive()

	if len(localAnimals) < 1 {
		animal := slasher{}
		animal.defaultStats()
		animal.moveSpeed = 50
		animal.rect.refreshShape(location{140, 30})
		animal.spawnSafe()
		la := &localAnimal{}
		la.locEnt.lSlasher = animal
		localAnimals[la] = true
	}

	for _, r := range remotePlayers {
		r.remoteMovement()
		if r.rSlasher.atkButton {
			r.rSlasher.startSwing()
		}
		if r.rSlasher.swangin {
			r.rSlasher.progressSwing()
		}
	}

	if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP > 0 {
		myLocalPlayer.updatePlayerControl()
		myLocalPlayer.locEnt.lSlasher.moveCollide()
		myLocalPlayer.locEnt.lSlasher.updateAim()
		myLocalPlayer.locEnt.handleSwing()
		myLocalPlayer.checkHitOthers()
	}

	for l, _ := range localAnimals {
		l.AIControl()
		l.locEnt.lSlasher.moveCollide()
		l.locEnt.lSlasher.updateAim()
		l.locEnt.handleSwing()
		l.checkHitOthers()
	}

	mycenterpoint = rectCenterPoint(myLocalPlayer.locEnt.lSlasher.rect)
	center := mycenterpoint
	center.x *= -1
	center.y *= -1
	center.x += screenWidth / 2
	center.y += screenHeight / 2
	offset = center

	bgShapesWork()
	return nil
}

func (g *SamGame) Draw(screen *ebiten.Image) {
	drawBackground(screen)
	updateSprites()
	renderEntSprites(screen)
	drawHitboxes(screen)
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf(
			"TPS: %0.2f FPS: %0.2f socket: %s",
			ebiten.CurrentTPS(),
			ebiten.CurrentFPS(),
			receiveDebug,
		),
		0,
		0,
	)
}

func (g *SamGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	//return outsideWidth, outsideHeight
	return screenWidth, screenHeight
}

func placeMap() {
	worldBoundRect := rectangle{}
	worldBoundRect.dimens = dimens{worldWidth, worldWidth}
	worldBoundRect.refreshShape(location{0, 0})
	wepBlockers[&worldBoundRect.shape] = true
}
