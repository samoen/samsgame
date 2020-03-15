package main

import (
	"errors"
	"fmt"

	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/ebiten/ebitenutil"

	"github.com/hajimehoshi/ebiten/inpututil"
)

const screenWidth = 1400
const screenHeight = 1000

func main() {

	initEntities()

	update := func(screen *ebiten.Image) error {

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return errors.New("game ended by player")
		}

		playerMoveSystem.work()
		botsMoveSystem.work()
		collideSystem.work()
		slashSystem.work()
		pivotingSystem.work()

		if ebiten.IsDrawingSkipped() {
			return nil
		}

		renderingSystem.work(screen)
		weaponRenderingSystem.work(screen)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()), 0, 0)

		return nil
	}

	if err := ebiten.Run(update, screenWidth, screenHeight, 1, "sam's cool game"); err != nil {
		log.Fatal(err)
	}
}
