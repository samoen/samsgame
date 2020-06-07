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

var basicSprites = make(map[*entityid]*baseSprite)

func addBasicSprite(ws *baseSprite, id *entityid) {
	basicSprites[id] = ws
	id.systems = append(id.systems, spriteRenderable)
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

func updateSprites() {
	offset := renderOffset()
	toRender = nil
	for _, bs := range basicSprites {
		bs.bOps.ColorM.Reset()
		bs.bOps.GeoM.Reset()
		toRender = append(toRender, bs)
	}

	for pid, bs := range basicSprites {

		if slasher, ok := slashers[pid]; ok {
			if slasher.deth.redScale > 0 {
				slasher.deth.redScale--
			}
			bs.bOps.ColorM.Translate(float64(slasher.deth.redScale), 0, 0, 0)
			if subbs, ok := basicSprites[slasher.deth.hBarid]; ok {
				subbs.yaxis = rectCenterPoint(*slasher.deth.deathableShape).y + 10
				healthbarlocation := location{slasher.deth.deathableShape.location.x, slasher.deth.deathableShape.location.y - (slasher.deth.deathableShape.dimens.height / 2) - 10}
				healthbardimenswidth := slasher.deth.hp.CurrentHP * slasher.deth.deathableShape.dimens.width / slasher.deth.hp.MaxHP
				scaleToDimension(dimens{healthbardimenswidth, 5}, images.empty, subbs.bOps)
				cameraShift(healthbarlocation, offset, subbs.bOps)
			}

			bs.yaxis = rectCenterPoint(*slasher.ent.rect).y

			spriteSelect := images.empty
			tolerance := math.Pi / 9
			if slasher.swangin {
				spriteSelect = images.playerSwing
			}else if math.Abs(slasher.startangle) < tolerance {
				spriteSelect = images.playerStand
			} else if math.Abs(slasher.startangle-(math.Pi/4)) < tolerance {
				spriteSelect = images.playerWalkDownAngle
			} else if math.Abs(slasher.startangle-(math.Pi/2)) < tolerance {
				spriteSelect = images.playerWalkDown
			} else if math.Abs(slasher.startangle-(3*math.Pi/4)) < tolerance {
				spriteSelect = images.playerWalkDownAngle
			} else if math.Abs(slasher.startangle-(-3*math.Pi/4)) < tolerance {
				spriteSelect = images.playerWalkUpAngle
			}else if math.Abs(slasher.startangle-(-math.Pi/2)) < tolerance {
				spriteSelect = images.playerWalkUp
			}else if math.Abs(slasher.startangle-(-math.Pi/4)) < tolerance {
				spriteSelect = images.playerWalkUpAngle
			}else if math.Abs(slasher.startangle)-math.Pi < tolerance {
				spriteSelect = images.playerStand
			}

			bs.sprite = spriteSelect

			intverted := 1
			if math.Abs(slasher.startangle) > math.Pi/2 {
				intverted = -1
				flipTrans := ebiten.GeoM{}
				flipTrans.Translate(float64(-slasher.ent.rect.dimens.width-(slasher.ent.rect.dimens.width/2)), 0)
				bs.bOps.GeoM.Add(flipTrans)
				bs.bOps.GeoM.Scale(-1, 1)
			}
			scaleto := dimens{}

			scaleto.width = slasher.ent.rect.dimens.width
			scaleto.width += (slasher.ent.rect.dimens.width / 2) * intverted

			scaleto.height = slasher.ent.rect.dimens.height
			scaleto.height += (slasher.ent.rect.dimens.height / 2)

			shiftto := location{}
			shiftto.x = slasher.ent.rect.location.x
			shiftto.x -= (slasher.ent.rect.dimens.width / 4)
			shiftto.y = slasher.ent.rect.location.y
			shiftto.y -= (slasher.ent.rect.dimens.height / 2)

			scaleToDimension(scaleto, bs.sprite, bs.bOps)
			cameraShift(shiftto, offset, bs.bOps)

			if bs, ok := basicSprites[slasher.wepid]; ok {
				_, imH := bs.sprite.Size()
				bs.yaxis = slasher.pivShape.pivoterShape.lines[0].p2.y
				ownerCenter := rectCenterPoint(*slasher.ent.rect)
				cameraShift(ownerCenter, offset, bs.bOps)
				addOp := ebiten.GeoM{}
				hRatio := float64(slasher.pivShape.bladeLength+slasher.pivShape.bladeLength/4) / float64(imH)
				addOp.Scale(hRatio, hRatio)
				addOp.Translate(-float64(slasher.ent.rect.dimens.width)/2, 0)
				addOp.Rotate(slasher.pivShape.animationCount - (math.Pi / 2))
				bs.bOps.GeoM.Add(addOp)
			}
		}

	}
	sort.Slice(toRender, func(i, j int) bool {
		return toRender[i].yaxis < toRender[j].yaxis
	})

}
