package gamecore

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

const ScreenWidth = 700
const ScreenHeight = 500
const worldWidth = ScreenWidth * 4

var bgTileWidth = ScreenWidth / 2

type SamGame struct{}

var pingFrames = 10

var receiveCount = pingFrames
var receiveDebug = ""

type sockSelecter struct {
	ll   LocationList
	sock *websocket.Conn
}

var receiveChan = make(chan sockSelecter)

func (g *SamGame) Update(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		closeConn()
		return errors.New("SamGame ended by player")
	}

	respawnsWork()
	socketReceive()
	updatePlayerControl()
	myLocalPlayer.locEnt.lSlasher.ent.moveCollide()
	myLocalPlayer.locEnt.lSlasher.updateAim()
	myLocalPlayer.locEnt.lSlasher.handleSwing()
	if myLocalPlayer.locEnt.lSlasher.swangin {
		myLocalPlayer.locEnt.hitremotes()
		myLocalPlayer.locEnt.HitAnimals()
	}
	animalsWork()
	remotePlayersWork()

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

type ServerMessage struct {
	Myloc    ServerLocation
	Mymom    Momentum
	Mydir    Directions
	Myaxe    Weapon
	Myhealth Hitpoints
	MyPNum   string
}

type Weapon struct {
	Swinging   bool
	Startangle float64
	IHit       []string
	Dmg        int
}

type LocationList struct {
	Locs     []ServerMessage
	YourPNum string
}

type ServerLocation struct {
	X int
	Y int
}

var myLocalPlayer *localPlayer

func (s *slasher) newSlasher() {
	accelplayer := acceleratingEnt{}
	accelplayer.rect = newRectangle(
		location{50,50},
		dimens{20, 40},
	)
	accelplayer.agility = 4
	accelplayer.moveSpeed = 100
	s.ent = accelplayer
	s.cooldownCount = 0
	s.pivShape = pivotingShape{}
	s.pivShape.damage = 2
	s.pivShape.pivoterShape = shape{}
	pDeathable := deathable{}
	pDeathable.hp = Hitpoints{6,6}
	s.deth = pDeathable
	hBarSprite := baseSprite{}
	hBarSprite.bOps = &ebiten.DrawImageOptions{}
	hBarSprite.sprite = images.empty
	s.hbarsprit = hBarSprite
	ps := baseSprite{}
	ps.bOps = &ebiten.DrawImageOptions{}
	ps.sprite = images.playerStand
	s.bsprit = ps
	bs := baseSprite{}
	bs.bOps = &ebiten.DrawImageOptions{}
	bs.sprite = images.sword
	s.wepsprit = bs
}

func placePlayer() {
	ps := &slasher{}
	ps.newSlasher()
	ps.ent.rect.refreshShape(location{50, 50})
	ps.deth.hp = Hitpoints{6, 6}
	mycenterpoint = rectCenterPoint(ps.ent.rect)
	myLocalEnt := localEnt{}
	myLocalEnt.lSlasher = ps
	locPlayer := localPlayer{}
	locPlayer.locEnt = myLocalEnt
	myLocalPlayer = &locPlayer
	myLocalPlayer.locEnt.lSlasher.ent.spawnSafe()
}

func (accelplayer *acceleratingEnt) spawnSafe() {
	for {
		if normalcollides(*accelplayer.rect.shape, accelplayer.rect.shape) {
			accelplayer.rect = newRectangle(
				location{accelplayer.rect.location.x, accelplayer.rect.location.y + 20},
				dimens{20, 40},
			)
		} else {
			break
		}
	}
}

func ClientInit() {
	images = imagesStruct{}
	images.newImages()

	if err := images.empty.Fill(color.White); err != nil {
		log.Fatal(err)
	}

	placePlayer()

	for i := 1; i < 10; i++ {
		enemyid := &entityid{}
		animal := &slasher{}
		animal.newSlasher()
		animal.ent.rect.refreshShape(location{i*50 + 50, i * 30})
		la := &localAnimal{}
		la.locEnt.lSlasher = animal
		slashers[enemyid] = la
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
	worldBoundaryID := &entityid{}
	worldBoundRect := newRectangle(
		location{0, 0},
		dimens{worldWidth, worldWidth},
	)
	wepBlockers[worldBoundaryID] = worldBoundRect.shape
}
