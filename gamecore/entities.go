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

const SENDRATE = 10
var sendCount int = SENDRATE
var receiveCount int = 1
var receiveChan = make(chan LocationList)
var otherPlayers = make(map[string]*ServeLocAndEntID)

type ServeLocAndEntID struct {
	//serveloc ServerLocation
	entID    *acceleratingEnt
}

var netbusy = false

func (g *SamGame) Update(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("SamGame ended by player")
	}
	if socketConnection != nil {
		if sendCount > 0 {
			sendCount--
		} else {
			if !netbusy {
				sendCount = SENDRATE
				netbusy = true
				message := ServerMessage{
					Myloc: ServerLocation{myAccelEnt.rect.location.x, myAccelEnt.rect.location.y},
					Mymom: myAccelEnt.moment,
					Mydir: myAccelEnt.directions,
				}
				go func(){
					writeErr := wsjson.Write(context.Background(), socketConnection, message)
					if writeErr != nil {
						log.Println(writeErr)
						closeConn()
						socketConnection = nil
						return
					}
					log.Println("sent my pos", message)
					//go func() {
					var v LocationList
					err1 := wsjson.Read(context.Background(), socketConnection, &v)
					if err1 != nil {
						log.Println(err1)
						closeConn()
						socketConnection = nil
						return
					}
					select{
					case ll:= <- receiveChan:
						log.Println("discarded",ll)
					default:
					}
					receiveChan <- v
					//}()
					netbusy = false
				}()
			}

		}
	}
	if receiveCount<SENDRATE{
		receiveCount++
	}
	select {
	case msg := <-receiveChan:
		log.Println("received message", msg)
		receiveCount = 1
		//message := ServerMessage{
		//	Myloc: ServerLocation{centerOn.location.x, centerOn.location.y},
		//	Mymom: myAccelEnt.moment,
		//	Mydir: myAccelEnt.directions,
		//}
		//writeErr := wsjson.Write(context.Background(), socketConnection, message)
		//if writeErr != nil {
		//	log.Println(writeErr)
		//	closeConn()
		//	socketConnection = nil
		//}
		//fmt.Println("sent my pos", message)

		for _, l := range msg.Locs {
			if res, ok := otherPlayers[l.PNum]; !ok {
				log.Println("adding new player")
				newOtherPlayer := &entityid{}
				accelEnt := newControlledEntity()

				//accelEnt.tracktion = 2
				accelEnt.rect.refreshShape(location{l.Loc.X, l.Loc.Y})
				accelEnt.baseloc = accelEnt.rect.location
				addRemoteMover(accelEnt, newOtherPlayer)
				//rect := newRectangle(location{l.Loc.X, l.Loc.Y}, dimens{20, 40})
				addHitbox(accelEnt.rect.shape, newOtherPlayer)
				addSolid(accelEnt.rect.shape,newOtherPlayer)
				otherPlay := &ServeLocAndEntID{entID: accelEnt}
				otherPlayers[l.PNum] = otherPlay
			} else {

				diffx:=l.Loc.X - res.entID.rect.location.x
				diffy:=l.Loc.Y - res.entID.rect.location.y

				//interpMomentX := res.entID.agility*10
				//if diffx > 0{
				//	l.HisDir.Right = true
				//}else if diffx<0{
				//	interpMomentX = -interpMomentX
				//	l.HisDir.Left = true
				//}
				//interpMomentY := res.entID.agility*10
				//if diffy > 0{
				//	l.HisDir.Up = true
				//	interpMomentY = - interpMomentY
				//}else if diffy<0{
				//	l.HisDir.Down = true
				//}

				//interpMoment := Momentum{int(float64(diffx)*1),int(float64(diffy)*1)}
				//interpMomentX := (diffx*1) + (l.HisMom.Xaxis/2)
				//interpMomentX := (l.HisMom.Xaxis/1)


				//maxinterp := int(res.entID.agility*10)
				//if interpMomentX > int(res.entID.agility){
				//	interpMomentX = int(res.entID.agility)
				//}
				//if interpMomentX < int(-res.entID.agility){
				//	interpMomentX = int(-res.entID.agility)
				//}
				//if math.Abs(float64(res.entID.moment.Xaxis)/10)<1{
				//	interpMomentX+=res.entID.moment.Xaxis
				//}
				//interpMomentY := (diffy*1) + (l.HisMom.Yaxis/2)
				//interpMomentY := (l.HisMom.Yaxis/1)
				//if interpMomentY > int(res.entID.agility){
				//	interpMomentY = int(res.entID.agility)
				//}
				//if interpMomentY < int(-res.entID.agility){
				//	interpMomentY = int(-res.entID.agility)
				//}


				//if math.Abs(float64(res.entID.moment.Yaxis)/10)<1{
				//	interpMomentY+=res.entID.moment.Yaxis
				//}
				//interpMomentX := (l.HisMom.Xaxis + res.entID.moment.Xaxis)/2
				//interpMomentY := (l.HisMom.Yaxis + res.entID.moment.Yaxis)/2
				res.entID.baseloc = res.entID.rect.location
				res.entID.destination = location{diffx/(SENDRATE/2), diffy/(SENDRATE/2)}
				//res.entID.rect.refreshShape(location{l.Loc.X, l.Loc.Y})
				res.entID.directions = l.HisDir
				res.entID.moment = l.HisMom
				//res.entID.moment.Xaxis = int(interpMomentX)
				//res.entID.moment.Yaxis = int(interpMomentY)
				//log.Println("updating player at:", res)
				lagcompcount = DEADRECKON
			}
		}

	default:
	}
	remoteMoversWork()

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
	Mymom Momentum       `json:"mymom"`
	Mydir Directions `json:"mydir"`
}
type LocationList struct {
	Locs []LocWithPNum `json:"locs"`
}

type LocWithPNum struct {
	Loc    ServerLocation `json:"locus"`
	PNum   string         `json:"pnum"`
	HisMom Momentum       `json:"itmom"`
	HisDir Directions       `json:"itdir"`
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
	//defer func() {
	//	closeConn()
	//}()
	//go func(){
	//for {
	//	var v LocationList
	//	err1 := wsjson.Read(context.Background(), socketConnection, &v)
	//	if err1 != nil {
	//		log.Println(err1)
	//		return
	//	}
	//	select{
	//	case ll:= <- receiveChan:
	//		log.Println("discarded",ll)
	//		default:
	//	}
	//	receiveChan <- v
	//}
	//}()
}

var myAccelEnt *acceleratingEnt

func ClientInit() {

	playerid := &entityid{}
	accelplayer := newControlledEntity()
	myAccelEnt = accelplayer
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

	//for i := 1; i < 30; i++ {
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
	//	botDeathable := deathable{}
	//	botDeathable.currentHP = 3
	//	botDeathable.maxHP = 3
	//	botDeathable.deathableShape = moveEnemy.rect
	//	addDeathable(enemyid, &botDeathable)
	//	es := &baseSprite{}
	//	es.swinging = &enemySlasher.swangin
	//	es.redScale = &botDeathable.redScale
	//	es.sprite = playerStandImage
	//	es.owner = moveEnemy
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
}
