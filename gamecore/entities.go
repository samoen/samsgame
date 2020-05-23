package gamecore

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const worldWidth = 5000

type SamGame struct{}

var pingFrames = 10

var receiveCount = pingFrames
var receiveChan = make(chan LocationList)
var otherPlayers = make(map[string]*entityid)

//type RemoteMover struct {
//	destination location
//	baseloc     location
//	endpoint    location
//	accelEnt    *acceleratingEnt
//}

func (g *SamGame) Update(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		closeConn()
		return errors.New("SamGame ended by player")
	}
	for _, ps := range basicSprites {
		ps.bOps.ColorM.Reset()
		ps.bOps.GeoM.Reset()
	}
	respawnsWork()
	socketReceive()
	//remoteMoversWork()
	updatePlayerControl()
	enemyControlWork()
	collisionSystemWork()
	slashersWork()
	//pivotSystemWork()
	deathSystemwork()
	return nil
}

func (g *SamGame) Draw(screen *ebiten.Image) {
	drawBackground(screen)
	renderEntSprites(screen)
	drawHitboxes(screen)
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf(
			"TPS: %0.2f FPS: %0.2f",
			ebiten.CurrentTPS(),
			ebiten.CurrentFPS(),
		),
		0,
		0,
	)
}

func (g *SamGame) Layout(outsideWidth, outsideHeight int) (w, h int) {
	//ScreenWidth = outsideWidth
	//ScreenHeight = outsideHeight
	return ScreenWidth, ScreenHeight
	//magnification := outsideWidth/ScreenWidth
	//return outsideWidth+outsideWidth-ScreenWidth, outsideHeight+outsideWidth-ScreenHeight
	//return outsideWidth, outsideHeight
}

var socketConnection *websocket.Conn

func closeConn() {
	if socketConnection != nil {
		err := socketConnection.Close(websocket.StatusNormalClosure, "closed from client defer")
		if err != nil {
			log.Println(err)
		}
	}
}

type ServerMessage struct {
	Myloc    ServerLocation
	Mymom    Momentum
	Mydir    Directions
	Myaxe    Weapon
	Myhealth Hitpoints
}

type Weapon struct {
	Swinging   bool
	Startangle float64
	IHit       []string
}

type LocationList struct {
	Locs []LocWithPNum
}

type LocWithPNum struct {
	Loc         ServerLocation
	PNum        string
	ServMessage ServerMessage
	YouCopped   bool
}

type ServerLocation struct {
	X int
	Y int
}

func connectToServer() {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//defer cancel()

	var err error
	socketConnection, _, err = websocket.Dial(context.Background(), "ws://localhost:8080/ws", nil)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			var v LocationList
			err1 := wsjson.Read(context.Background(), socketConnection, &v)
			if err1 != nil {
				log.Println(err1)
				closeConn()
				socketConnection = nil
				return
			}

			//go func(){
			//select {
			//case
			receiveChan <- v
			//:
			//default:
			//}
			//}()
		}
	}()
}

var myAccelEnt *acceleratingEnt
var mySlasher *slasher
var myDeathable *deathable

func addLocalPlayer() {
	playerid := &entityid{}
	accelplayer := newControlledEntity()
	addPlayerControlled(accelplayer, playerid)
	addMoveCollider(accelplayer, playerid)
	addSolid(accelplayer.rect.shape, playerid)
	addHitbox(accelplayer.rect.shape, playerid)
	centerOn = accelplayer.rect
	playerSlasher := newSlasher(accelplayer)
	addSlasher(playerid, playerSlasher)
	pDeathable := &deathable{}
	pDeathable.hp.CurrentHP = 6
	pDeathable.hp.MaxHP = 6
	pDeathable.deathableShape = accelplayer.rect
	addDeathable(playerid, pDeathable)

	hBarEnt := &entityid{}
	hBarSprite := &baseSprite{}
	hBarSprite.bOps = &ebiten.DrawImageOptions{}
	hBarSprite.sprite = emptyImage
	//hBarSprite.updateAsHealthbar(*pDeathable)
	pDeathable.hBarid = hBarEnt
	//playerid.linked = append(playerid.linked, hBarEnt)
	addBasicSprite(hBarSprite, hBarEnt)

	mySlasher = playerSlasher
	myAccelEnt = accelplayer
	myDeathable = pDeathable

	ps := &baseSprite{}
	ps.bOps = &ebiten.DrawImageOptions{}
	//ps.redScale = &pDeathable.redScale
	//ps.swinging = &playerSlasher.swangin
	ps.sprite = playerStandImage
	//ps.owner = accelplayer
	addBasicSprite(ps, playerid)
}

func ClientInit() {
	addLocalPlayer()

	//for i := 1; i < 1; i++ {
	//	enemyid := &entityid{}
	//	moveEnemy := newControlledEntity()
	//	moveEnemy.rect.refreshShape(location{i*50 + 50, i * 30})
	//	enemySlasher := newSlasher(moveEnemy)
	//	addSlasher(enemyid, enemySlasher)
	//	addHitbox(moveEnemy.rect.shape, enemyid)
	//	addMoveCollider(moveEnemy, enemyid)
	//	addSolid(moveEnemy.rect.shape, enemyid)
	//	eController := &enemyController{}
	//	eController.aEnt = moveEnemy
	//	addEnemyController(eController, enemyid)
	//
	//	botDeathable := &deathable{}
	//	botDeathable.hp.CurrentHP = 3
	//	botDeathable.hp.MaxHP = 3
	//	botDeathable.deathableShape = moveEnemy.rect
	//	addDeathable(enemyid, botDeathable)
	//
	//	hBarEnt := &entityid{}
	//	hBarSprite := &baseSprite{}
	//	hBarSprite.bOps = &ebiten.DrawImageOptions{}
	//	hBarSprite.sprite = emptyImage
	//	//hBarSprite.updateAsHealthbar(*botDeathable)
	//	botDeathable.hBarid = hBarEnt
	//	enemyid.linked = append(enemyid.linked, hBarEnt)
	//	addBasicSprite(hBarSprite,hBarEnt)
	//
	//	es := &baseSprite{}
	//	es.bOps = &ebiten.DrawImageOptions{}
	//	es.sprite = playerStandImage
	//	addBasicSprite(es, enemyid)
	//}

	worldBoundaryID := &entityid{}
	worldBoundRect := newRectangle(
		location{0, 0},
		dimens{worldWidth, worldWidth},
	)
	addHitbox(worldBoundRect.shape, worldBoundaryID)
	addSolid(worldBoundRect.shape, worldBoundaryID)
	addBlocker(worldBoundRect.shape, worldBoundaryID)

	diagonalWallID := &entityid{}
	diagonalWall := newShape()
	diagonalWall.lines = []line{
		{
			location{250, 310},
			location{600, 655},
		},
	}

	addHitbox(diagonalWall, diagonalWallID)
	addSolid(diagonalWall, diagonalWallID)
	addBlocker(diagonalWall, diagonalWallID)

	lilRoomID := &entityid{}
	lilRoom := newRectangle(
		location{45, 400},
		dimens{70, 20},
	)
	addBlocker(lilRoom.shape, lilRoomID)
	addHitbox(lilRoom.shape, lilRoomID)
	addSolid(lilRoom.shape, lilRoomID)

	anotherRoomID := &entityid{}
	anotherRoom := newRectangle(location{900, 1200}, dimens{90, 150})
	addHitbox(anotherRoom.shape, anotherRoomID)
	addSolid(anotherRoom.shape, anotherRoomID)

	go connectToServer()
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
