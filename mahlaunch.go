package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"

	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/examples/resources/images"

	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	screenWidth  = 240
	screenHeight = 240
	padding      = 20
)

var (
	entities []shape
	player   playerent
	bgImage  *ebiten.Image
)

type point struct {
	x, y int
}

type line struct {
	p1, p2 point
}

type shape []line

type playerent struct {
	pos   point
	size  int
	speed int
}

func (p *playerent) makeshape() shape {
	s := rect(
		p.pos.x,
		p.pos.y,
		p.size,
		p.size,
	)
	return s
}

// intersection calculates the intersection of given two lines.
func intersection(l1, l2 line) bool {
	// https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection#Given_two_points_on_each_line
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

func normalcollides(pp point) bool {
	checkplay := player
	checkplay.pos = pp
	plines := checkplay.makeshape()

	for _, obj := range entities {
		for _, subline := range obj {
			for _, li := range plines {
				if intersects := intersection(subline, li); intersects {
					return true
				}
			}
		}
	}
	return false
}

func handleMovement() {

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

	for i := 1; i < player.speed+1; i++ {
		checkpointx := player.pos
		if right {
			checkpointx.x++
		}

		if left {
			checkpointx.x--
		}
		if left || right {
			if !normalcollides(checkpointx) {
				player.pos = checkpointx
			}
		}
		checkpointy := player.pos
		if down {
			checkpointy.y++
		}

		if up {
			checkpointy.y--
		}

		if up || down {
			if !normalcollides(checkpointy) {
				player.pos = checkpointy
			}
		}
	}
}

func update(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("game ended by player")
	}

	handleMovement()

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	// Draw background
	screen.DrawImage(bgImage, nil)

	// Draw entities
	for _, obj := range entities {
		for _, w := range obj {
			ebitenutil.DrawLine(screen, float64(w.p1.x), float64(w.p1.y), float64(w.p2.x), float64(w.p2.y), color.RGBA{255, 0, 0, 255})
		}
	}
	for _, w := range player.makeshape() {
		ebitenutil.DrawLine(screen, float64(w.p1.x), float64(w.p1.y), float64(w.p2.x), float64(w.p2.y), color.RGBA{255, 0, 0, 255})
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()), 51, 51)
	return nil
}

func rect(x, y, w, h int) shape {
	l1 := line{point{x, y}, point{x, y + h}}
	l2 := line{point{x, y + h}, point{x + w, y + h}}
	l3 := line{point{x + w, y + h}, point{x + w, y}}
	l4 := line{point{x + w, y}, point{x, y}}
	return shape{l1, l2, l3, l4}
}

func runSamGame() {
	img, _, _ := image.Decode(bytes.NewReader(images.Tile_png))
	bgImage, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	player = playerent{
		point{
			screenWidth / 2,
			screenHeight / 2,
		},
		6,
		3,
	}

	// Add outer walls
	entities = append(entities, rect(padding, padding, screenWidth-2*padding, screenHeight-2*padding))

	entities = append(entities, shape{{point{50, 110}, point{100, 150}}})
	entities = append(entities, rect(45, 50, 70, 20))
	entities = append(entities, rect(150, 50, 30, 60))
	fmt.Println("before")

	if err := ebiten.Run(update, screenWidth, screenHeight, 2, "Ray casting and shadows (Ebiten demo)"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("after")

}

func main() {
	runSamGame()
}
