package main

import (
	"math"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type weaponSprite struct {
	weaponShape *shape
	angle       *float64
	basicSprite playerSprite
	// drawPoint   *rectangle
	// sprite      *ebiten.Image
}

type playerSprite struct {
	playerRect *rectangle
	sprite     *ebiten.Image
}

var playerSprites []*playerSprite

func addPlayerSprite(ws *playerSprite) {
	playerSprites = append(playerSprites, ws)
	ws.playerRect.shape.removals = append(ws.playerRect.shape.removals, func() {
		for i, renderable := range playerSprites {
			if ws.playerRect.shape == renderable.playerRect.shape {
				if i < len(playerSprites)-1 {
					copy(playerSprites[i:], playerSprites[i+1:])
				}
				playerSprites[len(playerSprites)-1] = nil
				playerSprites = playerSprites[:len(playerSprites)-1]
				break
			}
		}
	})
}

type weaponRenderSystem struct {
	weapons []*weaponSprite
}

var weaponRenderingSystem = weaponRenderSystem{}
var swordImage, _, _ = ebitenutil.NewImageFromFile("assets/sword.png", ebiten.FilterDefault)
var playerStandImage, _, _ = ebitenutil.NewImageFromFile("assets/playerstand.png", ebiten.FilterDefault)
var playerSpriteHitboxExceed = 10

func (w *weaponRenderSystem) addWeaponSprite(s *weaponSprite) {
	w.weapons = append(w.weapons, s)
	s.weaponShape.removals = append(s.weaponShape.removals, func() {
		w.removeWeaponSprite(s.weaponShape)
	})
}

func (w *weaponRenderSystem) removeWeaponSprite(s *shape) {
	for i, renderable := range w.weapons {
		if s == renderable.weaponShape {
			if i < len(w.weapons)-1 {
				copy(w.weapons[i:], w.weapons[i+1:])
			}
			w.weapons[len(w.weapons)-1] = nil
			w.weapons = w.weapons[:len(w.weapons)-1]
			break
		}
	}
}

var centerOn *rectangle

func (w *weaponRenderSystem) work(s *ebiten.Image) {
	center := location{(screenWidth / 2) - centerOn.location.x - (centerOn.dimens.width / 2), (screenHeight / 2) - centerOn.location.y - (centerOn.dimens.height / 2)}
	for _, ps := range playerSprites {
		pOps := &ebiten.DrawImageOptions{}
		// offsetx := -playerSpriteHitboxExceed
		// offsety :=-(2*playerSpriteHitboxExceed)
		pOps.GeoM.Translate(float64(ps.playerRect.location.x+center.x), float64(ps.playerRect.location.y+center.y))
		s.DrawImage(ps.sprite, pOps)
	}
	for _, wep := range w.weapons {
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
