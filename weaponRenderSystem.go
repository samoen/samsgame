package main

import (
	"math"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type weaponSprite struct {
	weaponShape *shape
	angle       *float64
}

type weaponRenderSystem struct {
	weapons  []*weaponSprite
	CenterOn *rectangle
}

var weaponRenderingSystem = weaponRenderSystem{}
var swordImage, _, _ = ebitenutil.NewImageFromFile("assets/sword.png", ebiten.FilterDefault)

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

func (w *weaponRenderSystem) work(s *ebiten.Image) {
	center := location{(screenWidth / 2) - w.CenterOn.location.x - (w.CenterOn.dimens.width / 2), (screenHeight / 2) - w.CenterOn.location.y - (w.CenterOn.dimens.height / 2)}
	for _, wep := range w.weapons {
		wepOps := &ebiten.DrawImageOptions{}
		point := wep.weaponShape.lines[0]
		point.p1.x += center.x
		point.p1.y += center.y
		wepOps.GeoM.Rotate(*wep.angle - (math.Pi / 2) + 0.2)
		wepOps.GeoM.Translate(float64(point.p1.x), float64(point.p1.y))
		s.DrawImage(swordImage, wepOps)
	}
}
