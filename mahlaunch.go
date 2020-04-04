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
const bgTileWidth = 2500

var playerStandImage, _, _ = ebitenutil.NewImageFromFile("assets/playerstand.png", ebiten.FilterDefault)

// var playerStandImage *ebiten.Image

var emptyImage, _, _ = ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)

var swordImage, _, _ = ebitenutil.NewImageFromFile("assets/sword.png", ebiten.FilterDefault)

// var swordImage *ebiten.Image

var bgImage, _, _ = ebitenutil.NewImageFromFile("assets/8000paint.png", ebiten.FilterDefault)

// var bgImage *ebiten.Image

func main() {
	// img, _, err := image.Decode(bytes.NewReader(playerStandPng))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// playerStandImage, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	// sword, _, err := image.Decode(bytes.NewReader(swordPng))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// swordImage, _ = ebiten.NewImageFromImage(sword, ebiten.FilterDefault)

	// bgIm, _, err := image.Decode(bytes.NewReader(backgroundPng))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// bgImage, _ = ebiten.NewImageFromImage(bgIm, ebiten.FilterDefault)

	initEntities()

	update := func(screen *ebiten.Image) error {

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return errors.New("game ended by player")
		}

		updatePlayerControl()
		// botsMoveSystem.work()
		collideSystem.work()
		slashSystem.work()
		pivotingSystem.work()
		deathSystemwork()
		if ebiten.IsDrawingSkipped() {
			return nil
		}

		myBgOps := *bgOps
		myBgOps.GeoM.Translate(float64(-centerOn.location.x), float64(-centerOn.location.y))
		myBgOps.GeoM.Translate(float64(-centerOn.dimens.width/2), float64(-centerOn.dimens.height/2))

		tilesAcross := worldWidth / bgTileWidth

		for i := 0; i < tilesAcross; i++ {
			for j := 0; j < tilesAcross; j++ {
				tileOps := myBgOps
				tileOps.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))
				screen.DrawImage(bgImage, &tileOps)
			}
		}

		renderWeaponSprites(screen)

		renderingSystem.work(screen)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()), 0, 0)

		return nil
	}

	if err := ebiten.Run(update, screenWidth, screenHeight, 1, "sam's cool game"); err != nil {
		log.Fatal(err)
	}
}

type deathable struct {
	dead bool
}

var deathables = make(map[*entityid]deathable)

func deathSystemwork() {
	for dID := range deathables {
		for _, associate := range dID.associates {
			eliminate(associate)
		}
		eliminate(dID)
		delete(deathables, dID)
	}
}

func eliminate(id *entityid) {

	for _, sys := range id.systems {
		switch sys {
		case spriteRenderable:
			delete(playerSprites, id)
		case hitBoxRenderable:
			delete(renderingSystem.shapes, id)
		case moveCollider:
			delete(collideSystem.movers, id)
		case solidCollider:
			delete(collideSystem.solids, id)
		case enemyControlled:
			delete(botsMoveSystem.movers, id)
		case abilityActivator:
			delete(slashSystem.slashers, id)
		case hurtable:
			delete(pivotingSystem.slashees, id)
		case pivotingHitbox:
			delete(pivotingSystem.pivoters, id)
		case rotatingSprite:
			delete(weapons, id)
		case playerControlled:
			delete(playerControllables, id)
		case weaponBlocker:
			delete(pivotingSystem.blockers, id)
		}
	}
}
