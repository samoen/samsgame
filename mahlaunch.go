package main

import (
	// "bytes"
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

type line struct {
	X1, Y1, X2, Y2 int
}

// func (l *line) angle() int {
// 	return math.Atan2(l.Y2-l.Y1, l.X2-l.X1)
// }

type playerent struct {
	// shaper shape
	pos   point
	size  int
	speed int
	// oldverts []point
}
type point struct {
	x, y int
}
type shape struct {
	boundLines []line
}

// func (o shape) points() []point {
// 	// + the startpoint of the first one, for non-closed paths
// 	var points []point
// 	for _, wall := range o.boundLines {
// 		points = append(
// 			points, point{wall.X2, wall.Y2})
// 	}
// 	points = append(points, point{o.boundLines[0].X1, o.boundLines[0].Y1})
// 	return points
// }

// func newRay(x, y, length, angle int) line {
// 	return line{
// 		X1: x,
// 		Y1: y,
// 		X2: x + length*math.Cos(angle),
// 		Y2: y + length*math.Sin(angle),
// 	}
// }

// intersection calculates the intersection of given two lines.
func intersection(l1, l2 line) bool {
	// https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection#Given_two_points_on_each_line
	denom := (l1.X1-l1.X2)*(l2.Y1-l2.Y2) - (l1.Y1-l1.Y2)*(l2.X1-l2.X2)
	tNum := (l1.X1-l2.X1)*(l2.Y1-l2.Y2) - (l1.Y1-l2.Y1)*(l2.X1-l2.X2)
	uNum := -((l1.X1-l1.X2)*(l1.Y1-l2.Y1) - (l1.Y1-l1.Y2)*(l1.X1-l2.X1))

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
	// tdf := u + 1
	// x := l1.X1 + t*(l1.X2-l1.X1)
	// y := l1.Y1 + t*(l1.Y2-l1.Y1)
	return true
}

// var maxspeed = 10

var collisionvec = point{}

// func moverCollides(check point) bool {
// 	oldyverts := player.makeshape().points()
// 	newplayer := player
// 	newplayer.pos = check
// 	newverts := newplayer.makeshape().points()

// 	var collisionLines []line
// 	for i, vert := range oldyverts {
// 		collisionLines = append(collisionLines, line{vert.x, vert.y, newverts[i].x, newverts[i].y})
// 	}
// 	// safe := true
// 	// var stoppedAt point
// 	// collisions:
// 	for _, linerino := range collisionLines {
// 		for _, obj := range entities {
// 			for _, subline := range obj.boundLines {
// 				if intersects := intersection(subline, linerino); intersects {
// 					return true
// 				}
// 			}
// 		}
// 	}
// 	return false
// }

func normalcollides(pp point) bool {
	checkplay := player
	checkplay.pos = pp
	plines := checkplay.makeshape().boundLines

	for _, obj := range entities {
		for _, subline := range obj.boundLines {
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
		// if !normalcollides(checkpoint) {
		// 	player.pos = checkpoint
		// } else if slidex := (point{checkpoint.x, player.pos.y}); !normalcollides(slidex) {
		// 	player.pos = slidex
		// } else if slidey := (point{player.pos.x, checkpoint.y}); !normalcollides(slidey) {
		// 	player.pos = slidey
		// } else {
		// 	break
		// }
		// checkplay:=player
		// checkplay.pos=checkpoint

		// if !moverCollides(checkpoint) {
		// 	player.pos = checkpoint
		// } else if slidex := (point{checkpoint.x, player.pos.y}); !moverCollides(slidex) {
		// 	player.pos = slidex
		// } else if slidey := (point{player.pos.x, checkpoint.y}); !moverCollides(slidey) {
		// 	player.pos = slidey
		// }
	}

	// if !moverCollides(optimistic) {
	// 	playerspeed = maxspeed
	// 	player.movePlayerTo(optimistic)
	// } else if !moverCollides(checkpoint) {
	// 	player.movePlayerTo(checkpoint)
	// } else {

	// 	if slidex := (point{optimistic.x, player.pos.y}); !moverCollides(slidex) {
	// 		player.movePlayerTo(slidex)
	// 		return
	// 	}
	// 	slidey := point{player.pos.x, optimistic.y}
	// 	if !moverCollides(slidey) {
	// 		player.movePlayerTo(slidey)
	// 		return
	// 	}
	// 	playerspeed = 1
	// }

}

// func (p *playerent) movePlayerTo(newpos point) {
// 	p.pos = newpos
// 	p.shaper.boundLines = rect(newpos.x, newpos.y, 5, 5)
// }
func (p *playerent) makeshape() shape {
	s := shape{}
	s.boundLines = rect(
		p.pos.x-(p.size/2),
		p.pos.y-(p.size/2),
		p.size,
		p.size,
	)
	return s
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

	// Draw walls
	for _, obj := range entities {
		for _, w := range obj.boundLines {
			ebitenutil.DrawLine(screen, float64(w.X1), float64(w.Y1), float64(w.X2), float64(w.Y2), color.RGBA{255, 0, 0, 255})
		}
	}
	for _, w := range player.makeshape().boundLines {
		ebitenutil.DrawLine(screen, float64(w.X1), float64(w.Y1), float64(w.X2), float64(w.Y2), color.RGBA{255, 0, 0, 255})
	}
	// Draw player as a rect
	// ebitenutil.DrawRect(screen, int(player.pos.x)-2, int(player.pos.y)-2, 4, 4, color.Black)
	// ebitenutil.DrawRect(screen, int(player.pos.x)-1, int(player.pos.y)-1, 2, 2, color.RGBA{255, 100, 100, 255})

	// ebitenutil.DebugPrintAt(screen, "WASD: move", 160, 0)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()), 51, 51)
	return nil
}

func rect(x, y, w, h int) []line {
	return []line{
		{x, y, x, y + h},
		{x, y + h, x + w, y + h},
		{x + w, y + h, x + w, y},
		{x + w, y, x, y},
	}
}

func main() {
	img, _, err := image.
		Decode(bytes.NewReader(images.Tile_png))
	if err != nil {
		log.Fatal(err)
	}
	bgImage, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	player = playerent{
		// shape{
		// 	rect(50, 50, 4, 4),
		// },
		point{screenWidth / 2,
			screenHeight / 2},
		6,
		3,
		// []point{},
	}

	// Add outer walls
	entities = append(entities,
		shape{
			rect(padding, padding, screenWidth-2*padding, screenHeight-2*padding),
		})

	// Angled wall
	entities = append(entities, shape{[]line{{50, 110, 100, 150}}})

	// Rectangles
	room := shape{}
	room.boundLines = rect(45, 50, 70, 20)

	entities = append(entities, room)
	entities = append(entities, shape{rect(150, 50, 30, 60)})

	if err := ebiten.Run(update, screenWidth, screenHeight, 2, "Ray casting and shadows (Ebiten demo)"); err != nil {
		log.Fatal(err)
	}
}
