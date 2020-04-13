package main

import (
	"errors"
	"fmt"

	"image/color"
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

var swordImage, _, _ = ebitenutil.NewImageFromFile("assets/axe.png", ebiten.FilterDefault)

// var swordImage *ebiten.Image

var bgImage, _, _ = ebitenutil.NewImageFromFile("assets/8000paint.png", ebiten.FilterDefault)

// var bgImage *ebiten.Image

func main() {

	emptyImage.Fill(color.White)
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
		collideSystem.work()
		slashSystem.work()
		pivotingSystem.work()
		deathSystemwork()

		if ebiten.IsDrawingSkipped() {
			return nil
		}

		drawBackground(screen)
		renderEntSprites(screen)
		drawHitboxes(screen)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()), 0, 0)

		return nil
	}

	if err := ebiten.Run(update, screenWidth, screenHeight, 1, "sam's cool game"); err != nil {
		log.Fatal(err)
	}
}

type deathable struct {
	gotHit         bool
	deathableShape *rectangle
	redScale       int
	currentHP      int
	maxHP          int
	associates     []*entityid
	healthbar      *rectangle
}

var deathables = make(map[*entityid]*deathable)

func addDeathable(id *entityid, d *deathable) {
	id.systems = append(id.systems, hurtable)
	deathables[id] = d

	d.healthbar = newRectangle(d.deathableShape.location, dimens{d.deathableShape.dimens.width, 10})
	hBarEnt := &entityid{}
	hBarRedScale := new(int)
	healthBarSprite := &baseSprite{}
	healthBarSprite.playerRect = d.healthbar
	healthBarSprite.sprite = emptyImage
	healthBarSprite.redScale = hBarRedScale
	healthBarSprite.flip = &directions{}
	d.associates = append(d.associates, hBarEnt)
	addBasicSprite(healthBarSprite, hBarEnt)
}

func deathSystemwork() {
	for dID, mDeathable := range deathables {
		mDeathable.healthbar.location = location{mDeathable.deathableShape.location.x, mDeathable.deathableShape.location.y - 10}
		mDeathable.healthbar.dimens.width = mDeathable.currentHP * mDeathable.deathableShape.dimens.width / mDeathable.maxHP
		if mDeathable.redScale > 0 {
			mDeathable.redScale--
		}
		if mDeathable.gotHit {
			mDeathable.redScale = 10
			mDeathable.gotHit = false
			mDeathable.currentHP--
		}
		if mDeathable.currentHP < 1 {
			for _, associate := range mDeathable.associates {
				eliminate(associate)
			}
			eliminate(dID)
		}
	}
}

func eliminate(id *entityid) {

	for _, sys := range id.systems {
		switch sys {
		case spriteRenderable:
			delete(basicSprites, id)
		case hitBoxRenderable:
			delete(hitBoxes, id)
		case moveCollider:
			delete(collideSystem.movers, id)
		case solidCollider:
			delete(collideSystem.solids, id)
		case enemyControlled:
			delete(botsMoveSystem.movers, id)
		case abilityActivator:
			delete(slashSystem.slashers, id)
		case hurtable:
			delete(deathables, id)
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
