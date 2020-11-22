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
	axeRotateSpeed = 0.19
	axeArc         = 3.9
	screenWidth    = 900
	screenHeight   = 500
	worldWidth     = 4000
	bgTileWidth    = 20
	tilesperChunk  = 50
	chunkWidth     = tilesperChunk * bgTileWidth
	tilesAcross    = worldWidth / bgTileWidth
	downscale      = 20.0
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
var bgtilesNew = make(map[location]*backgroundTile)
var tileRenderBuffer *ebiten.Image
var mapChunks = make(map[location]*ebiten.Image)
var currentTShapes = make(map[location]shape)
var mycenterpoint location
var images imagesStruct
var zoom = 1.0

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

	//zoom += float64(inpututil.KeyPressDuration(ebiten.KeyR))/200
	//zoom -= float64(inpututil.KeyPressDuration(ebiten.KeyF))/200
	//if zoom < 0.1 {
	//	zoom = 0.1
	//}
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		if zoom > -240 {
			zoom -= 2
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyF) {
		if zoom < 240 {
			zoom += 2
		}
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

	mycenterpoint = myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint()
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
	drawBufferedTiles(screen)
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
