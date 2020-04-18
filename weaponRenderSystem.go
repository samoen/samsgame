package main

import (
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
	sprite   *ebiten.Image
	redScale *int
	lastflip bool
}

type healthBarSprite struct {
	ownerDeathable *deathable
}

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
	center.x += screenWidth / 2
	center.y += screenHeight / 2
	return center
}

func rectCenterPoint(r rectangle) location {
	x := r.location.x + (r.dimens.width / 2)
	y := r.location.y + (r.dimens.height / 2)
	return location{x, y}
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
			screen.DrawImage(bgImage, &tileOps)
		}
	}
}

func makeOps(dims dimens, img *ebiten.Image) *ebiten.DrawImageOptions {
	pOps := &ebiten.DrawImageOptions{}
	imW, imH := img.Size()
	wRatio := float64(dims.width) / float64(imW)
	hRatio := float64(dims.height) / float64(imH)
	pOps.GeoM.Scale(float64(wRatio), float64(hRatio))

	return pOps
}

func cameraShift(ops *ebiten.DrawImageOptions, loc location, pSpriteOffset location) {
	pSpriteOffset.x += loc.x
	pSpriteOffset.y += loc.y
	ops.GeoM.Translate(float64(pSpriteOffset.x), float64(pSpriteOffset.y))
}

func renderEntSprites(s *ebiten.Image) {
	center := renderOffset()
	for _, ps := range basicSprites {
		if ps.owner.directions.left && !ps.owner.directions.right {
			ps.lastflip = true
		}
		if ps.owner.directions.right && !ps.owner.directions.left {
			ps.lastflip = false
		}
		pOps := makeOps(ps.owner.rect.dimens, ps.sprite)
		if ps.lastflip {
			pOps.GeoM.Scale(-1, 1)
			pOps.GeoM.Translate(float64(ps.owner.rect.dimens.width), 0)
		}
		cameraShift(pOps, ps.owner.rect.location, center)

		pOps.ColorM.Translate(float64(*ps.redScale), 0, 0, 0)

		s.DrawImage(ps.sprite, pOps)
	}
	for _, wep := range weapons {

		wepOps := &ebiten.DrawImageOptions{}
		_, imH := wep.sprite.Size()
		hRatio := float64(swordLength+swordLength/4) / float64(imH)
		wepOps.GeoM.Scale(float64(hRatio), float64(hRatio))

		wepOps.GeoM.Translate(-float64(wep.owner.ent.rect.dimens.width)/2, 0)
		wepOps.GeoM.Rotate(*wep.angle - (math.Pi / 2))

		ownerCenter := rectCenterPoint(*wep.owner.ent.rect)
		cameraShift(wepOps, ownerCenter, center)

		s.DrawImage(wep.sprite, wepOps)
	}

	for _, hBarSprite := range healthbars {
		healthbarlocation := location{hBarSprite.ownerDeathable.deathableShape.location.x, hBarSprite.ownerDeathable.deathableShape.location.y - 10}
		healthbardimenswidth := hBarSprite.ownerDeathable.currentHP * hBarSprite.ownerDeathable.deathableShape.dimens.width / hBarSprite.ownerDeathable.maxHP
		pOps := makeOps(dimens{healthbardimenswidth, 5}, emptyImage)
		cameraShift(pOps, healthbarlocation, center)
		s.DrawImage(emptyImage, pOps)
	}
}
