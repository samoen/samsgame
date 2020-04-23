package gamecore

import (
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten"
)

type weaponSprite struct {
	angle  *float64
	owner  *slasher
	sprite *ebiten.Image
}

type baseSprite struct {
	owner    *acceleratingEnt
	swinging *bool
	sprite   *ebiten.Image
	redScale *int
	lastflip bool
}

type healthBarSprite struct {
	ownerDeathable *deathable
}
var ScreenWidth = 1400
var ScreenHeight = 1000
var bgTileWidth = 2500

var playerStandImage, _, _ = ebitenutil.NewImageFromFile(
	"assets/playerstand.png",
	ebiten.FilterDefault,
)

// var playerStandImage *ebiten.Image

var emptyImage, _, _ = ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)

var swordImage, _, _ = ebitenutil.NewImageFromFile("assets/axe.png", ebiten.FilterDefault)

// var swordImage *ebiten.Image

var bgImage, _, _ = ebitenutil.NewImageFromFile("assets/8000paint.png", ebiten.FilterDefault)

// var bgImage *ebiten.Image
var healthbars = make(map[*entityid]*healthBarSprite)

func addHealthBarSprite(h *healthBarSprite, id *entityid) {
	healthbars[id] = h
	id.systems = append(id.systems, healthBarRenderable)
}

var basicSprites = make(map[*entityid]*baseSprite)

func addBasicSprite(ws *baseSprite, id *entityid) {
	basicSprites[id] = ws
	id.systems = append(id.systems, spriteRenderable)
}

var weapons = make(map[*entityid]*weaponSprite)

// var playerSpriteHitboxExceed = 10

func addWeaponSprite(s *weaponSprite, id *entityid) {
	weapons[id] = s
	id.systems = append(id.systems, rotatingSprite)
}

var centerOn *rectangle

func renderingCenter() location {
	l := rectCenterPoint(*centerOn)
	l.x *= -1
	l.y *= -1
	return l
}

func renderOffset() location {
	center := renderingCenter()
	center.x += ScreenWidth / 2
	center.y += ScreenHeight / 2
	return center
}

func rectCenterPoint(r rectangle) location {
	x := r.location.x + (r.dimens.width / 2)
	y := r.location.y + (r.dimens.height / 2)
	return location{x, y}
}

var bgOps = &ebiten.DrawImageOptions{}

func init() {
	bgOps.GeoM.Translate(float64(ScreenWidth/2), float64(ScreenHeight/2))
}

func drawBackground(screen *ebiten.Image) {
	myBgOps := *bgOps
	myBgOps.GeoM.Translate(float64(-centerOn.location.x), float64(-centerOn.location.y))
	myBgOps.GeoM.Translate(float64(-centerOn.dimens.width/2), float64(-centerOn.dimens.height/2))

	tilesAcross := worldWidth / bgTileWidth

	for i := 0; i < tilesAcross; i++ {
		for j := 0; j < tilesAcross; j++ {
			tileOps := myBgOps
			tileOps.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))
			if err:=screen.DrawImage(bgImage, &tileOps);err !=nil{
				log.Fatal(err)
			}
		}
	}
}

func scaleToDimension(dims dimens, img *ebiten.Image) {
	imW, imH := img.Size()
	wRatio := float64(dims.width) / float64(imW)
	hRatio := float64(dims.height) / float64(imH)
	drawOps.GeoM.Scale(wRatio, hRatio)
}

func cameraShift(loc location, pSpriteOffset location) {
	pSpriteOffset.x += loc.x
	pSpriteOffset.y += loc.y
	drawOps.GeoM.Translate(float64(pSpriteOffset.x), float64(pSpriteOffset.y))
}
var drawOps = &ebiten.DrawImageOptions{}

func renderEntSprites(s *ebiten.Image) {
	center := renderOffset()
	for _, ps := range basicSprites {
		if !*ps.swinging{
			if ps.owner.directions.left && !ps.owner.directions.right {
				ps.lastflip = true
			}
			if ps.owner.directions.right && !ps.owner.directions.left {
				ps.lastflip = false
			}
		}
		drawOps.GeoM.Reset()
		scaleToDimension(ps.owner.rect.dimens, ps.sprite)
		if ps.lastflip {
			drawOps.GeoM.Scale(-1, 1)
			drawOps.GeoM.Translate(float64(ps.owner.rect.dimens.width), 0)
		}
		cameraShift(ps.owner.rect.location, center)

		drawOps.ColorM.Translate(float64(*ps.redScale), 0, 0, 0)

		if err:= s.DrawImage(ps.sprite, drawOps);err != nil{
			log.Fatal(err)
		}
		drawOps.ColorM.Reset()
	}
	for _, wep := range weapons {
		_, imH := wep.sprite.Size()
		hRatio := float64(swordLength+swordLength/4) / float64(imH)

		drawOps.GeoM.Reset()
		drawOps.GeoM.Scale(hRatio, hRatio)
		drawOps.GeoM.Translate(-float64(wep.owner.ent.rect.dimens.width)/2, 0)
		drawOps.GeoM.Rotate(*wep.angle - (math.Pi / 2))

		ownerCenter := rectCenterPoint(*wep.owner.ent.rect)
		cameraShift(ownerCenter, center)

		if err:=s.DrawImage(wep.sprite, drawOps);err != nil{
			log.Fatal(err)
		}
	}

	for _, hBarSprite := range healthbars {
		healthbarlocation := location{hBarSprite.ownerDeathable.deathableShape.location.x, hBarSprite.ownerDeathable.deathableShape.location.y - 10}
		healthbardimenswidth := hBarSprite.ownerDeathable.currentHP * hBarSprite.ownerDeathable.deathableShape.dimens.width / hBarSprite.ownerDeathable.maxHP
		drawOps.GeoM.Reset()
		scaleToDimension(dimens{healthbardimenswidth, 5}, emptyImage)
		cameraShift(healthbarlocation, center)
		if err:=s.DrawImage(emptyImage, drawOps);err != nil{
			log.Fatal(err)
		}
	}
}
