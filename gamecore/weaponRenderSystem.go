package gamecore

import (
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten"
)

type baseSprite struct {
	sprite *ebiten.Image
	bOps   *ebiten.DrawImageOptions
	layer  int
}

var ScreenWidth = 700

var ScreenHeight = 500
var bgTileWidth = 2500

var images imagesStruct

type imagesStruct struct {
	playerStand *ebiten.Image
	playerSwing *ebiten.Image
	empty       *ebiten.Image
	sword       *ebiten.Image
	bg          *ebiten.Image
}

func newImages(assetsDir string) (imagesStruct, error) {
	playerStandImage, _, err := ebitenutil.NewImageFromFile(
		assetsDir+"/playerstand.png",
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

	return imagesStruct{
		playerStand: playerStandImage,
		playerSwing: playerSwing,
		empty:       emptyImage,
		sword:       swordImage,
		bg:          bgImage,
	}, nil
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
	ops.GeoM.Scale(wRatio, hRatio)
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

var toRender []baseSprite

func updateSprites() {
	offset := renderOffset()
	for _, bs := range basicSprites {
		bs.bOps.ColorM.Reset()
		bs.bOps.GeoM.Reset()
	}
	for pid, bs := range basicSprites {

		if p, ok := movers[pid]; ok {
			if !p.ignoreflip {
				if p.directions.Left && !p.directions.Right {
					p.lastflip = true
				}
				if p.directions.Right && !p.directions.Left {
					p.lastflip = false
				}
			}

			if p.lastflip {
				invertGeom := ebiten.GeoM{}
				invertGeom.Scale(-1, 1)
				invertGeom.Translate(float64(p.rect.dimens.width), 0)
				bs.bOps.GeoM.Add(invertGeom)
			}

			scaleToDimension(p.rect.dimens, bs.sprite, bs.bOps)
			cameraShift(p.rect.location, offset, bs.bOps)
		}
		if mDeathable, ok := deathables[pid]; ok {
			bs.bOps.ColorM.Translate(float64(mDeathable.redScale), 0, 0, 0)
			if subbs, ok := basicSprites[mDeathable.hBarid]; ok {
				healthbarlocation := location{mDeathable.deathableShape.location.x, mDeathable.deathableShape.location.y - 10}
				healthbardimenswidth := mDeathable.hp.CurrentHP * mDeathable.deathableShape.dimens.width / mDeathable.hp.MaxHP
				scaleToDimension(dimens{healthbardimenswidth, 5}, images.empty, subbs.bOps)
				cameraShift(healthbarlocation, offset, subbs.bOps)
			}
		}
		if bot, ok := slashers[pid]; ok {
			if bot.swangin {
				bs.sprite = images.playerSwing
			} else {
				bs.sprite = images.playerStand
			}
			if bs, ok := basicSprites[bot.wepid]; ok {
				_, imH := bs.sprite.Size()
				ownerCenter := rectCenterPoint(*bot.ent.rect)
				cameraShift(ownerCenter, offset, bs.bOps)
				addOp := ebiten.GeoM{}
				hRatio := float64(bot.pivShape.bladeLength+bot.pivShape.bladeLength/4) / float64(imH)
				addOp.Scale(hRatio, hRatio)
				addOp.Translate(-float64(bot.ent.rect.dimens.width)/2, 0)
				addOp.Rotate(bot.pivShape.animationCount - (math.Pi / 2))
				bs.bOps.GeoM.Add(addOp)
			}
		}
	}
	toRender = nil
	for i := 0; i < 4; i++ {
		for _, bs := range basicSprites {
			if bs.layer == i {
				toRender = append(toRender, *bs)
			}
		}
	}

}
