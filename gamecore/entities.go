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

const worldWidth = 5000

type SamGame struct{}

var pingFrames = 10

var receiveCount = pingFrames
var receiveDebug = ""

type sockSelecter struct{
	ll LocationList
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
	enemyControlWork()
	slashersWork()
	remotePlayersWork()
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

var mySlasher *slasher
var myId *entityid

func addPlayerEntity(playerid *entityid, startloc location, heath Hitpoints) *slasher{
	accelplayer := &acceleratingEnt{}
	accelplayer.rect = newRectangle(
		startloc,
		dimens{20, 40},
	)
	for {
		if normalcollides(*accelplayer.rect.shape, playerid) {
			accelplayer.rect = newRectangle(
				location{startloc.x, accelplayer.rect.location.y + 20},
				dimens{20, 40},
			)
		} else {
			break
		}
	}
	accelplayer.tracktion = 3
	accelplayer.agility = 4
	accelplayer.moveSpeed = 100
	playerSlasher := &slasher{}
	playerSlasher.ent = accelplayer
	playerSlasher.cooldownCount = 0
	playerSlasher.pivShape = &pivotingShape{}
	playerSlasher.pivShape.damage = 2
	playerSlasher.pivShape.pivoterShape = newShape()
	playerSlasher.pivShape.pivotPoint = playerSlasher.ent.rect
	pDeathable := &deathable{}
	pDeathable.hp = heath
	pDeathable.deathableShape = accelplayer.rect
	playerSlasher.deth = pDeathable
	hBarSprite := &baseSprite{}
	hBarSprite.bOps = &ebiten.DrawImageOptions{}
	hBarSprite.sprite = images.empty
	playerSlasher.hbarsprit = hBarSprite
	ps := &baseSprite{}
	ps.bOps = &ebiten.DrawImageOptions{}
	ps.sprite = images.playerStand
	playerSlasher.bsprit = ps
	bs := &baseSprite{}
	bs.bOps = &ebiten.DrawImageOptions{}
	bs.sprite = images.sword
	playerSlasher.wepsprit = bs

	return playerSlasher

}

func placePlayer(){
	pid := &entityid{}
	ps := addPlayerEntity(pid, location{50, 50}, Hitpoints{6, 6})
	centerOn = ps.ent.rect
	mySlasher = ps
	myId = pid
	addSlasher(pid, ps)
}

func ClientInit() {
	i, err := newImages("assets")
	if err != nil {
		log.Fatal(err)
	}
	images = i

	if err := images.empty.Fill(color.White); err != nil {
		log.Fatal(err)
	}
	bgOps.GeoM.Translate(float64(ScreenWidth/2), float64(ScreenHeight/2))

	placePlayer()
	for i := 1; i < 10; i++ {
		enemyid := &entityid{}
		animal := addPlayerEntity(enemyid, location{i*50 + 50, i * 30}, Hitpoints{3, 3})
		addSlasher(enemyid,animal)
		if a, ok := slashers[enemyid]; ok {
			eController := &enemyController{}
			eController.aEnt = a.ent
			addEnemyController(eController, enemyid)
		}
	}

	placeMap()
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
	addBlocker(worldBoundRect.shape, worldBoundaryID)

	diagonalWallID := &entityid{}
	diagonalWall := newShape()
	diagonalWall.lines = []line{
		{
			location{250, 310},
			location{600, 655},
		},
	}

	addBlocker(diagonalWall, diagonalWallID)

	lilRoomID := &entityid{}
	lilRoom := newRectangle(
		location{45, 400},
		dimens{70, 20},
	)
	addBlocker(lilRoom.shape, lilRoomID)

	anotherRoomID := &entityid{}
	anotherRoom := newRectangle(location{900, 1200}, dimens{90, 150})
	addBlocker(anotherRoom.shape,anotherRoomID)
}
