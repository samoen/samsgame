package main

import (
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"image/color"
	"log"
	"nhooyr.io/websocket"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	maxAxeLength   = 45
	axeRotateSpeed = 0.12
	axeArc         = 3.9
	ScreenWidth    = 700
	ScreenHeight   = 500
	worldWidth     = ScreenWidth * 4
	bgTileWidth    = ScreenWidth
	interpTime     = 4
	deathreckTime  = 4
)

var pingFrames = 10
var receiveCount = pingFrames
var receiveDebug = ""
var receiveChan = make(chan sockSelecter)
var socketConnection *websocket.Conn
var othersock *websocket.Conn
var toRender []baseSprite
var offset location
var myPNum string
var myLocalPlayer localPlayer
var slashers = make(map[*localAnimal]bool)
var remotePlayers = make(map[string]*remotePlayer)
var wepBlockers = make(map[*shape]bool)
var deathAnimations = make(map[*deathAnim]bool)

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

	for _, r := range remotePlayers {
		r.remoteMovement()
		r.rSlasher.handleSwing()
	}

	if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP > 0 {
		myLocalPlayer.updatePlayerControl()
		myLocalPlayer.locEnt.lSlasher.ent.moveCollide()
		myLocalPlayer.locEnt.lSlasher.updateAim()
		myLocalPlayer.locEnt.lSlasher.handleSwing()
		myLocalPlayer.checkHitOthers()
	}

	for l, _ := range slashers {
		l.AIControl()
		l.locEnt.lSlasher.ent.moveCollide()
		l.locEnt.lSlasher.updateAim()
		l.locEnt.lSlasher.handleSwing()
		l.checkHitOthers()
	}


	mycenterpoint = rectCenterPoint(myLocalPlayer.locEnt.lSlasher.ent.rect)
	center := mycenterpoint
	center.x *= -1
	center.y *= -1
	center.x += ScreenWidth / 2
	center.y += ScreenHeight / 2
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
	//ScreenWidth = outsideWidth
	//ScreenHeight = outsideHeight
	return ScreenWidth, ScreenHeight
	//magnification := outsideWidth/ScreenWidth
	//return outsideWidth+outsideWidth-ScreenWidth, outsideHeight+outsideWidth-ScreenHeight
	//return outsideWidth, outsideHeight
}

func ClientInit() {
	images = imagesStruct{}
	images.newImages()

	if err := images.empty.Fill(color.White); err != nil {
		log.Fatal(err)
	}
	myLocalPlayer = localPlayer{}
	myLocalPlayer.locEnt.lSlasher.defaultStats()
	myLocalPlayer.placePlayer()

	for i := 1; i < 10; i++ {
		animal := slasher{}
		animal.defaultStats()
		animal.ent.moveSpeed = 50
		animal.ent.rect.refreshShape(location{i*50 + 50, i * 30})
		la := &localAnimal{}
		la.locEnt.lSlasher = animal
		slashers[la] = true
	}

	placeMap()

	tilesAcross := worldWidth / bgTileWidth
	for i := -1; i < tilesAcross+1; i++ {
		for j := -1; j < tilesAcross+1; j++ {
			ttype := blank
			if j > tilesAcross-1 || i > tilesAcross-1 || j < 0 || i < 0 {
				ttype = offworld
			} else if j%3 == 0 || i%3 == 0 {
				ttype = rocky
			}
			bgl := &bgLoading{}
			bgl.tiletyp = ttype
			bgl.ops = &ebiten.DrawImageOptions{}
			bgtiles[location{i, j}] = bgl
		}
	}

	ttshapes[blank] = shape{lines: []line{line{location{180, 5}, location{140, 60}}}}
	ttshapes[rocky] = shape{lines: []line{line{location{80, 20}, location{80, 120}}}}

	go func() {
		time.Sleep(1500 * time.Millisecond)
		connectToServer()
	}()

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("sams cool game")
	ebiten.SetWindowResizable(true)

	samgame := &SamGame{}

	if err := ebiten.RunGame(samgame); err != nil {
		closeConn()
		log.Fatal(err)
		return
	}
	closeConn()
	log.Println("exited main")
}

func placeMap() {
	worldBoundRect := rectangle{}
	worldBoundRect.dimens = dimens{worldWidth, worldWidth}
	worldBoundRect.refreshShape(location{0, 0})
	wepBlockers[&worldBoundRect.shape] = true
}
