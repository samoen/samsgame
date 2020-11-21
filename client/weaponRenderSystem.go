package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"golang.org/x/image/colornames"
	"log"
	"math"
	"sort"
)

type baseSprite struct {
	sprite *ebiten.Image
	bOps   *ebiten.DrawImageOptions
	yaxis  int
}

type imagesStruct struct {
	playerStand         *ebiten.Image
	playerWalkUp        *ebiten.Image
	playerWalkDown      *ebiten.Image
	playerWalkDownAngle *ebiten.Image
	playerWalkUpAngle   *ebiten.Image
	playerSwing         *ebiten.Image
	playerfall0         *ebiten.Image
	playerfall1         *ebiten.Image
	playerfall2         *ebiten.Image
	empty               *ebiten.Image
	sword               *ebiten.Image
	tile1               *ebiten.Image
	tile2               *ebiten.Image
	bigBackground       *ebiten.Image
	water       *ebiten.Image
}

func cacheImage(name string) (img *ebiten.Image) {
	img, _, err := ebitenutil.NewImageFromFile(
		"assets/"+name+".png",
		ebiten.FilterDefault,
	)
	if err != nil {
		panic(err)
	}
	return img
}

func (is *imagesStruct) newImages() {
	is.playerStand = cacheImage("playerstand")
	is.playerWalkUp = cacheImage("playerwalkup")
	is.playerWalkDown = cacheImage("playerwalkdown")
	is.playerWalkDownAngle = cacheImage("playerwalkdownangle")
	is.playerWalkUpAngle = cacheImage("playerwalkupangle")
	is.playerSwing = cacheImage("playerswing1")
	is.empty = cacheImage("floor")
	is.sword = cacheImage("axe")
	is.playerfall0 = cacheImage("man1")
	is.playerfall1 = cacheImage("man2")
	is.playerfall2 = cacheImage("man3")
	is.tile1 = cacheImage("tile31")
	is.tile2 = cacheImage("tile2")
	is.bigBackground = cacheImage("bigbg")
	is.water = cacheImage("water")
}

func (r rectangle) rectCenterPoint() location {
	x := r.location.x + (r.dimens.width / 2)
	y := r.location.y + (r.dimens.height / 2)
	return location{x, y}
}

func drawBufferedTiles(screen *ebiten.Image) {
	ops := &ebiten.DrawImageOptions{}

	if err := screen.DrawImage(tileRenderBuffer, ops); err != nil {
		log.Fatal(err)
	}
}

func fullRenderOp(im *baseSprite, loc location, flip bool, scaleto dimens, rot float64, rotOffsetx float64) {
	im.bOps.GeoM.Reset()

	im.bOps.GeoM.Translate(float64(loc.x), float64(loc.y))
	im.bOps.GeoM.Translate(float64(-mycenterpoint.x), float64(-mycenterpoint.y))

	scaleToDimension(scaleto, im.sprite, im.bOps, flip)
	zoomScale := math.Pow(1.01, zoom)
	im.bOps.GeoM.Scale(
		float64(zoomScale),
		float64(zoomScale),
	)

	im.bOps.GeoM.Translate(float64(int(float64(screenWidth)/2)), float64(int(float64(screenHeight)/2)))

	tx := im.bOps.GeoM.Element(0, 2)
	ty := im.bOps.GeoM.Element(1, 2)
	im.bOps.GeoM.Translate(-tx, -ty)
	im.bOps.GeoM.Translate(-zoomScale*rotOffsetx, 0)
	im.bOps.GeoM.Rotate(rot)
	im.bOps.GeoM.Translate(tx, ty)
}

func bufferTiles() {
	tileRenderBuffer.Clear()
	myCoordx := mycenterpoint.x / bgTileWidth
	myCoordy := mycenterpoint.y / bgTileWidth
	correctedZoom := 1 / math.Pow(1.01, zoom)
	numsee := int((23)*correctedZoom) + 2

	if zoom < -50{
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

func scaleToDimension(dims dimens, img *ebiten.Image, ops *ebiten.DrawImageOptions, flip bool) {
	imW, imH := img.Size()
	dh := float64(dims.height)
	dw := float64(dims.width)
	wRatio := float64(dw) / float64(imW)
	hRatio := float64(dh) / float64(imH)

	toAdd := ebiten.GeoM{}

	if flip {
		toAdd.Scale(-float64(wRatio), float64(hRatio))
		toAdd.Translate(float64(int(dw)), 0)
	} else {
		toAdd.Scale(wRatio, hRatio)
	}
	ops.GeoM.Add(toAdd)
}

func renderEntSprites(s *ebiten.Image) {
	for _, bs := range toRender {
		if err := s.DrawImage(bs.sprite, bs.bOps); err != nil {
			log.Fatal(err)
		}
	}
}

func drawHitboxes(s *ebiten.Image) {

	for _, shape := range currentTShapes {
		for _, l := range shape.lines {
			l.samDrawLine(s)
		}
	}

	for shape, _ := range wepBlockers {
		for _, l := range shape.lines {
			l.samDrawLine(s)
		}
	}

	for slshr, _ := range localAnimals {
		slshr.locEnt.lSlasher.hitbox(s)
	}
	if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP > 0 {
		myLocalPlayer.locEnt.lSlasher.hitbox(s)
	}

	for _, slshr := range remotePlayers {
		slshr.rSlasher.hitbox(s)
	}
}

func updateSprites() {
	toRender = nil

	for bs, _ := range localAnimals {
		bs.locEnt.lSlasher.updateSlasherSprite()

	}
	for _, bs := range remotePlayers {
		bs.rSlasher.updateSlasherSprite()

	}
	if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP > 0 {
		myLocalPlayer.locEnt.lSlasher.updateSlasherSprite()
	}

	for bs, _ := range deathAnimations {
		bs.animcount++
		framesperswitch := 10
		animframe := int(math.Floor(float64(bs.animcount) / float64(framesperswitch)))
		if animframe >= len(bs.sprites) {
			delete(deathAnimations, bs)
			continue
		}
		toupdate := bs.sprites[animframe]
		fullRenderOp(&toupdate, bs.rect.location, bs.inverted, bs.rect.dimens, 0, 0)
		toRender = append(toRender, toupdate)
	}

	sort.Slice(toRender, func(i, j int) bool {
		return toRender[i].yaxis < toRender[j].yaxis
	})
}

func (bs *slasher) updateSlasherSprite() {
	bs.bsprit.bOps.GeoM.Reset()
	bs.bsprit.bOps.ColorM.Reset()

	bs.hbarsprit.bOps.GeoM.Reset()
	bs.hbarsprit.bOps.ColorM.Reset()

	if bs.swangin {
		bs.wepsprit.bOps.GeoM.Reset()
		bs.wepsprit.bOps.ColorM.Reset()
	}
	if bs.deth.redScale > 0 {
		bs.deth.redScale--
	}
	bs.bsprit.bOps.ColorM.Translate(float64(bs.deth.redScale), 0, 0, 0)
	bs.hbarsprit.yaxis = bs.rect.rectCenterPoint().y + 10
	healthbarlocation := location{bs.rect.location.x, bs.rect.location.y - (bs.rect.dimens.height / 2) - 10}
	healthbardimenswidth := bs.deth.hp.CurrentHP * bs.rect.dimens.width / bs.deth.hp.MaxHP
	fullRenderOp(&bs.hbarsprit, healthbarlocation, false, dimens{healthbardimenswidth, 5}, 0, 0)
	//scaleToDimension(dimens{healthbardimenswidth, 5}, images.empty, bs.hbarsprit.bOps,false)
	//cameraShift(healthbarlocation, bs.hbarsprit.bOps)

	bs.bsprit.yaxis = bs.rect.rectCenterPoint().y

	spriteSelect := images.empty
	tolerance := math.Pi / 9
	if bs.swangin {
		spriteSelect = images.playerSwing
	} else if math.Abs(bs.startangle) < tolerance {
		spriteSelect = images.playerStand
	} else if math.Abs(bs.startangle-(math.Pi/4)) < tolerance {
		spriteSelect = images.playerWalkDownAngle
	} else if math.Abs(bs.startangle-(math.Pi/2)) < tolerance {
		spriteSelect = images.playerWalkDown
	} else if math.Abs(bs.startangle-(3*math.Pi/4)) < tolerance {
		spriteSelect = images.playerWalkDownAngle
	} else if math.Abs(bs.startangle-(-3*math.Pi/4)) < tolerance {
		spriteSelect = images.playerWalkUpAngle
	} else if math.Abs(bs.startangle-(-math.Pi/2)) < tolerance {
		spriteSelect = images.playerWalkUp
	} else if math.Abs(bs.startangle-(-math.Pi/4)) < tolerance {
		spriteSelect = images.playerWalkUpAngle
	} else if math.Abs(bs.startangle)-math.Pi < tolerance {
		spriteSelect = images.playerStand
	}

	bs.bsprit.sprite = spriteSelect
	invertbool := math.Abs(bs.startangle) > math.Pi/2
	fullRenderOp(&bs.bsprit, bs.rect.location, invertbool, bs.rect.dimens, 0, 0)

	if bs.swangin {
		bs.wepsprit.yaxis = bs.pivShape.pivoterShape.lines[0].p2.y
		wepSprightLen := bs.pivShape.bladeLength + bs.pivShape.bladeLength/6
		ownerCenter := bs.rect.rectCenterPoint()
		scaleto := dimens{int(wepSprightLen / 2), int(wepSprightLen)}
		fullRenderOp(&bs.wepsprit, ownerCenter, false, scaleto, bs.pivShape.animationCount-(math.Pi/2), float64(wepSprightLen/2)/2)
		toRender = append(toRender, bs.wepsprit)
	}
	toRender = append(toRender, bs.bsprit)
	toRender = append(toRender, bs.hbarsprit)
}

func playerSpriteLargerScale(rect rectangle) dimens {
	scaleto := dimens{}
	scaleto.width = rect.dimens.width
	scaleto.width += (rect.dimens.width / 2)

	scaleto.height = rect.dimens.height
	scaleto.height += rect.dimens.height / 2
	return scaleto
}
func playerSpriteLargerShift(rect rectangle) location {
	shiftto := location{}
	shiftto.x = rect.location.x
	shiftto.x -= rect.dimens.width / 4
	shiftto.y = rect.location.y
	shiftto.y -= rect.dimens.height / 2
	return shiftto
}
