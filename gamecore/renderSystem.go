package gamecore

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten"
)

func drawHitboxes(s *ebiten.Image) {
	samDrawLine := func(l line) {
		op := ebiten.DrawImageOptions{}
		op.ColorM.Scale(0, 230, 64, 1)
		l.p1.x += offset.x
		l.p1.y += offset.y
		l.p2.x += offset.x
		l.p2.y += offset.y

		x1 := float64(l.p1.x)
		x2 := float64(l.p2.x)
		y1 := float64(l.p1.y)
		y2 := float64(l.p2.y)

		imgToDraw := *images.empty
		ew, eh := imgToDraw.Size()
		length := math.Hypot(x2-x1, y2-y1)

		op.GeoM.Scale(
			length/float64(ew),
			2/float64(eh),
		)
		op.GeoM.Rotate(math.Atan2(y2-y1, x2-x1))
		op.GeoM.Translate(x1, y1)
		if err := s.DrawImage(&imgToDraw, &op); err != nil {
			log.Fatal(err)
		}
	}

	for _, shape := range wepBlockers {
		for _, l := range shape.lines {
			samDrawLine(l)
		}
	}
	for _, slshr := range slashers {
		for _, l := range slshr.ent.rect.shape.lines {
			samDrawLine(l)
		}
		if slshr.swangin {
			for _, l := range slshr.pivShape.pivoterShape.lines {
				samDrawLine(l)
			}
		}
	}
	for _, slshr := range remotePlayers {
		for _, l := range slshr.ent.rect.shape.lines {
			samDrawLine(l)
		}
		if slshr.swangin {
			for _, l := range slshr.pivShape.pivoterShape.lines {
				samDrawLine(l)
			}
		}
	}
}
