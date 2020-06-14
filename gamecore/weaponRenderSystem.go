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
	bg                  *ebiten.Image
}

var assetsDir = "assets"

func newImages() (imagesStruct, error) {
	playerStandImage, _, err := ebitenutil.NewImageFromFile(
		assetsDir+"/playerstand.png",
		ebiten.FilterDefault,
	)
	if err != nil {
		return imagesStruct{}, err
	}

	playerWup, _, err := ebitenutil.NewImageFromFile(
		assetsDir+"/playerwalkup.png",
		ebiten.FilterDefault,
	)
	if err != nil {
		return imagesStruct{}, err
	}
	playerWdown, _, err := ebitenutil.NewImageFromFile(
		assetsDir+"/playerwalkdown.png",
		ebiten.FilterDefault,
	)
	if err != nil {
		return imagesStruct{}, err
	}
	playerDang, _, err := ebitenutil.NewImageFromFile(
		assetsDir+"/playerwalkdownangle.png",
		ebiten.FilterDefault,
	)
	if err != nil {
		return imagesStruct{}, err
	}
	playerUang, _, err := ebitenutil.NewImageFromFile(
		assetsDir+"/playerwalkupangle.png",
		ebiten.FilterDefault,
	)
	if err != nil {
		return imagesStruct{}, err
	}

	playerSwing, _, err := ebitenutil.NewImageFromFile(
		"assets/playerswing1.png",
		ebiten.FilterDefault,
	)
	if err != nil {
		return imagesStruct{}, err
	}

	emptyImage, _, err := ebitenutil.NewImageFromFile(assetsDir+"/floor.png", ebiten.FilterDefault)
	if err != nil {
		return imagesStruct{}, err
	}
	swordImage, _, err := ebitenutil.NewImageFromFile(assetsDir+"/axe.png", ebiten.FilterDefault)
	if err != nil {
		return imagesStruct{}, err
	}

	bgImage, _, err := ebitenutil.NewImageFromFile(assetsDir+"/8000paint.png", ebiten.FilterDefault)
	if err != nil {
		return imagesStruct{}, err
	}

	is := imagesStruct{}
	is.playerStand = playerStandImage
	is.playerWalkUp = playerWup
	is.playerWalkDown = playerWdown
	is.playerWalkDownAngle = playerDang
	is.playerWalkUpAngle = playerUang
	is.playerSwing = playerSwing
	is.empty = emptyImage
	is.sword = swordImage
	is.bg = bgImage
	return is, nil
}

var centerOn *rectangle

func renderingCenter() location {
	l := rectCenterPoint(*centerOn)
	l.x *= -1
	l.y *= -1
	return l
}

func renderOffset() location {
	center := renderingCenter()
	center.x += ScreenWidth / 2
	center.y += ScreenHeight / 2
	return center
}

func rectCenterPoint(r rectangle) location {
	x := r.location.x + (r.dimens.width / 2)
	y := r.location.y + (r.dimens.height / 2)
	return location{x, y}
}

//var bgOps = &ebiten.DrawImageOptions{}
var backgrounds = make(map[location]bgLoading)
var bgchan = make(chan bgLoading)
var bgtiles = make(map[location]tileAndIm)
var ttmap = make(map[tileType]*ebiten.Image)
var ttshapes = make(map[tileType]shape)
var currentTShapes = make(map[location]shape)

type tileAndIm struct {
	tt tileType
	im *ebiten.Image
}
type tileType int

const (
	blank tileType = iota
	rocky
	offworld
)

func drawBackground(screen *ebiten.Image) {
	myCoordx := centerOn.location.x / bgTileWidth
	myCoordy := centerOn.location.y / bgTileWidth

	select {
	case bgl := <-bgchan:
		newbg := backgrounds[bgl.coords]
		newbg.image = bgl.image
		backgrounds[bgl.coords] = newbg
		//if ti,ok := bgtiles[bgl.coords];ok{
		ttmap[bgl.tiletyp] = bgl.image
		//}
	default:
	}
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			handleBgtile(myCoordx+i, myCoordy+j, screen)
		}
	}
}

func handleBgtile(i int, j int, screen *ebiten.Image) {
	if _, ok := backgrounds[location{i, j}]; !ok {
		log.Println(i, j)
		loadingim := bgLoading{}
		loadingim.image = images.empty
		loadingim.coords = location{i, j}
		loadingim.ops = &ebiten.DrawImageOptions{}
		done := false
		prett := tileType(0)
		if ti, ok := bgtiles[location{i, j}]; ok {
			prett = ti.tt
			if img, ok := ttmap[ti.tt]; ok {
				if img != nil {
					loadingim.image = img
					done = true
				}
			}
		}
		//else {
		//	prett = offworld
		//	if img, ok := ttmap[offworld]; ok {
		//		if img != nil {
		//			loadingim.image = img
		//			done = true
		//		}
		//	}
		//}
		backgrounds[loadingim.coords] = loadingim
		if !done {
			rei := i
			rej := j
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

			go func() {
				im, _, err := ebitenutil.NewImageFromFile(imstring, ebiten.FilterDefault)
				if err != nil {
					panic(err)
				}
				bgl := bgLoading{}
				bgl.coords = location{rei, rej}
				bgl.image = im
				bgl.tiletyp = prett
				bgchan <- bgl
			}()
		}
	}
	if im, ok := backgrounds[location{i, j}]; ok {
		im.ops.GeoM.Reset()
		im.ops.GeoM.Translate(float64(-centerOn.location.x), float64(-centerOn.location.y))
		im.ops.GeoM.Translate(float64(-centerOn.dimens.width/2), float64(-centerOn.dimens.height/2))
		im.ops.GeoM.Translate(float64(ScreenWidth/2), float64(ScreenHeight/2))
		scaleToDimension(dimens{bgTileWidth, bgTileWidth}, im.image, im.ops)
		im.ops.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))

		if err := screen.DrawImage(im.image, im.ops); err != nil {
			log.Fatal(err)
		}
	}
}

func bgShapesWork() {
	myCoordx := centerOn.location.x / bgTileWidth
	myCoordy := centerOn.location.y / bgTileWidth

	for i := -2; i < 3; i++ {
		for j := -2; j < 3; j++ {
			if math.Abs(float64(i)) < 2 && math.Abs(float64(j)) < 2 {
				continue
			}
			checkbgshape(myCoordx+i, myCoordy+j)
		}
	}
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			addTshape(myCoordx+i, myCoordy+j)
		}
	}
}

func addTshape(i, j int) {
	if ti, ok := bgtiles[location{i, j}]; ok {
		currentTShapes[location{i, j}] = ttshapes[ti.tt]
	}
}
func checkbgshape(i, j int) {
	delete(currentTShapes, location{i, j})
}

type bgLoading struct {
	coords  location
	image   *ebiten.Image
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

func cameraShift(loc location, pSpriteOffset location, ops *ebiten.DrawImageOptions) {
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

var toRender []*baseSprite
var offset location

func updateSprites() {
	offset = renderOffset()
	toRender = nil

	for _, bs := range slashers {
		updateSlasherSprite(bs)

	}
	for _, bs := range remotePlayers {
		updateSlasherSprite(bs)

	}
	sort.Slice(toRender, func(i, j int) bool {
		return toRender[i].yaxis < toRender[j].yaxis
	})

}

func updateSlasherSprite(bs *slasher) {
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
	bs.hbarsprit.yaxis = rectCenterPoint(*bs.deth.deathableShape).y + 10
	healthbarlocation := location{bs.deth.deathableShape.location.x, bs.deth.deathableShape.location.y - (bs.deth.deathableShape.dimens.height / 2) - 10}
	healthbardimenswidth := bs.deth.hp.CurrentHP * bs.deth.deathableShape.dimens.width / bs.deth.hp.MaxHP
	scaleToDimension(dimens{healthbardimenswidth, 5}, images.empty, bs.hbarsprit.bOps)
	cameraShift(healthbarlocation, offset, bs.hbarsprit.bOps)

	bs.bsprit.yaxis = rectCenterPoint(*bs.ent.rect).y

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
	scaleto.height += (bs.ent.rect.dimens.height / 2)

	shiftto := location{}
	shiftto.x = bs.ent.rect.location.x
	shiftto.x -= (bs.ent.rect.dimens.width / 4)
	shiftto.y = bs.ent.rect.location.y
	shiftto.y -= (bs.ent.rect.dimens.height / 2)

	scaleToDimension(scaleto, bs.bsprit.sprite, bs.bsprit.bOps)
	cameraShift(shiftto, offset, bs.bsprit.bOps)

	if bs.swangin {
		_, imH := bs.wepsprit.sprite.Size()
		bs.wepsprit.yaxis = bs.pivShape.pivoterShape.lines[0].p2.y
		ownerCenter := rectCenterPoint(*bs.ent.rect)
		cameraShift(ownerCenter, offset, bs.wepsprit.bOps)
		addOp := ebiten.GeoM{}
		hRatio := float64(bs.pivShape.bladeLength+bs.pivShape.bladeLength/4) / float64(imH)
		addOp.Scale(hRatio, hRatio)
		addOp.Translate(-float64(bs.ent.rect.dimens.width)/2, 0)
		addOp.Rotate(bs.pivShape.animationCount - (math.Pi / 2))
		bs.wepsprit.bOps.GeoM.Add(addOp)
	}
}
