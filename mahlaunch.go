package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"math"
	"math/rand"
	"time"

	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/examples/resources/images"

	"github.com/hajimehoshi/ebiten/inpututil"
)

type location struct {
	x, y int
}

type line struct {
	p1, p2 location
}
type directions struct {
	right, down, left, up bool
}

type shape []line

type dimens struct {
	width, height int
}

type rectangle struct {
	location location
	dimens   dimens
	shape    shape
}

type moveSpeed struct {
	currentSpeed int
	maxSpeed     int
}
type playerent struct {
	rectangle  rectangle
	moveSpeed  moveSpeed
	directions directions
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

func (r *rectangle) refreshShape(newpoint location) {
	r.location = newpoint
	left := line{location{r.location.x, r.location.y}, location{r.location.x, r.location.y + r.dimens.height}}
	bottom := line{location{r.location.x, r.location.y + r.dimens.height}, location{r.location.x + r.dimens.width, r.location.y + r.dimens.height}}
	right := line{location{r.location.x + r.dimens.width, r.location.y + r.dimens.height}, location{r.location.x + r.dimens.width, r.location.y}}
	top := line{location{r.location.x + r.dimens.width, r.location.y}, location{r.location.x, r.location.y}}
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
	center := location{(screenWidth / 2) - centerOn.location.x - (centerOn.dimens.width / 2), (screenHeight / 2) - centerOn.location.y - (centerOn.dimens.height / 2)}
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
	movers []*acceleratingEnt
	solids []*shape
}

func (c *collisionSystem) addEnt(p *acceleratingEnt) {
	c.movers = append(c.movers, p)
}

// func (c *collisionSystem) addMover(p *playerent) {
// 	a := acceleratingEnt{p, momentum{}}
// 	c.movers = append(c.movers, &a)
// }
func (c *collisionSystem) addSolid(s *shape) {
	c.solids = append(c.solids, s)
}
func (c *collisionSystem) work() {
	for i, p := range c.movers {

		// agility := 0.8
		// canGoFaster := math.Abs(p.moment.yaxis)+math.Abs(p.moment.xaxis) < float64(p.ent.moveSpeed.currentSpeed)

		speedLimitx := float64(p.ent.moveSpeed.currentSpeed)/2 + ((float64(p.ent.moveSpeed.currentSpeed) - math.Abs(p.moment.yaxis/1)) / 2)
		speedLimity := float64(p.ent.moveSpeed.currentSpeed)/2 + ((float64(p.ent.moveSpeed.currentSpeed) - math.Abs(p.moment.xaxis/1)) / 2)

		if p.ent.directions.left {
			if p.moment.xaxis > -speedLimitx {
				p.moment.xaxis -= p.agility
			}
		}
		if p.ent.directions.right {
			if p.moment.xaxis < speedLimitx {
				p.moment.xaxis += p.agility
			}
		}
		if p.ent.directions.down {
			if p.moment.yaxis < speedLimity {
				p.moment.yaxis += p.agility
			}
		}
		if p.ent.directions.up {
			if p.moment.yaxis > -speedLimity {
				p.moment.yaxis -= p.agility
			}
		}
		// traction := float64(p.ent.moveSpeed.currentSpeed) / 50
		if !p.ent.directions.left && !p.ent.directions.right {
			if p.moment.xaxis > 0 {
				p.moment.xaxis -= p.tracktion
			}
			if p.moment.xaxis < 0 {
				p.moment.xaxis += p.tracktion
			}
		}
		if !p.ent.directions.up && !p.ent.directions.down {
			if p.moment.yaxis > 0 {
				p.moment.yaxis -= p.tracktion
			}
			if p.moment.yaxis < 0 {
				p.moment.yaxis += p.tracktion
			}
		}

		unitmovex := 1
		if p.moment.xaxis < 0 {
			unitmovex = -1
		}
		unitmovey := 1
		if p.moment.yaxis < 0 {
			unitmovey = -1
		}

		absSpdx := math.Abs(p.moment.xaxis)
		absSpdy := math.Abs(p.moment.yaxis)
		maxSpd := math.Max(absSpdx, absSpdy)

		var totalSolids []shape
		for _, sol := range c.solids {
			totalSolids = append(totalSolids, *sol)
		}
		for j, movingSolid := range c.movers {
			if i == j {
				continue
			}
			totalSolids = append(totalSolids, movingSolid.ent.rectangle.shape)
		}

		for i := 1; i < int(maxSpd+1); i++ {
			xcollided := false
			ycollided := false
			if int(absSpdx) > 0 {
				absSpdx--
				checkrect := p.ent.rectangle
				checklocx := checkrect.location
				checklocx.x += unitmovex
				checkrect.refreshShape(checklocx)
				if !checkrect.shape.normalcollides(totalSolids) {
					p.ent.rectangle.refreshShape(checklocx)
				} else {
					p.moment.xaxis = 0
					xcollided = true
				}
			}
			if int(absSpdy) > 0 {
				absSpdy--
				checkrecty := p.ent.rectangle
				checklocy := checkrecty.location
				checklocy.y += unitmovey
				checkrecty.refreshShape(checklocy)
				if !checkrecty.shape.normalcollides(totalSolids) {
					p.ent.rectangle.refreshShape(checklocy)
				} else {
					p.moment.yaxis = 0
					ycollided = true
				}
			}

			if xcollided && ycollided {
				break
			}
		}
	}
}

func newRectangle(loc location, dims dimens) rectangle {
	r := rectangle{}
	r.dimens = dims
	r.refreshShape(loc)
	return r
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

func newPlayerMovementSystem() botMovementSystem {
	b := botMovementSystem{}
	b.events = time.NewTicker(50 * time.Millisecond).C
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
func (b *botMovementSystem) workForPlayer() {
	select {
	case <-b.events:
		for _, bot := range b.bots {
			bot.directions = directions{
				ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight),
				ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown),
				ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft),
				ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp),
			}
		}
	default:
	}
}

type acceleratingEnt struct {
	ent       *playerent
	moment    momentum
	tracktion float64
	agility   float64
}

type momentum struct {
	xaxis, yaxis float64
}

type accelerationSystem struct {
	events <-chan time.Time
	bots   []*acceleratingEnt
}

func newAccelerationSystem() accelerationSystem {
	a := accelerationSystem{}
	a.events = time.NewTicker(50 * time.Millisecond).C
	return a
}

// func (a *accelerationSystem) addAccelerator(m *playerent) {
// 	aEnt := acceleratingEnt{m, momentum{}}
// 	a.bots = append(a.bots, &aEnt)
// }

// func (a *accelerationSystem) handleAcceleration() {
// 	select {
// 	case <-a.events:
// 		for range a.bots {
// 		}
// 	default:
// 	}
// }

const screenWidth = 1400
const screenHeight = 1000

var emptyImage *ebiten.Image
var emptyop *ebiten.DrawImageOptions
var (
	layers = [][]int{
		{
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 218, 243, 243, 243, 243, 243, 243, 243, 243, 243, 218, 243, 244, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,

			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 243, 244, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 219, 243, 243, 243, 219, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,

			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			243, 218, 243, 243, 243, 243, 243, 243, 243, 243, 243, 244, 243, 243, 243,
			243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 26, 27, 28, 29, 30, 31, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 51, 52, 53, 54, 55, 56, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 76, 77, 78, 79, 80, 81, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 101, 102, 103, 104, 105, 106, 0, 0, 0, 0,

			0, 0, 0, 0, 0, 126, 127, 128, 129, 130, 131, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 303, 303, 245, 242, 303, 303, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0,

			0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 245, 242, 0, 0, 0, 0, 0, 0,
		},
	}
)

func main() {
	renderingSystem := renderSystem{}
	collideSystem := collisionSystem{}
	botsMoveSystem := newBotMovementSystem()
	playerMoveSystem := newPlayerMovementSystem()
	// accelerationSystem := newAccelerationSystem()

	player := playerent{
		newRectangle(
			location{1, 1},
			dimens{20, 20},
		),
		moveSpeed{9, 9},
		directions{},
	}
	accelplayer := acceleratingEnt{&player, momentum{}, 0.4, 0.4}
	playerMoveSystem.addBot(&player)
	// accelerationSystem.addAccelerator(&player)
	renderingSystem.addShape(&player.rectangle.shape)
	collideSystem.addEnt(&accelplayer)

	for i := 1; i < 50; i++ {
		enemy := playerent{
			newRectangle(
				location{
					i * 30,
					1,
				},
				dimens{20, 20},
			),
			moveSpeed{9, 9},
			directions{},
		}
		moveEnemy := acceleratingEnt{
			&enemy,
			momentum{},
			0.4,
			0.4,
		}
		// accelerationSystem.addAccelerator(&enemy)
		renderingSystem.addShape(&enemy.rectangle.shape)
		collideSystem.addEnt(&moveEnemy)
		botsMoveSystem.addBot(&enemy)
	}

	mapBounds := newRectangle(
		location{0, 0},
		dimens{2000, 2000},
	)
	renderingSystem.addShape(&mapBounds.shape)
	collideSystem.addSolid(&mapBounds.shape)

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
		dimens{70, 20},
	)
	renderingSystem.addShape(&lilRoom.shape)
	collideSystem.addSolid(&lilRoom.shape)

	anotherRoom := newRectangle(location{900, 1200}, dimens{90, 150})
	renderingSystem.addShape(&anotherRoom.shape)
	collideSystem.addSolid(&anotherRoom.shape)

	img, _, err := image.Decode(bytes.NewReader(images.Tiles_png))
	if err != nil {
		log.Fatal(err)
	}

	emptyImagea, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	emptyImage = emptyImagea
	emptyop = &ebiten.DrawImageOptions{}
	emptyop.ColorM.Scale(0, 230, 64, 1)

	// bgImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	// bgSizex, sgsizey := bgImage.Size()
	// bgOps := &ebiten.DrawImageOptions{}
	// bgOps.GeoM.Scale(float64(mapBounds.dimens.width)/float64(bgSizex), float64(mapBounds.dimens.height)/float64(sgsizey))
	// bgOps.GeoM.Translate(float64(screenWidth/2), float64(screenHeight/2))
	// pImage, _, _ := ebitenutil.NewImageFromFile("assets/floor.png", ebiten.FilterDefault)
	// pSizex, pSizey := pImage.Size()
	// pOps := &ebiten.DrawImageOptions{}
	// pOps.GeoM.Scale(float64(player.width)/float64(pSizex), float64(player.height)/float64(pSizey))

	const (
		tilescreenWidth  = 240
		tilescreenHeight = 240
		tileSize         = 16
		tileXNum         = 25
		xNum             = tilescreenWidth / tileSize
	)

	tilesImage, _ := ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	update := func(screen *ebiten.Image) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return errors.New("game ended by player")
		}

		playerMoveSystem.workForPlayer()
		// accelerationSystem.handleAcceleration()
		botsMoveSystem.work()
		collideSystem.work()

		if ebiten.IsDrawingSkipped() {
			return nil
		}

		for _, l := range layers {
			for i, t := range l {
				tileOps := &ebiten.DrawImageOptions{}
				tileOps.GeoM.Translate(float64((i%xNum)*tileSize), float64((i/xNum)*tileSize))

				tileOps.GeoM.Translate(float64(screenWidth/2), float64(screenHeight/2))
				tileOps.GeoM.Translate(float64(-player.rectangle.location.x), float64(-player.rectangle.location.y))
				tileOps.GeoM.Translate(float64(-player.rectangle.dimens.width/2), float64(-player.rectangle.dimens.height/2))
				// tileOps.GeoM.Scale(2, 2)
				// tileOps.GeoM.Scale(float64(mapBounds.dimens.width)/float64(tileImSizex), float64(mapBounds.dimens.height)/float64(tileImSizey))

				sx := (t % tileXNum) * tileSize
				sy := (t / tileXNum) * tileSize
				subImage := tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image)
				screen.DrawImage(subImage, tileOps)
			}
		}

		// newops := *bgOps
		// newops.GeoM.Translate(float64(-player.rectangle.location.x), float64(-player.rectangle.location.y))
		// newops.GeoM.Translate(float64(-player.rectangle.dimens.width/2), float64(-player.rectangle.dimens.height/2))
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
