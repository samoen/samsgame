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
	"time"
)

const worldWidth = 5000
type SamGame struct{}

func (g *SamGame) Update(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("SamGame ended by player")
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

func connectToServer(){
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://localhost:8080/ws", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	for{
		err = wsjson.Write(ctx, c, "hi")
		if err != nil {
			log.Println(err)
			return
		}
		time.Sleep(500 * time.Millisecond)
		fmt.Println("sent it and waited")
	}

	c.Close(websocket.StatusNormalClosure, "")
}

func init() {
	go connectToServer()
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
}
