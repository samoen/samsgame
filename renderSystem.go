package main

import (
	"math"

	"github.com/hajimehoshi/ebiten"
)

var renderingSystem = newRenderSystem()

var bgOps = &ebiten.DrawImageOptions{}

func init() {
	bgOps.GeoM.Translate(float64(screenWidth/2), float64(screenHeight/2))
}

func newRenderSystem() renderSystem {
	r := renderSystem{}
	r.shapes = make(map[*entityid]*shape)
	return r
}

type renderSystem struct {
	shapes map[*entityid]*shape
}

// func (r *renderSystem) removeShape(s *shape) {
// 	for i, renderable := range r.shapes {
// 		if s == renderable {
// 			if i < len(r.shapes)-1 {
// 				copy(r.shapes[i:], r.shapes[i+1:])
// 			}
// 			r.shapes[len(r.shapes)-1] = nil
// 			r.shapes = r.shapes[:len(r.shapes)-1]
// 		}
// 	}
// }

func (r *renderSystem) addShape(s *shape, id *entityid) {
	// r.shapes = append(r.shapes, s)
	r.shapes[id] = s
	id.systems = append(id.systems, hitBoxRenderable)
}

func (r *renderSystem) work(s *ebiten.Image) {
	center := location{(screenWidth / 2) - centerOn.location.x - (centerOn.dimens.width / 2), (screenHeight / 2) - centerOn.location.y - (centerOn.dimens.height / 2)}
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
		s.DrawImage(&imgToDraw, &op)
	}

	for _, shape := range r.shapes {
		for _, l := range shape.lines {
			samDrawLine(l)
		}
	}
}
