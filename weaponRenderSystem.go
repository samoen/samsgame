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
	redScale   *int
	flip       *directions
	lastflip   bool
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

func renderEntSprites(s *ebiten.Image) {
	center := renderOffset()
	for _, ps := range basicSprites {
		pOps := &ebiten.DrawImageOptions{}
		// offsetx := -playerSpriteHitboxExceed
		// offsety :=-(2*playerSpriteHitboxExceed)

		imW, imH := ps.sprite.Size()
		wRatio := float64(ps.playerRect.dimens.width) / float64(imW)
		hRatio := float64(ps.playerRect.dimens.height) / float64(imH)
		pOps.GeoM.Scale(float64(wRatio), float64(hRatio))

		if ps.flip.left && !ps.flip.right {
			ps.lastflip = true
		}
		if ps.flip.right && !ps.flip.left {
			ps.lastflip = false
		}

		if ps.lastflip {
			pOps.GeoM.Scale(-1, 1)
			pOps.GeoM.Translate(float64(ps.playerRect.dimens.width), 0)
		}

		pOps.ColorM.Translate(float64(*ps.redScale), 0, 0, 0)

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

		wepOps := &ebiten.DrawImageOptions{}

		_, imH := wep.basicSprite.sprite.Size()
		hRatio := float64(swordLength+swordLength/4) / float64(imH)
		wepOps.GeoM.Scale(float64(hRatio), float64(hRatio))

		wepOps.GeoM.Translate(-float64(wep.basicSprite.playerRect.dimens.width)/2, 0)
		wepOps.GeoM.Rotate(*wep.angle - (math.Pi / 2))
		wepOps.GeoM.Translate(float64(wepOffset.x), float64(wepOffset.y))
		s.DrawImage(wep.basicSprite.sprite, wepOps)
	}
}
