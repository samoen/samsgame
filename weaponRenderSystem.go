package main

import (
	"math"

	"github.com/hajimehoshi/ebiten"
)

type weaponSprite struct {
	// weaponShape *shape
	angle       *float64
	basicSprite baseSprite
}

type baseSprite struct {
	playerRect *rectangle
	sprite     *ebiten.Image
}

var playerSprites = make(map[*entityid]*baseSprite)

func addPlayerSprite(ws *baseSprite, id *entityid) {
	playerSprites[id] = ws
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

func renderEntSprites(s *ebiten.Image) {
	center := renderOffset()
	for _, ps := range playerSprites {
		pOps := &ebiten.DrawImageOptions{}
		// offsetx := -playerSpriteHitboxExceed
		// offsety :=-(2*playerSpriteHitboxExceed)
		pSpriteOffset := center
		pSpriteOffset.x += ps.playerRect.location.x
		pSpriteOffset.y += ps.playerRect.location.y
		pOps.GeoM.Translate(float64(pSpriteOffset.x), float64(pSpriteOffset.y))
		s.DrawImage(ps.sprite, pOps)
	}
	for _, wep := range weapons {
		wepOffset := center
		ownerCenter := rectCenterPoint(*wep.basicSprite.playerRect)
		wepOffset.x += ownerCenter.x
		wepOffset.y += ownerCenter.y

		ew, _ := wep.basicSprite.sprite.Size()
		wepOps := &ebiten.DrawImageOptions{}
		wepOps.GeoM.Translate(-float64(ew)/2, 0)
		wepOps.GeoM.Rotate(*wep.angle - (math.Pi / 2))
		wepOps.GeoM.Translate(float64(wepOffset.x), float64(wepOffset.y))
		s.DrawImage(wep.basicSprite.sprite, wepOps)
	}
}
