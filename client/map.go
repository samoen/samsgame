package main

import (
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/image/colornames"
	"log"
	"math"
)

type backgroundTile struct {
	baseSprite
	passable bool
}

func drawBufferedTiles(screen *ebiten.Image) {
	ops := &ebiten.DrawImageOptions{}

	if err := screen.DrawImage(tileRenderBuffer, ops); err != nil {
		log.Fatal(err)
	}
}

func placeMap() {
	worldBoundRect := rectangle{}
	worldBoundRect.dimens = dimens{worldWidth, worldWidth}
	worldBoundRect.refreshShape(location{0, 0})
	wepBlockers[&worldBoundRect.shape] = true
}

func bufferTiles() {
	tileRenderBuffer.Clear()
	myCoordx := mycenterpoint.x / bgTileWidth
	myCoordy := mycenterpoint.y / bgTileWidth
	correctedZoom := 1 / math.Pow(1.01, zoom)
	numsee := int((23)*correctedZoom) + 2

	if zoom < -50 {
		tileRenderBuffer.Fill(colornames.Blue)

		ops := &ebiten.DrawImageOptions{}
		bs := &baseSprite{images.bigBackground, ops, 0}
		fullRenderOp(bs, location{0, 0}, false, dimens{worldWidth, worldWidth}, 0, 0)
		if err := tileRenderBuffer.DrawImage(bs.sprite, bs.bOps); err != nil {
			log.Fatal(err)
		}
		return
	}

	for i := myCoordx - numsee; i < myCoordx+numsee; i++ {
		for j := myCoordy - numsee; j < myCoordy+numsee; j++ {
			if im, ok := bgtilesNew[location{i, j}]; ok {
				fullRenderOp(&im.baseSprite, location{i * bgTileWidth, j * bgTileWidth}, false, dimens{bgTileWidth + 1, bgTileWidth + 1}, 0, 0)
				if err := tileRenderBuffer.DrawImage(im.sprite, im.bOps); err != nil {
					log.Fatal(err)
				}
			} else {
				ops := &ebiten.DrawImageOptions{}
				bs := &baseSprite{images.water, ops, 0}
				fullRenderOp(bs, location{i * bgTileWidth, j * bgTileWidth}, false, dimens{bgTileWidth + 1, bgTileWidth + 1}, 0, 0)
				if err := tileRenderBuffer.DrawImage(bs.sprite, bs.bOps); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func bgShapesWork() {
	myCoordx := mycenterpoint.x / bgTileWidth
	myCoordy := mycenterpoint.y / bgTileWidth

	currentTShapes = make(map[location]shape)
	for i := -3; i <= 3; i++ {
		for j := -3; j <= 3; j++ {
			if v, ok := bgtilesNew[location{myCoordx + i, myCoordy + j}]; ok {
				if !v.passable {
					impassShapeX := myCoordx + i
					impassShapeY := myCoordy + j
					r := rectangle{}
					r.dimens = dimens{bgTileWidth, bgTileWidth}
					r.refreshShape(location{impassShapeX * bgTileWidth, impassShapeY * bgTileWidth})
					currentTShapes[location{i, j}] = r.shape
				}
			}
		}
	}
}
