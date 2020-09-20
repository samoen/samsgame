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
}

func (r rectangle)rectCenterPoint() location {
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

func drawBackground(screen *ebiten.Image) {

	select {
	case bgl := <-bgchan:
		ttmap[bgl.tyti] = bgl.imim
	default:
	}

	myCoordx := mycenterpoint.x / bgTileWidth
	myCoordy := mycenterpoint.y / bgTileWidth
	//remx := mycenterpoint.x % bgTileWidth
	//remy := mycenterpoint.y % bgTileWidth
	numsee := int(2+(zoom*9))
	//if remx < bgTileWidth/2 {
	//	upx = -1
	//}
	//upy := 0
	//if remy < bgTileWidth/2 {
	//	upy = -1
	//}

	for i := myCoordx-numsee; i < myCoordx+numsee; i++ {
		for j := myCoordy-numsee; j < myCoordy+numsee; j++ {
			handleBgtile(i, j, screen)
		}
	}
}

func handleBgtile(i int, j int, screen *ebiten.Image) {
	if ti, ok := bgtiles[location{i, j}]; ok {
		prett := ti.tiletyp

		if _, ok := ttmap[prett]; !ok {
			iwl := images.sword
			ttmap[prett] = iwl
			go func() {
				imstring := "assets"
				switch prett {
				case blank:
					imstring = imstring + "/floor.png"
				case rocky:
					imstring = imstring + "/tile31.png"
				case offworld:
					imstring = imstring + "/8000paint.png"
				default:
					imstring = imstring + "/sword.png"
				}

				im, _, err := ebitenutil.NewImageFromFile(imstring, ebiten.FilterDefault)
				//time.Sleep(500*time.Millisecond)
				if err != nil {
					panic(err)
				}
				bgl := ttwithIm{}
				bgl.imim = im
				bgl.tyti = prett
				bgchan <- bgl
			}()
		}
	}

	if im, ok := bgtiles[location{i, j}]; ok {
		if ttim, ok := ttmap[im.tiletyp]; ok {
			if ttim != nil {
				tiledraw(im.ops,i,j,screen,ttim)
				//im.ops.GeoM.Reset()
				//im.ops.GeoM.Translate(float64(offset.x), float64(offset.y))
				//scaleToDimension(dimens{bgTileWidth, bgTileWidth}, ttim, im.ops,false)
				//im.ops.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))
				//
				//xdiff := (float64(i)*float64(bgTileWidth))+(float64(bgTileWidth)/2) - float64(myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint().x)
				//ydiff := (float64(j)*float64(bgTileWidth))+(float64(bgTileWidth)/2) - float64(myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint().y)
				//ydiff = float64(ydiff) / zoom
				//xdiff = float64(xdiff) / zoom
				//im.ops.GeoM.Translate(float64(xdiff)*(1.0-zoom),float64(ydiff)*(1.0-zoom))
				//
				//if err := screen.DrawImage(ttim, im.ops); err != nil {
				//	log.Fatal(err)
				//}
			}
		}
	}else{

		tiledraw(&ebiten.DrawImageOptions{},i,j,screen,images.empty)
	}
}

func tiledraw(ops *ebiten.DrawImageOptions, i,j int, screen *ebiten.Image, tileim *ebiten.Image){
	ops.GeoM.Reset()
	ops.GeoM.Translate(float64(offset.x), float64(offset.y))
	scaleToDimension(dimens{bgTileWidth, bgTileWidth}, tileim, ops,false)
	ops.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))

	xdiff := (float64(i)*float64(bgTileWidth))+(float64(bgTileWidth)/2) - float64(myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint().x)
	ydiff := (float64(j)*float64(bgTileWidth))+(float64(bgTileWidth)/2) - float64(myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint().y)
	ydiff = float64(ydiff) / zoom
	xdiff = float64(xdiff) / zoom
	ops.GeoM.Translate(float64(xdiff)*(1.0-zoom),float64(ydiff)*(1.0-zoom))

	if err := screen.DrawImage(tileim, ops); err != nil {
		log.Fatal(err)
	}
}

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
	dh := float64(dims.height)/zoom
	dw := float64(dims.width)/zoom
	wRatio := float64(dw) / float64(imW)
	hRatio := float64(dh) / float64(imH)

	toAdd := ebiten.GeoM{}
	if flip{
		toAdd.Scale(-wRatio,hRatio)
		toAdd.Translate(float64(dw),0)
	}else{
		toAdd.Scale(wRatio, hRatio)
	}
	toAdd.Translate((1-zoom)*-float64(dw)/2,(1-zoom)*-float64(dh)/2)
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
		toupdate.bOps.GeoM.Reset()

		scaleto := playerSpriteLargerScale(bs.rect)
		scaleToDimension(scaleto, toupdate.sprite, toupdate.bOps,bs.inverted)


		shiftto := playerSpriteLargerShift(bs.rect)
		cameraShift(shiftto, toupdate.bOps)

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
	scaleToDimension(dimens{healthbardimenswidth, 5}, images.empty, bs.hbarsprit.bOps,false)
	cameraShift(healthbarlocation, bs.hbarsprit.bOps)

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
	//scaleto := playerSpriteLargerScale(bs.rect)
	scaleto := bs.rect.dimens
	scaleToDimension(scaleto, bs.bsprit.sprite, bs.bsprit.bOps,invertbool)
	//shiftto := playerSpriteLargerShift(bs.rect)
	//shiftto :=
	cameraShift(bs.rect.location, bs.bsprit.bOps)

	xdiff := float64(bs.rect.rectCenterPoint().x - myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint().x)
	//xdiff = int(math.Abs(float64(xdiff)))
	xdiff = float64(xdiff) / zoom

	ydiff := float64(bs.rect.rectCenterPoint().y - myLocalPlayer.locEnt.lSlasher.rect.rectCenterPoint().y)
	//ydiff = int(math.Abs(float64(ydiff)))
	ydiff = float64(ydiff) / zoom
	bs.bsprit.bOps.GeoM.Translate(float64(xdiff)*(1-zoom),float64(ydiff)*(1-zoom))

	if bs.swangin {
		_, imH := bs.wepsprit.sprite.Size()
		bs.wepsprit.yaxis = bs.pivShape.pivoterShape.lines[0].p2.y
		ownerCenter := bs.rect.rectCenterPoint()
		cameraShift(ownerCenter, bs.wepsprit.bOps)
		addOp := ebiten.GeoM{}
		hRatio := float64(bs.pivShape.bladeLength+bs.pivShape.bladeLength/4) / float64(imH)
		addOp.Scale(hRatio, hRatio)
		addOp.Translate(-float64(bs.rect.dimens.width)/2, 0)
		addOp.Rotate(bs.pivShape.animationCount - (math.Pi / 2))
		bs.wepsprit.bOps.GeoM.Add(addOp)
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
