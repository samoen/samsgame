package main

import (
	"errors"
	"fmt"
	"math"

	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/ebiten/ebitenutil"

	"github.com/hajimehoshi/ebiten/inpututil"
)

type location struct {
	x, y int
}

type line struct {
	p1, p2 location
}

func (l line) intersects(l2 line) (int, int, bool) {
	denom := (l.p1.x-l.p2.x)*(l2.p1.y-l2.p2.y) - (l.p1.y-l.p2.y)*(l2.p1.x-l2.p2.x)
	tNum := (l.p1.x-l2.p1.x)*(l2.p1.y-l2.p2.y) - (l.p1.y-l2.p1.y)*(l2.p1.x-l2.p2.x)
	uNum := -((l.p1.x-l.p2.x)*(l.p1.y-l2.p1.y) - (l.p1.y-l.p2.y)*(l.p1.x-l2.p1.x))

	if denom == 0 {
		return 0, 0, false
	}

	t := float64(tNum) / float64(denom)
	if t > 1 || t < 0 {
		return 0, 0, false
	}

	u := float64(uNum) / float64(denom)
	if u > 1 || u < 0 {
		return 0, 0, false
	}
	x := l.p1.x + int(t*float64(l.p2.x-l.p1.x))
	y := l.p1.y + int(t*float64(l.p2.y-l.p1.y))
	return x, y, true
}

type shape []line

func samDrawLine(screen, emptyImage *ebiten.Image, center location, l line, op ebiten.DrawImageOptions) {

	l.p1.x += center.x
	l.p1.y += center.y
	l.p2.x += center.x
	l.p2.y += center.y

	x1 := float64(l.p1.x)
	x2 := float64(l.p2.x)
	y1 := float64(l.p1.y)
	y2 := float64(l.p2.y)

	imgToDraw := *emptyImage

	ew, eh := imgToDraw.Size()
	length := math.Hypot(x2-x1, y2-y1)

	op.GeoM.Scale(
		length/float64(ew),
		2/float64(eh),
	)
	op.GeoM.Rotate(math.Atan2(y2-y1, x2-x1))
	op.GeoM.Translate(x1, y1)
	screen.DrawImage(&imgToDraw, &op)
}

type dimens struct {
	width, height int
}

type rectangle struct {
	location
	dimens
	shape
}

func (r rectangle) makeShape() shape {
	left := line{location{r.x, r.y}, location{r.x, r.y + r.height}}
	bottom := line{location{r.x, r.y + r.height}, location{r.x + r.width, r.y + r.height}}
	right := line{location{r.x + r.width, r.y + r.height}, location{r.x + r.width, r.y}}
	top := line{location{r.x + r.width, r.y}, location{r.x, r.y}}
	return shape{left, bottom, right, top}
}

type moveSpeed int

type playerent struct {
	rectangle
	moveSpeed
}

func (r *rectangle) movePlayer(newpoint location) {
	r.location = newpoint
	left := line{location{r.x, r.y}, location{r.x, r.y + r.height}}
	bottom := line{location{r.x, r.y + r.height}, location{r.x + r.width, r.y + r.height}}
	right := line{location{r.x + r.width, r.y + r.height}, location{r.x + r.width, r.y}}
	top := line{location{r.x + r.width, r.y}, location{r.x, r.y}}
	r.shape = shape{left, bottom, right, top}
}

func (s shape) normalcollides(entities []*shape) bool {

	for _, li := range s {
		for _, obj := range entities {
			for _, subline := range *obj {
				if _, _, intersects := subline.intersects(li); intersects {
					return true
				}
			}
		}
	}
	return false
}

type renderSystem struct {
	shapes []*shape
}

func (r renderSystem) work(s, i *ebiten.Image, centerOn rectangle, op ebiten.DrawImageOptions) {
	o := location{(screenWidth / 2) - centerOn.x - (centerOn.width / 2), (screenHeight / 2) - centerOn.y - (centerOn.height / 2)}
	for _, shape := range r.shapes {
		for _, l := range *shape {
			samDrawLine(s, i, o, l, op)
		}
	}
}

func (p *playerent) handleMovement(entities []*shape, right, down, left, up bool) {

	diagonalCorrectedSpeed := p.moveSpeed
	if (up || down) && (left || right) {
		diagonalCorrectedSpeed = moveSpeed(float32(p.moveSpeed) * 0.75)
	}

	for i := 1; i < int(diagonalCorrectedSpeed)+1; i++ {
		checkpointx := p.location
		xcollided := false
		if right {
			checkpointx.x++
		}

		if left {
			checkpointx.x--
		}
		if left || right {
			checkplay := *p
			checkplay.movePlayer(checkpointx)
			if !checkplay.shape.normalcollides(entities) {
				p.movePlayer(checkpointx)
			} else {
				xcollided = true
			}
		}
		checkpointy := p.location
		ycollided := false
		if down {
			checkpointy.y++
		}

		if up {
			checkpointy.y--
		}

		if up || down {
			checkplay := *p
			checkplay.movePlayer(checkpointy)
			if !checkplay.shape.normalcollides(entities) {
				p.movePlayer(checkpointy)
			} else {
				ycollided = true
			}
		}
		if xcollided && ycollided {
			break
		}
	}
}

func newRectangle(loc location, w, h int) rectangle {
	r := rectangle{}
	r.width = w
	r.height = h
	r.movePlayer(loc)
	return r
}

const screenWidth = 1400
const screenHeight = 1000

var emptyImage *ebiten.Image

func init() {
	emptyImagea, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	emptyImage = emptyImagea
}

func main() {

	renderingSystem := renderSystem{}

	player := playerent{
		newRectangle(
			location{
				1,
				1,
			},
			20,
			20,
		),
		moveSpeed(9),
	}

	renderingSystem.shapes = append(renderingSystem.shapes, &player.shape)

	// enemy := playerent{
	// 	rectangle{
	// 		point{
	// 			30,
	// 			1,
	// 		},
	// 		15,
	// 		215,
	// 	},
	// 	mover{9},
	// }

	mapBounds := newRectangle(
		location{0, 0},
		2000,
		2000,
	)

	// bgImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	// bgSizex, sgsizey := bgImage.Size()
	// bgOps := &ebiten.DrawImageOptions{}
	// bgOps.GeoM.Scale(float64(maprect.w)/float64(bgSizex), float64(maprect.h)/float64(sgsizey))

	pImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	pSizex, pSizey := pImage.Size()
	pOps := &ebiten.DrawImageOptions{}
	pOps.GeoM.Scale(float64(player.width)/float64(pSizex), float64(player.height)/float64(pSizey))

	diagonalWall := shape{
		line{
			location{250, 310},
			location{600, 655},
		},
	}
	lilRoom := newRectangle(
		location{45, 400},
		70,
		20,
	)

	anotherRoom := newRectangle(location{900, 1200}, 90, 150)
	ents := []*shape{
		&mapBounds.shape,
		&diagonalWall,
		&lilRoom.shape,
		&anotherRoom.shape,
	}

	renderingSystem.shapes = append(renderingSystem.shapes, ents...)

	// emptyImage, _ := ebiten.NewImage(1, 1, ebiten.FilterDefault)

	emptyop := &ebiten.DrawImageOptions{}
	emptyop.ColorM.Scale(0, 230, 64, 1)

	update := func(screen *ebiten.Image) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return errors.New("game ended by player")
		}

		player.handleMovement(
			ents,
			ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight),
			ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown),
			ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft),
			ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp),
		)
		if ebiten.IsDrawingSkipped() {
			return nil
		}

		// newops := *bgOps
		// newops.GeoM.Translate(float64(-player.x), float64(-player.y))
		// screen.DrawImage(bgImage, &newops)

		renderingSystem.work(screen, emptyImage, player.rectangle, *emptyop)

		// newPOps := *pOps
		// newPOps.GeoM.Translate(float64((screenWidth/2)-(player.w/2)), float64((screenHeight/2)-(player.h/2)))
		// screen.DrawImage(pImage, &newPOps)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()), 0, 0)
		return nil
	}

	if err := ebiten.Run(update, screenWidth, screenHeight, 1, "sam's cool game"); err != nil {
		log.Fatal(err)
	}
}
