package gamecore

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

var images imagesStruct

type imagesStruct struct {
	playerStand         *ebiten.Image
	playerWalkUp        *ebiten.Image
	playerWalkDown      *ebiten.Image
	playerWalkDownAngle *ebiten.Image
	playerWalkUpAngle   *ebiten.Image
	playerSwing         *ebiten.Image
	empty               *ebiten.Image
	sword               *ebiten.Image
}

var assetsDir = "assets"

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
}

var mycenterpoint location

func rectCenterPoint(r rectangle) location {
	x := r.location.x + (r.dimens.width / 2)
	y := r.location.y + (r.dimens.height / 2)
	return location{x, y}
}

var bgchan = make(chan ttwithIm)
var bgtiles = make(map[location]*bgLoading)
var ttmap = make(map[tileType]*ebiten.Image)
var ttshapes = make(map[tileType]shape)
var currentTShapes = make(map[location]shape)

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
	remx := mycenterpoint.x % bgTileWidth
	remy := mycenterpoint.y % bgTileWidth
	upx := 0
	if remx < bgTileWidth/2 {
		upx = -1
	}
	upy := 0
	if remy < bgTileWidth/2 {
		upy = -1
	}

	for i := upx; i < 2+upx; i++ {
		for j := upy; j < 2+upy; j++ {
			handleBgtile(myCoordx+i, myCoordy+j, screen)
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
				var imstring string
				switch prett {
				case blank:
					imstring = assetsDir + "/floor.png"
				case rocky:
					imstring = assetsDir + "/tile31.png"
				case offworld:
					imstring = assetsDir + "/8000paint.png"
				default:
					imstring = assetsDir + "/sword.png"
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
				im.ops.GeoM.Reset()
				im.ops.GeoM.Translate(float64(offset.x), float64(offset.y))
				scaleToDimension(dimens{bgTileWidth, bgTileWidth}, ttim, im.ops)
				im.ops.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))
				if err := screen.DrawImage(ttim, im.ops); err != nil {
					log.Fatal(err)
				}
			}
		}
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

func scaleToDimension(dims dimens, img *ebiten.Image, ops *ebiten.DrawImageOptions) {
	imW, imH := img.Size()
	wRatio := float64(dims.width) / float64(imW)
	hRatio := float64(dims.height) / float64(imH)
	toAdd := ebiten.GeoM{}
	toAdd.Scale(wRatio, hRatio)
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

	for shape := range wepBlockers {
		for _, l := range shape.lines {
			l.samDrawLine(s)
		}
	}
	for slshr := range slashers {
		for _, l := range slshr.locEnt.lSlasher.ent.rect.shape.lines {
			l.samDrawLine(s)
		}
		if slshr.locEnt.lSlasher.swangin {
			for _, l := range slshr.locEnt.lSlasher.pivShape.pivoterShape.lines {
				l.samDrawLine(s)
			}
		}
	}
	for _, l := range myLocalPlayer.locEnt.lSlasher.ent.rect.shape.lines {
		l.samDrawLine(s)
	}
	if myLocalPlayer.locEnt.lSlasher.swangin {
		for _, l := range myLocalPlayer.locEnt.lSlasher.pivShape.pivoterShape.lines {
			l.samDrawLine(s)
		}
	}
	for _, slshr := range remotePlayers {
		for _, l := range slshr.rSlasher.ent.rect.shape.lines {
			l.samDrawLine(s)
		}
		if slshr.rSlasher.swangin {
			for _, l := range slshr.rSlasher.pivShape.pivoterShape.lines {
				l.samDrawLine(s)
			}
		}
	}
}

func updateSprites() {
	toRender = nil

	for bs := range slashers {
		bs.locEnt.lSlasher.updateSlasherSprite()

	}
	for _, bs := range remotePlayers {
		bs.rSlasher.updateSlasherSprite()

	}
	if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP > 0 {
		myLocalPlayer.locEnt.lSlasher.updateSlasherSprite()
	}
	sort.Slice(toRender, func(i, j int) bool {
		return toRender[i].yaxis < toRender[j].yaxis
	})
}

func (bs *slasher) updateSlasherSprite() {
	bs.bsprit.bOps.GeoM.Reset()
	bs.bsprit.bOps.ColorM.Reset()
	toRender = append(toRender, bs.bsprit)

	bs.hbarsprit.bOps.GeoM.Reset()
	bs.hbarsprit.bOps.ColorM.Reset()
	toRender = append(toRender, bs.hbarsprit)

	if bs.swangin {
		bs.wepsprit.bOps.GeoM.Reset()
		bs.wepsprit.bOps.ColorM.Reset()
		toRender = append(toRender, bs.wepsprit)
	}
	if bs.deth.redScale > 0 {
		bs.deth.redScale--
	}
	bs.bsprit.bOps.ColorM.Translate(float64(bs.deth.redScale), 0, 0, 0)
	bs.hbarsprit.yaxis = rectCenterPoint(bs.ent.rect).y + 10
	healthbarlocation := location{bs.ent.rect.location.x, bs.ent.rect.location.y - (bs.ent.rect.dimens.height / 2) - 10}
	healthbardimenswidth := bs.deth.hp.CurrentHP * bs.ent.rect.dimens.width / bs.deth.hp.MaxHP
	scaleToDimension(dimens{healthbardimenswidth, 5}, images.empty, bs.hbarsprit.bOps)
	cameraShift(healthbarlocation, bs.hbarsprit.bOps)

	bs.bsprit.yaxis = rectCenterPoint(bs.ent.rect).y

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

	intverted := 1
	if math.Abs(bs.startangle) > math.Pi/2 {
		intverted = -1
		flipTrans := ebiten.GeoM{}
		flipTrans.Translate(float64(-bs.ent.rect.dimens.width-(bs.ent.rect.dimens.width/2)), 0)
		bs.bsprit.bOps.GeoM.Add(flipTrans)
		bs.bsprit.bOps.GeoM.Scale(-1, 1)
	}
	scaleto := dimens{}

	scaleto.width = bs.ent.rect.dimens.width
	scaleto.width += (bs.ent.rect.dimens.width / 2) * intverted

	scaleto.height = bs.ent.rect.dimens.height
	scaleto.height += bs.ent.rect.dimens.height / 2

	shiftto := location{}
	shiftto.x = bs.ent.rect.location.x
	shiftto.x -= bs.ent.rect.dimens.width / 4
	shiftto.y = bs.ent.rect.location.y
	shiftto.y -= bs.ent.rect.dimens.height / 2

	scaleToDimension(scaleto, bs.bsprit.sprite, bs.bsprit.bOps)
	cameraShift(shiftto, bs.bsprit.bOps)

	if bs.swangin {
		_, imH := bs.wepsprit.sprite.Size()
		bs.wepsprit.yaxis = bs.pivShape.pivoterShape.lines[0].p2.y
		ownerCenter := rectCenterPoint(bs.ent.rect)
		cameraShift(ownerCenter, bs.wepsprit.bOps)
		addOp := ebiten.GeoM{}
		hRatio := float64(bs.pivShape.bladeLength+bs.pivShape.bladeLength/4) / float64(imH)
		addOp.Scale(hRatio, hRatio)
		addOp.Translate(-float64(bs.ent.rect.dimens.width)/2, 0)
		addOp.Rotate(bs.pivShape.animationCount - (math.Pi / 2))
		bs.wepsprit.bOps.GeoM.Add(addOp)
	}
}
