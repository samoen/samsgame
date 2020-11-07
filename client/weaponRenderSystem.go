package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
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
}

func (r rectangle) rectCenterPoint() location {
	x := r.location.x + (r.dimens.width / 2)
	y := r.location.y + (r.dimens.height / 2)
	return location{x, y}
}

type tileType int

const (
	blank tileType = iota
	rocky
	offworld
)

//func drawBgNew(screen *ebiten.Image){
//		for _, im := range tilesToDraw {
//				if err := screen.DrawImage(im.sprite, im.bOps); err != nil {
//					log.Fatal(err)
//				}
//		}
//}

func drawBufferedTiles(screen *ebiten.Image) {
	ops := &ebiten.DrawImageOptions{}
	//tiledraw(ops,0,0,tileRenderBuffer)
	if err := screen.DrawImage(tileRenderBuffer, ops); err != nil {
		log.Fatal(err)
	}
}

func fullRenderOp(im *baseSprite, loc location, flip bool, scaleto dimens, rot float64, rotoffset int) {
	im.bOps.GeoM.Reset()
	im.bOps.GeoM.Translate(float64(loc.x), float64(loc.y))
	im.bOps.GeoM.Translate(float64(-mycenterpoint.x), float64(-mycenterpoint.y))
	scaleToDimension(scaleto, im.sprite, im.bOps, flip)
	im.bOps.GeoM.Scale(
		math.Pow(1.01, zoom),
		math.Pow(1.01, zoom),
	)
	im.bOps.GeoM.Translate(float64(screenWidth)/2, float64(screenHeight)/2)


	//im.bOps.GeoM.Reset()
	////im.bOps.GeoM.Translate(-float64(rotoffset), 0)
	////im.bOps.GeoM.Rotate(rot)
	//im.bOps.GeoM.Translate(float64(screenWidth)/2, float64(screenHeight)/2)
	//ownerCenterx := loc.x
	////- (scaleto.width/2)
	//ownerCentery := loc.y
	////- (scaleto.height/2)
	//im.bOps.GeoM.Translate(float64(ownerCenterx), float64(ownerCentery))
	//im.bOps.GeoM.Translate(float64(-mycenterpoint.x), float64(-mycenterpoint.y))
	//
	//scaleToDimension(scaleto, im.sprite, im.bOps, flip)
	//im.bOps.GeoM.Scale(
	//	math.Pow(1.01, float64(zoom)),
	//	math.Pow(1.01, float64(zoom)),
	//)
}

func bufferTiles() {
	tileRenderBuffer.Clear()
	myCoordx := mycenterpoint.x / bgTileWidth
	myCoordy := mycenterpoint.y / bgTileWidth
	//numsee := int(2+(zoom*8))
	//numsee := 10
	//if zoom < 0 {
	correctedZoom := zoom
	if correctedZoom > 0 {
		correctedZoom = 1 / correctedZoom
	}
	correctedZoom *= -1
	//numsee := int((math.Sqrt((math.Abs(correctedZoom)+3)*1)*5)+1)
	//numsee := int((correctedZoom+10)*0.7)+ int(math.Sqrt(correctedZoom+80))
	numsee := int(correctedZoom/2.7) + 20
	//}
	log.Println("numsee: ", numsee)
	for i := myCoordx - numsee; i < myCoordx+numsee; i++ {
		for j := myCoordy - numsee; j < myCoordy+numsee; j++ {
			if im, ok := bgtilesNew[location{i, j}]; ok {
				fullRenderOp(im, location{i * bgTileWidth, j * bgTileWidth}, false, dimens{bgTileWidth, bgTileWidth}, 0, 0)
				if err := tileRenderBuffer.DrawImage(im.sprite, im.bOps); err != nil {
					log.Fatal(err)
				}
			}
			//else{
			//	ops := &ebiten.DrawImageOptions{}
			//	//tiledraw(ops,i,j,images.tile1)
			//	bs := &baseSprite{images.tile1,ops,0}
			//	fullRenderOp(bs,location{i*bgTileWidth,j*bgTileWidth},false,dimens{bgTileWidth,bgTileWidth})
			//	if err:= tileRenderBuffer.DrawImage(bs.sprite,bs.bOps); err != nil{
			//		log.Fatal(err)
			//	}
			//}
		}
	}
}

var tilesToDraw []*baseSprite
var tileRenderBuffer *ebiten.Image

//func updateTilesNew(){
//	tilesToDraw = nil
//	myCoordx := mycenterpoint.x / bgTileWidth
//	myCoordy := mycenterpoint.y / bgTileWidth
//	numsee := int(2+(zoom*bgTileWidth/1.5))
//
//	for i := myCoordx-numsee; i < myCoordx+numsee; i++ {
//		for j := myCoordy-numsee; j < myCoordy+numsee; j++ {
//			if im, ok := bgtilesNew[location{i, j}]; ok {
//				tiledraw(im.bOps,i,j,im.sprite)
//				tilesToDraw = append(tilesToDraw,im)
//			}
//		}
//	}
//}
//
//func drawBackground(screen *ebiten.Image) {
//
//	select {
//	case bgl := <-bgchan:
//		ttmap[bgl.tyti] = bgl.imim
//	default:
//	}
//
//	myCoordx := mycenterpoint.x / bgTileWidth
//	myCoordy := mycenterpoint.y / bgTileWidth
//	//remx := mycenterpoint.x % bgTileWidth
//	//remy := mycenterpoint.y % bgTileWidth
//	numsee := int(2+(zoom*9))
//	//if remx < bgTileWidth/2 {
//	//	upx = -1
//	//}
//	//upy := 0
//	//if remy < bgTileWidth/2 {
//	//	upy = -1
//	//}
//
//	for i := myCoordx-numsee; i < myCoordx+numsee; i++ {
//		for j := myCoordy-numsee; j < myCoordy+numsee; j++ {
//			if im, ok := bgtiles[location{i, j}]; ok {
//				if ttim, ok := ttmap[im.tiletyp]; ok {
//					if ttim != nil {
//						handleBgtile(i,j,screen)
//					}
//				}
//			}
//		}
//	}
//
//}
//
//func handleBgtile(i int, j int, screen *ebiten.Image) {
//	if ti, ok := bgtiles[location{i, j}]; ok {
//		prett := ti.tiletyp
//
//		if _, ok := ttmap[prett]; !ok {
//			iwl := images.sword
//			ttmap[prett] = iwl
//			go func() {
//				imstring := "assets"
//				switch prett {
//				case blank:
//					imstring = imstring + "/floor.png"
//				case rocky:
//					imstring = imstring + "/tile31.png"
//				case offworld:
//					imstring = imstring + "/8000paint.png"
//				default:
//					imstring = imstring + "/sword.png"
//				}
//
//				im, _, err := ebitenutil.NewImageFromFile(imstring, ebiten.FilterDefault)
//				//time.Sleep(500*time.Millisecond)
//				if err != nil {
//					panic(err)
//				}
//				bgl := ttwithIm{}
//				bgl.imim = im
//				bgl.tyti = prett
//				bgchan <- bgl
//			}()
//		}
//	}
//
//	if im, ok := bgtiles[location{i, j}]; ok {
//		if ttim, ok := ttmap[im.tiletyp]; ok {
//			if ttim != nil {
//				tiledraw(im.ops,i,j,ttim)
//				if err := screen.DrawImage(ttim, im.ops); err != nil {
//					log.Fatal(err)
//				}
//			}
//		}
//	}
//
//	//else{
//	//
//	//	tiledraw(&ebiten.DrawImageOptions{},i,j,screen,images.empty)
//	//}
//}

//func tiledraw(ops *ebiten.DrawImageOptions, i,j int, tileim *ebiten.Image){
//	ops.GeoM.Reset()
//	ops.GeoM.Translate(float64(offset.x), float64(offset.y))
//	scaleToDimension(dimens{bgTileWidth, bgTileWidth}, tileim, ops,false)
//	ops.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))
//
//	//xdiff := (float64(i)*float64(bgTileWidth))+(float64(bgTileWidth)/2) - float64(myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint().x)
//	//ydiff := (float64(j)*float64(bgTileWidth))+(float64(bgTileWidth)/2) - float64(myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint().y)
//	//ydiff = float64(ydiff) / zoom
//	//xdiff = float64(xdiff) / zoom
//	//ops.GeoM.Translate(float64(xdiff)*(1.0-zoom),float64(ydiff)*(1.0-zoom))
//
//}

type ttwithIm struct {
	tyti tileType
	imim *ebiten.Image
}

func bgShapesWork() {
	myCoordx := mycenterpoint.x / bgTileWidth
	myCoordy := mycenterpoint.y / bgTileWidth

	remx := mycenterpoint.x % bgTileWidth
	remy := mycenterpoint.y % bgTileWidth
	upx := -1
	if remx < bgTileWidth/2 {
		upx = 1
	}
	upy := -1
	if remy < bgTileWidth/2 {
		upy = 1
	}

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if upx == i || upy == j {
				delete(currentTShapes, location{myCoordx + i, myCoordy + j})
			} else {
				addTshape(myCoordx+i, myCoordy+j)
			}
		}
	}
}

func addTshape(i, j int) {
	loc := location{i, j}
	if ti, ok := bgtiles[loc]; ok {
		transcoord := ttshapes[ti.tiletyp]
		var newlines []line
		for _, l := range transcoord.lines {
			newlines = append(
				newlines,
				line{
					location{
						l.p1.x + (loc.x * bgTileWidth),
						l.p1.y + (loc.y * bgTileWidth),
					},
					location{
						l.p2.x + (loc.x * bgTileWidth),
						l.p2.y + (loc.y * bgTileWidth),
					},
				},
			)
		}
		transcoord.lines = newlines

		//log.Println("put bgshape ", transcoord)
		currentTShapes[location{i, j}] = transcoord
	}
}

type bgLoading struct {
	ops     *ebiten.DrawImageOptions
	tiletyp tileType
}

func scaleToDimension(dims dimens, img *ebiten.Image, ops *ebiten.DrawImageOptions, flip bool) {
	imW, imH := img.Size()
	dh := float64(dims.height + 1)
	dw := float64(dims.width + 1)
	wRatio := float64(dw) / float64(imW)
	hRatio := float64(dh) / float64(imH)

	toAdd := ebiten.GeoM{}
	if flip {
		toAdd.Scale(-wRatio, hRatio)
		toAdd.Translate(float64(dw), 0)
	} else {
		toAdd.Scale(wRatio, hRatio)
	}
	ops.GeoM.Add(toAdd)
}

func cameraShift(loc location, ops *ebiten.DrawImageOptions) {
	pSpriteOffset := offset
	pSpriteOffset.x += loc.x
	pSpriteOffset.y += loc.y
	addOp := ebiten.GeoM{}
	addOp.Translate(float64(pSpriteOffset.x), float64(pSpriteOffset.y))
	ops.GeoM.Add(addOp)
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

		bs.wepsprit.bOps.GeoM.Reset()

		wepSprightLen := bs.pivShape.bladeLength + bs.pivShape.bladeLength/4

		bs.wepsprit.bOps.GeoM.Translate(-float64(wepSprightLen/2)/2, 0)

		scaleto := dimens{int(wepSprightLen / 2), int(wepSprightLen)}
		scaleToDimension(scaleto, bs.wepsprit.sprite, bs.wepsprit.bOps, false)
		bs.wepsprit.bOps.GeoM.Scale(
			math.Pow(1.01, zoom),
			math.Pow(1.01, zoom),
		)

		rot := bs.pivShape.animationCount - (math.Pi / 2)
		bs.wepsprit.bOps.GeoM.Rotate(rot)

		bs.wepsprit.bOps.GeoM.Translate(float64(screenWidth)/2, float64(screenHeight)/2)
		ownerCenter := bs.rect.rectCenterPoint()
		bs.wepsprit.bOps.GeoM.Translate(float64(ownerCenter.x), float64(ownerCenter.y))
		bs.wepsprit.bOps.GeoM.Translate(float64(-mycenterpoint.x), float64(-mycenterpoint.y))

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
