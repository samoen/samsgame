package main

import (
	"errors"
	"fmt"

	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/ebiten/ebitenutil"

	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	screenWidth  = 240
	screenHeight = 240
	padding      = 20
)

type point struct {
	x, y int
}

type line struct {
	p1, p2 point
}

func (l1 line) intersects(l2 line) bool {
	denom := (l1.p1.x-l1.p2.x)*(l2.p1.y-l2.p2.y) - (l1.p1.y-l1.p2.y)*(l2.p1.x-l2.p2.x)
	tNum := (l1.p1.x-l2.p1.x)*(l2.p1.y-l2.p2.y) - (l1.p1.y-l2.p1.y)*(l2.p1.x-l2.p2.x)
	uNum := -((l1.p1.x-l1.p2.x)*(l1.p1.y-l2.p1.y) - (l1.p1.y-l1.p2.y)*(l1.p1.x-l2.p1.x))

	if denom == 0 {
		return false
	}

	t := float64(tNum) / float64(denom)
	if t > 1 || t < 0 {
		return false
	}

	u := float64(uNum) / float64(denom)
	if u > 1 || u < 0 {
		return false
	}
	// x := l1.p1.x + t*(l1.p2.x-l1.p1.x)
	// y := l1.Y1 + t*(l1.Y2-l1.Y1)
	return true
}

type shape []line

type rectangle struct {
	location point
	w, h     int
}

func (r rectangle) makeShape() shape {
	l1 := line{point{r.location.x, r.location.y}, point{r.location.x, r.location.y + r.h}}
	l2 := line{point{r.location.x, r.location.y + r.h}, point{r.location.x + r.w, r.location.y + r.h}}
	l3 := line{point{r.location.x + r.w, r.location.y + r.h}, point{r.location.x + r.w, r.location.y}}
	l4 := line{point{r.location.x + r.w, r.location.y}, point{r.location.x, r.location.y}}
	return shape{l1, l2, l3, l4}
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
				if intersects := subline.intersects(li); intersects {
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
	// img, _, _ := image.Decode(bytes.NewReader(images.Tile_png))
	// bgImage, _ := ebiten.NewImageFromImage(img, ebiten.FilterDefault)
	bgImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	bgSizex, sgsizey := bgImage.Size()
	bgOps := &ebiten.DrawImageOptions{}
	bgOps.GeoM.Scale(float64(screenWidth)/float64(bgSizex), float64(screenHeight)/float64(sgsizey))

	ents := []shape{}
	player := playerent{
		rectangle{
			point{
				screenWidth / 2,
				screenHeight / 2,
			},
			6,
			6,
		},
		mover{3},
	}
	mapBounds := rectangle{
		point{
			padding,
			padding,
		},
		screenWidth - 2*padding,
		screenHeight - 2*padding,
	}.makeShape()
	diagonalWall := shape{
		line{
			point{50, 110},
			point{100, 150},
		},
	}
	lilRoom := rectangle{point{
		45, 50,
	}, 70, 20}.makeShape()
	anotherRoom := rectangle{point{150, 50}, 30, 60}.makeShape()
	ents = append(ents, mapBounds, diagonalWall, lilRoom, anotherRoom)

	update := func(screen *ebiten.Image) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return errors.New("game ended by player")
		}

		player.handleMovement(ents)

		if ebiten.IsDrawingSkipped() {
			return nil
		}

		screen.DrawImage(bgImage, bgOps)

		for _, shape := range ents {
			for _, line := range shape {
				ebitenutil.DrawLine(screen, float64(line.p1.x), float64(line.p1.y), float64(line.p2.x), float64(line.p2.y), color.RGBA{255, 0, 0, 255})
			}
		}
		for _, line := range player.makeShape() {
			ebitenutil.DrawLine(screen, float64(line.p1.x), float64(line.p1.y), float64(line.p2.x), float64(line.p2.y), color.RGBA{255, 0, 0, 255})
		}

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()), 0, 0)
		return nil
	}

	if err := ebiten.Run(update, screenWidth, screenHeight, 2, "sam's cool game"); err != nil {
		log.Fatal(err)
	}
}
