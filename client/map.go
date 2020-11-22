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

func chuckStuff(screen *ebiten.Image) {
	mychunkx := mycenterpoint.x / chunkWidth
	mychunky := mycenterpoint.y / chunkWidth
	correctedZoom := 1 / math.Pow(1.01, zoom)
	numsee := int(correctedZoom/4) + 1

	for i := mychunkx - numsee; i <= mychunkx+numsee; i++ {
		for j := mychunky - numsee; j <= mychunky+numsee; j++ {
			k := location{i, j}
			v, ok := mapChunks[location{i, j}]
			if !ok {
				continue
			}
			ops := &ebiten.DrawImageOptions{}
			bs := &baseSprite{v, ops, 0}
			fullRenderOp(bs, location{k.x * chunkWidth, k.y * chunkWidth}, false, dimens{chunkWidth, chunkWidth}, 0, 0)
			if err := screen.DrawImage(v, bs.bOps); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func drawBufferedTiles(screen *ebiten.Image) {
	ops := &ebiten.DrawImageOptions{}
	if zoom < -90 {
		screen.Fill(colornames.Blue)
		//images.minimap.Fill(colornames.Blue)
		//minx := mycenterpoint.x-(screenWidth/2)
		//miny := mycenterpoint.y-(screenHeight/2)
		//maxx := mycenterpoint.x+(screenWidth/2)
		//maxy := mycenterpoint.y+(screenHeight/2)
		//window := images.minimap.SubImage(image.Rect(minx/downscale,miny/downscale,maxx/downscale,maxy/downscale)).(*ebiten.Image)
		bs := &baseSprite{images.minimap, ops, 0}
		fullRenderOp(bs, location{0, 0}, false, dimens{worldWidth, worldWidth}, 0, 0)
		if err := screen.DrawImage(images.minimap, ops); err != nil {
			log.Fatal(err)
		}
		return
	}
	myCoordx := mycenterpoint.x / bgTileWidth
	myCoordy := mycenterpoint.y / bgTileWidth
	correctedZoom := 1 / math.Pow(1.01, zoom)
	numsee := int((23)*correctedZoom) + 2

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
	//tileRenderBuffer.Clear()

}

func bgShapesWork() {
	currentTShapes = make(map[location]shape)
	addProximityShapes(myLocalPlayer.locEnt.lSlasher.rect)
	for la, _ := range localAnimals {
		addProximityShapes(la.locEnt.lSlasher.rect)
	}
}

func addProximityShapes(rect rectangle) {
	myCoordx := rect.rectCenterPoint().x / bgTileWidth
	myCoordy := rect.rectCenterPoint().y / bgTileWidth

	for i := -3; i <= 3; i++ {
		for j := -3; j <= 3; j++ {
			if v, ok := bgtilesNew[location{myCoordx + i, myCoordy + j}]; ok {
				if !v.passable {
					impassShapeX := myCoordx + i
					impassShapeY := myCoordy + j
					r := rectangle{}
					r.dimens = dimens{bgTileWidth, bgTileWidth}
					r.refreshShape(location{impassShapeX * bgTileWidth, impassShapeY * bgTileWidth})
					currentTShapes[location{impassShapeX, impassShapeY}] = r.shape
				}
			}
		}
	}
}
