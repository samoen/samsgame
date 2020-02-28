package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

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

type dimens struct {
	width, height int
}

type rectangle struct {
	location
	dimens
	shape
}

type moveSpeed int

type playerent struct {
	rectangle
	moveSpeed
	directions
}

func (r *rectangle) movePlayer(newpoint location) {
	r.location = newpoint
	left := line{location{r.x, r.y}, location{r.x, r.y + r.height}}
	bottom := line{location{r.x, r.y + r.height}, location{r.x + r.width, r.y + r.height}}
	right := line{location{r.x + r.width, r.y + r.height}, location{r.x + r.width, r.y}}
	top := line{location{r.x + r.width, r.y}, location{r.x, r.y}}
	r.shape = shape{left, bottom, right, top}
}

func (s shape) normalcollides(entities []shape) bool {
	for _, li := range s {
		for _, obj := range entities {
			for _, subline := range obj {
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

func (r renderSystem) work(s *ebiten.Image, centerOn rectangle) {
	center := location{(screenWidth / 2) - centerOn.x - (centerOn.width / 2), (screenHeight / 2) - centerOn.y - (centerOn.height / 2)}
	samDrawLine := func(l line) {
		op := *emptyop
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
		s.DrawImage(&imgToDraw, &op)
	}

	for _, shape := range r.shapes {
		for _, l := range *shape {
			samDrawLine(l)
		}
	}
}

type collisionSystem struct {
	movers []*playerent
	solids []*shape
}

func (c *collisionSystem) addMover(p *playerent) {
	c.movers = append(c.movers, p)
}
func (c *collisionSystem) addSolid(s *shape) {
	c.solids = append(c.solids, s)
}
func (c *collisionSystem) work() {
	for i, p := range c.movers {
		diagonalCorrectedSpeed := p.moveSpeed
		if (p.directions.up || p.directions.down) && (p.directions.left || p.directions.right) {
			diagonalCorrectedSpeed = moveSpeed(float32(p.moveSpeed) * 0.75)
		}
		var totalSolids []shape
		for _, sol := range c.solids {
			totalSolids = append(totalSolids, *sol)
		}
		for j, movingSolid := range c.movers {
			if i == j {
				continue
			}
			totalSolids = append(totalSolids, movingSolid.shape)
		}
		for i := 1; i < int(diagonalCorrectedSpeed)+1; i++ {
			xcollided := false
			ycollided := false

			doaxis := func(pos *int, collided *bool, checkpoint *location, dir1, dir2 bool) {
				*checkpoint = p.location
				if dir1 {
					*pos++
				}
				if dir2 {
					*pos--
				}
				if dir1 || dir2 {
					checkplay := *p
					checkplay.movePlayer(*checkpoint)
					if !checkplay.shape.normalcollides(totalSolids) {
						p.movePlayer(*checkpoint)
					} else {
						*collided = true
					}
				}
			}
			loca := location{}

			doaxis(&loca.x, &xcollided, &loca, p.directions.right, p.directions.left)
			doaxis(&loca.y, &ycollided, &loca, p.directions.down, p.directions.up)
			if xcollided && ycollided {
				break
			}
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
var emptyop *ebiten.DrawImageOptions

func init() {
	emptyImagea, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	emptyImage = emptyImagea

	emptyop = &ebiten.DrawImageOptions{}
	emptyop.ColorM.Scale(0, 230, 64, 1)
}

func (r *renderSystem) addShape(s *shape) {
	r.shapes = append(r.shapes, s)
}

type botMovementSystem struct {
	events <-chan time.Time
	bots   []*playerent
}

func newBotMovementSystem() botMovementSystem {
	b := botMovementSystem{}
	b.events = time.NewTicker(500 * time.Millisecond).C
	return b
}

func (b *botMovementSystem) addBot(m *playerent) {
	b.bots = append(b.bots, m)
}

func (b *botMovementSystem) work() {
	select {
	case <-b.events:
		for _, bot := range b.bots {
			bot.directions = directions{
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
				rand.Intn(2) == 0,
			}
		}
	default:
	}
}

func main() {
	renderingSystem := renderSystem{}
	collideSystem := collisionSystem{}
	botsMoveSystem := newBotMovementSystem()

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
		directions{},
	}
	renderingSystem.addShape(&player.shape)
	collideSystem.addMover(&player)

	for i := 1; i < 50; i++ {
		enemy := playerent{
			newRectangle(
				location{
					i * 30,
					1,
				},
				20,
				20,
			),
			moveSpeed(5),
			directions{},
		}

		renderingSystem.addShape(&enemy.shape)
		collideSystem.addMover(&enemy)
		botsMoveSystem.addBot(&enemy)
	}

	mapBounds := newRectangle(
		location{0, 0},
		2000,
		2000,
	)
	renderingSystem.addShape(&mapBounds.shape)
	collideSystem.addSolid(&mapBounds.shape)

	// bgImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	// bgSizex, sgsizey := bgImage.Size()
	// bgOps := &ebiten.DrawImageOptions{}
	// bgOps.GeoM.Scale(float64(maprect.w)/float64(bgSizex), float64(maprect.h)/float64(sgsizey))

	// pImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	// pSizex, pSizey := pImage.Size()
	// pOps := &ebiten.DrawImageOptions{}
	// pOps.GeoM.Scale(float64(player.width)/float64(pSizex), float64(player.height)/float64(pSizey))

	diagonalWall := shape{
		line{
			location{250, 310},
			location{600, 655},
		},
	}
	renderingSystem.addShape(&diagonalWall)
	collideSystem.addSolid(&diagonalWall)

	lilRoom := newRectangle(
		location{45, 400},
		70,
		20,
	)
	renderingSystem.addShape(&lilRoom.shape)
	collideSystem.addSolid(&lilRoom.shape)

	anotherRoom := newRectangle(location{900, 1200}, 90, 150)
	renderingSystem.addShape(&anotherRoom.shape)
	collideSystem.addSolid(&anotherRoom.shape)

	update := func(screen *ebiten.Image) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return errors.New("game ended by player")
		}

		player.directions = directions{
			ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight),
			ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown),
			ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft),
			ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp),
		}

		botsMoveSystem.work()
		collideSystem.work()

		if ebiten.IsDrawingSkipped() {
			return nil
		}

		// newops := *bgOps
		// newops.GeoM.Translate(float64(-player.x), float64(-player.y))
		// screen.DrawImage(bgImage, &newops)

		renderingSystem.work(screen, player.rectangle)

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

type directions struct {
	right, down, left, up bool
}
