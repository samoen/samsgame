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

type point struct {
	x, y int
}

type line struct {
	p1, p2 point
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

// var fogspace = 20
// var leftbound = line{point{0 + fogspace, 0 + fogspace}, point{0 + fogspace, screenHeight - fogspace}}
// var rightbound = line{point{screenWidth - fogspace, 0 + fogspace}, point{screenWidth - fogspace, screenHeight - fogspace}}
// var topbound = line{point{0 + fogspace, 0 + fogspace}, point{screenWidth - fogspace, 0 + fogspace}}
// var bottombound = line{point{0 + fogspace, screenHeight - fogspace}, point{screenWidth - fogspace, screenHeight - fogspace}}

// func clip(val line) (line, bool) {
// 	newpoint1 := val.p1
// 	newpoint2 := val.p2
// 	totallyOut := false

// 	checkbound := func(bound line, extreme func(point) bool) {
// 		if extreme(newpoint1) && extreme(newpoint2) {
// 			totallyOut = true
// 			return
// 		}
// 		secx, secy, does := line{newpoint1, newpoint2}.intersects(bound)
// 		if does {
// 			if extreme(newpoint1) {
// 				newpoint1 = point{secx, secy}
// 			} else if extreme(newpoint2) {
// 				newpoint2 = point{secx, secy}
// 			}
// 		}

// 	}

// 	checkbound(leftbound, func(p point) bool { return p.x < 0+fogspace })
// 	checkbound(rightbound, func(p point) bool { return p.x > screenWidth-fogspace })
// 	checkbound(topbound, func(p point) bool { return p.y < 0+fogspace })
// 	checkbound(bottombound, func(p point) bool { return p.y > screenHeight-fogspace })

// 	return line{newpoint1, newpoint2}, totallyOut
// }

func samDrawLine(screen, emptyImage *ebiten.Image, center point, l line, op ebiten.DrawImageOptions) {
	// amtx := screenWidth / 2
	// amty := screenHeight / 2

	l.p1.x += center.x
	l.p1.y += center.y
	l.p2.x += center.x
	l.p2.y += center.y

	// l.p1.x -= center.x
	// l.p1.y -= center.y
	// l.p2.x -= center.x
	// l.p2.y -= center.y

	x1 := float64(l.p1.x)
	x2 := float64(l.p2.x)
	y1 := float64(l.p1.y)
	y2 := float64(l.p2.y)
	imgToDraw := *emptyImage

	ew, eh := imgToDraw.Size()
	length := math.Hypot(x2-x1, y2-y1)
	// op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(
		length/float64(ew),
		1/float64(eh),
	)
	op.GeoM.Rotate(math.Atan2(y2-y1, x2-x1))
	op.GeoM.Translate(x1, y1)
	screen.DrawImage(&imgToDraw, &op)
	// ebitenutil.DrawLine()

}

// func (s shape) drawtoScreen(screen *ebiten.Image, center point) {
// 	for _, line := range s {

// 		amtx := screenWidth / 2
// 		amty := screenHeight / 2
// 		line.p1.x += amtx
// 		line.p1.y += amty
// 		line.p2.x += amtx
// 		line.p2.y += amty

// 		line.p1.x -= center.x
// 		line.p1.y -= center.y
// 		line.p2.x -= center.x
// 		line.p2.y -= center.y

// 		newline, totallyOut := clip(line)
// 		if totallyOut == true {
// 			continue
// 		}
// 		newline := line
// 		ebitenutil.DrawLine(
// 			screen,
// 			float64(newline.p1.x),
// 			float64(newline.p1.y),
// 			float64(newline.p2.x),
// 			float64(newline.p2.y),
// 			color.White,
// 		)
// 	}
// }

type rectangle struct {
	location point
	w, h     int
}

func (r rectangle) makeShape() shape {
	left := line{point{r.location.x, r.location.y}, point{r.location.x, r.location.y + r.h}}
	bottom := line{point{r.location.x, r.location.y + r.h}, point{r.location.x + r.w, r.location.y + r.h}}
	right := line{point{r.location.x + r.w, r.location.y + r.h}, point{r.location.x + r.w, r.location.y}}
	top := line{point{r.location.x + r.w, r.location.y}, point{r.location.x, r.location.y}}
	return shape{left, bottom, right, top}
}

type mover struct {
	speed int
}

type playerent struct {
	rectangle
	mover
}

func (r rectangle) normalcollides(entities []shape) bool {
	rectShape := r.makeShape()
	for _, obj := range entities {
		for _, subline := range obj {
			for _, li := range rectShape {
				if _, _, intersects := subline.intersects(li); intersects {
					return true
				}
			}
		}
	}
	return false
}

func (p *playerent) handleMovement(entities []shape) {

	right, down, left, up := false, false, false, false

	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		right = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		down = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		left = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		up = true
	}

	diagonalCorrectedSpeed := p.speed
	if (up || down) && (left || right) {
		diagonalCorrectedSpeed = int(float32(p.speed) * 0.75)
	}

	for i := 1; i < diagonalCorrectedSpeed+1; i++ {
		checkpointx := p.location
		if right {
			checkpointx.x++
		}

		if left {
			checkpointx.x--
		}
		if left || right {
			checkplay := *p
			checkplay.location = checkpointx
			if !checkplay.normalcollides(entities) {
				p.location = checkpointx
			}
		}
		checkpointy := p.location
		if down {
			checkpointy.y++
		}

		if up {
			checkpointy.y--
		}

		if up || down {
			checkplay := *p
			checkplay.location = checkpointy
			if !checkplay.normalcollides(entities) {
				p.location = checkpointy
			}
		}
	}
}

func main() {
	const screenWidth = 1400
	const screenHeight = 1000

	player := playerent{
		rectangle{
			point{
				50,
				50,
			},
			90,
			90,
		},
		mover{9},
	}
	maprect := rectangle{
		point{20, 20},
		2000,
		2000,
	}
	mapBounds := maprect.makeShape()

	// bgImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	// bgSizex, sgsizey := bgImage.Size()
	// bgOps := &ebiten.DrawImageOptions{}
	// bgOps.GeoM.Scale(float64(maprect.w)/float64(bgSizex), float64(maprect.h)/float64(sgsizey))

	diagonalWall := shape{
		line{
			point{250, 310},
			point{600, 655},
		},
	}
	lilRoom := rectangle{
		point{45, 400},
		70,
		20,
	}.makeShape()

	anotherRoom := rectangle{point{300, 100}, 30, 60}.makeShape()
	ents := []shape{
		mapBounds,
		diagonalWall,
		lilRoom,
		anotherRoom,
	}

	emptyImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	// emptyImage, _ := ebiten.NewImage(1, 1, ebiten.FilterDefault)

	emptyop := &ebiten.DrawImageOptions{}
	emptyop.ColorM.Scale(0, 230, 64, 1)

	update := func(screen *ebiten.Image) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return errors.New("game ended by player")
		}

		player.handleMovement(ents)

		if ebiten.IsDrawingSkipped() {
			return nil
		}

		// newops := *bgOps
		// newops.GeoM.Translate(float64(-player.location.x), float64(-player.location.y))
		// screen.DrawImage(bgImage, &newops)

		renderOffset := point{(screenWidth / 2) - player.location.x - (player.w / 2), (screenHeight / 2) - player.location.y - (player.h / 2)}

		for _, shape := range ents {
			// shape.drawtoScreen(screen, player.location)
			for _, l := range shape {
				samDrawLine(screen, emptyImage, renderOffset, l, *emptyop)
			}
		}
		for _, l := range player.makeShape() {
			samDrawLine(screen, emptyImage, renderOffset, l, *emptyop)
		}
		// player.makeShape().drawtoScreen(screen, player.location)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()), 0, 0)
		return nil
	}

	if err := ebiten.Run(update, screenWidth, screenHeight, 1, "sam's cool game"); err != nil {
		log.Fatal(err)
	}
}
