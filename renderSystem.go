package main

import (
	"bytes"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/examples/resources/images"
)

const (
	tilescreenWidth  = 240
	tilescreenHeight = 240
	tileSize         = 16
	tileXNum         = 25
	xNum             = tilescreenWidth / tileSize
)

var renderingSystem = renderSystem{}
var img, _, _ = image.Decode(bytes.NewReader(images.Tiles_png))
var tilesImage, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)
var emptyImage, _, _ = ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)

// func initRenderSystem() {
// bgImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
// bgSizex, sgsizey := bgImage.Size()
// bgOps := &ebiten.DrawImageOptions{}
// bgOps.GeoM.Scale(float64(mapBounds.dimens.width)/float64(bgSizex), float64(mapBounds.dimens.height)/float64(sgsizey))
// bgOps.GeoM.Translate(float64(screenWidth/2), float64(screenHeight/2))

// pImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
// pSizex, pSizey := pImage.Size()
// pOps := &ebiten.DrawImageOptions{}
// pOps.GeoM.Scale(float64(player.width)/float64(pSizex), float64(player.height)/float64(pSizey))
// }

type renderSystem struct {
	shapes   []*shape
	CenterOn *rectangle
}

func (r *renderSystem) addShape(s *shape) {
	r.shapes = append(r.shapes, s)
}
func (r renderSystem) work(s *ebiten.Image) {
	for _, l := range layers {
		for i, t := range l {

			tileOps := &ebiten.DrawImageOptions{}
			tileOps.GeoM.Translate(float64((i%xNum)*tileSize), float64((i/xNum)*tileSize))

			tileOps.GeoM.Translate(float64(screenWidth/2), float64(screenHeight/2))
			tileOps.GeoM.Translate(float64(-r.CenterOn.location.x), float64(-r.CenterOn.location.y))
			tileOps.GeoM.Translate(float64(-r.CenterOn.dimens.width/2), float64(-r.CenterOn.dimens.height/2))
			// tileOps.GeoM.Scale(2, 2)
			// tileOps.GeoM.Scale(float64(mapBounds.dimens.width)/float64(tileImSizex), float64(mapBounds.dimens.height)/float64(tileImSizey))

			sx := (t % tileXNum) * tileSize
			sy := (t / tileXNum) * tileSize
			subImage := tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image)
			s.DrawImage(subImage, tileOps)
		}
	}

	center := location{(screenWidth / 2) - r.CenterOn.location.x - (r.CenterOn.dimens.width / 2), (screenHeight / 2) - r.CenterOn.location.y - (r.CenterOn.dimens.height / 2)}
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
		for _, l := range *shape {
			samDrawLine(l)
		}
	}

	// newops := *bgOps
	// newops.GeoM.Translate(float64(-player.rectangle.location.x), float64(-player.rectangle.location.y))
	// newops.GeoM.Translate(float64(-player.rectangle.dimens.width/2), float64(-player.rectangle.dimens.height/2))
	// screen.DrawImage(bgImage, &newops)

	// newPOps := *pOps
	// newPOps.GeoM.Translate(float64((screenWidth/2)-(player.w/2)), float64((screenHeight/2)-(player.h/2)))
	// screen.DrawImage(pImage, &newPOps)

}
