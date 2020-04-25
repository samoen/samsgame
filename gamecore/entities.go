package gamecore

import (
	"context"
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const worldWidth = 5000

type SamGame struct{}

var sendCount int = 60
var receiveChan = make(chan LocationList)
var otherPlayers = make(map[string]*ServeLocAndEntID)

type ServeLocAndEntID struct {
	serveloc ServerLocation
	entID    *rectangle
}

func (g *SamGame) Update(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("SamGame ended by player")
	}
	//if socketConnection != nil {
	//	if sendCount > 0 {
	//		sendCount--
	//	} else {
	//		sendCount = 30
	//		message := ServerMessage{Myloc: ServerLocation{centerOn.location.x, centerOn.location.y}}
	//		writeErr := wsjson.Write(context.Background(), socketConnection, message)
	//		if writeErr != nil {
	//			log.Println(writeErr)
	//			closeConn()
	//			socketConnection = nil
	//		}
	//		fmt.Println("sent my pos", message)
	//	}
	//}
	select {
	case msg := <-receiveChan:
		fmt.Println("received message", msg)
		for _, l := range msg.Locs {
			if res, ok := otherPlayers[l.PNum]; !ok {
				fmt.Println("adding new player")
				newOtherPlayer := &entityid{}
				rect := newRectangle(location{l.Loc.X, l.Loc.Y}, dimens{20, 40})
				addHitbox(rect.shape, newOtherPlayer)
				otherPlay := &ServeLocAndEntID{serveloc: l.Loc, entID: rect}
				otherPlayers[l.PNum] = otherPlay
			} else {
				fmt.Println("updating player at:", l.Loc)
				res.entID.refreshShape(location{l.Loc.X, l.Loc.Y})
			}
		}
		message := ServerMessage{Myloc: ServerLocation{centerOn.location.x, centerOn.location.y}}
		writeErr := wsjson.Write(context.Background(), socketConnection, message)
		if writeErr != nil {
			log.Println(writeErr)
			closeConn()
			socketConnection = nil
		}
		fmt.Println("sent my pos", message)
	default:
	}

	updatePlayerControl()
	enemyControlWork()
	collisionSystemWork()
	slashersWork()
	pivotSystemWork()
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
	err := socketConnection.Close(websocket.StatusInternalError, "closed from client defer")
	if err != nil {
		log.Println(err)
	}
}

type ServerMessage struct {
	Myloc ServerLocation `json:"myloc"`
}
type LocationList struct {
	Locs []LocWithPNum `json:"locs"`
}

type LocWithPNum struct {
	Loc  ServerLocation
	PNum string
}

type ServerLocation struct {
	X int `json:"x"`
	Y int `json:"y"`
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
	defer func() {
		closeConn()
	}()
	//go func(){
	for {
		var v LocationList
		err1 := wsjson.Read(context.Background(), socketConnection, &v)
		if err1 != nil {
			log.Println(err1)
			return
		}
		receiveChan <- v
	}
	//}()
}

func ClientInit() {

	playerid := &entityid{}
	accelplayer := newControlledEntity()
	addPlayerControlled(accelplayer, playerid)
	addMoveCollider(accelplayer, playerid)
	addSolid(accelplayer.rect.shape, playerid)
	addHitbox(accelplayer.rect.shape, playerid)
	centerOn = accelplayer.rect
	playerSlasher := newSlasher(accelplayer)
	addSlasher(playerid, playerSlasher)
	pDeathable := deathable{}
	pDeathable.currentHP = 6
	pDeathable.maxHP = 6
	pDeathable.deathableShape = accelplayer.rect
	addDeathable(playerid, &pDeathable)

	ps := &baseSprite{}
	ps.redScale = new(int)
	ps.swinging = &playerSlasher.swangin
	ps.sprite = playerStandImage
	ps.owner = accelplayer
	addBasicSprite(ps, playerid)

	for i := 1; i < 30; i++ {
		enemyid := &entityid{}
		moveEnemy := newControlledEntity()
		moveEnemy.rect.refreshShape(location{i*50 + 50, i * 30})
		enemySlasher := newSlasher(moveEnemy)
		addSlasher(enemyid, enemySlasher)
		addHitbox(moveEnemy.rect.shape, enemyid)
		addMoveCollider(moveEnemy, enemyid)
		addSolid(moveEnemy.rect.shape, enemyid)
		eController := &enemyController{}
		eController.aEnt = moveEnemy
		addEnemyController(eController, enemyid)

		botDeathable := deathable{}
		botDeathable.currentHP = 3
		botDeathable.maxHP = 3
		botDeathable.deathableShape = moveEnemy.rect
		addDeathable(enemyid, &botDeathable)
		es := &baseSprite{}
		es.swinging = &enemySlasher.swangin
		es.redScale = &botDeathable.redScale
		es.sprite = playerStandImage
		es.owner = moveEnemy
		addBasicSprite(es, enemyid)
	}

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
}
