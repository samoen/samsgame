package gamecore

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten"
)

var hitBoxes = make(map[*entityid]*shape)

func addHitbox(s *shape, id *entityid) {
	hitBoxes[id] = s
	id.systems = append(id.systems, hitBoxRenderable)
}

func drawHitboxes(s *ebiten.Image) {
	center := renderOffset()
	samDrawLine := func(l line) {
		op := ebiten.DrawImageOptions{}
		op.ColorM.Scale(0, 230, 64, 1)
		l.p1.x += center.x
		l.p1.y += center.y
		l.p2.x += center.x
		l.p2.y += center.y

		x1 := float64(l.p1.x)
		x2 := float64(l.p2.x)
		y1 := float64(l.p1.y)
		y2 := float64(l.p2.y)

		imgToDraw := *emptyImage
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

	for _, shape := range hitBoxes {
		for _, l := range shape.lines {
			samDrawLine(l)
		}
	}
}
