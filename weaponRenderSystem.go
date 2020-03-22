package main

import (
	"math"

	"github.com/hajimehoshi/ebiten"
)

type weaponSprite struct {
	weaponShape *shape
	angle       *float64
	basicSprite playerSprite
}

type playerSprite struct {
	playerRect *rectangle
	sprite     *ebiten.Image
}

var playerSprites = make(map[*entityid]*playerSprite)

func addPlayerSprite(ws *playerSprite, id *entityid) {
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

func renderWeaponSprites(s *ebiten.Image) {
	center := location{(screenWidth / 2) - centerOn.location.x - (centerOn.dimens.width / 2), (screenHeight / 2) - centerOn.location.y - (centerOn.dimens.height / 2)}
	for _, ps := range playerSprites {
		pOps := &ebiten.DrawImageOptions{}
		// offsetx := -playerSpriteHitboxExceed
		// offsety :=-(2*playerSpriteHitboxExceed)
		pOps.GeoM.Translate(float64(ps.playerRect.location.x+center.x), float64(ps.playerRect.location.y+center.y))
		s.DrawImage(ps.sprite, pOps)
	}
	for _, wep := range weapons {
		midPlayer := wep.basicSprite.playerRect.location
		midPlayer.x += wep.basicSprite.playerRect.dimens.width / 2
		midPlayer.y += wep.basicSprite.playerRect.dimens.height / 2
		midPlayer.x += center.x
		midPlayer.y += center.y
		ew, _ := wep.basicSprite.sprite.Size()
		wepOps := &ebiten.DrawImageOptions{}
		wepOps.GeoM.Translate(-float64(ew)/2, 0)
		wepOps.GeoM.Rotate(*wep.angle - (math.Pi / 2))
		wepOps.GeoM.Translate(float64(midPlayer.x), float64(midPlayer.y))
		s.DrawImage(wep.basicSprite.sprite, wepOps)
	}
}
