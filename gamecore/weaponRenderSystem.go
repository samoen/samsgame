package gamecore

import (
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"log"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten"
)

type baseSprite struct {
	sprite *ebiten.Image
	bOps   *ebiten.DrawImageOptions
	yaxis  int
}

var ScreenWidth = 700

var ScreenHeight = 500
var bgTileWidth = 2500

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

func newImages(assetsDir string) (imagesStruct, error) {
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

var bgOps = &ebiten.DrawImageOptions{}

func drawBackground(screen *ebiten.Image) {
	myBgOps := *bgOps
	myBgOps.GeoM.Translate(float64(-centerOn.location.x), float64(-centerOn.location.y))
	myBgOps.GeoM.Translate(float64(-centerOn.dimens.width/2), float64(-centerOn.dimens.height/2))

	tilesAcross := worldWidth / bgTileWidth

	for i := 0; i < tilesAcross; i++ {
		for j := 0; j < tilesAcross; j++ {
			tileOps := myBgOps
			tileOps.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))
			if err := screen.DrawImage(images.bg, &tileOps); err != nil {
				log.Fatal(err)
			}
		}
	}
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

func updateSlasherSprite(bs *slasher){
	bs.bsprit.bOps.GeoM.Reset()
	bs.bsprit.bOps.ColorM.Reset()
	toRender = append(toRender, bs.bsprit)

	bs.hbarsprit.bOps.GeoM.Reset()
	bs.hbarsprit.bOps.ColorM.Reset()
	toRender = append(toRender, bs.hbarsprit)

	if bs.swangin{
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

	if bs.swangin{
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
