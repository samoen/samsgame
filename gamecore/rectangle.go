package gamecore

import (
	"github.com/hajimehoshi/ebiten"
	"log"
	"math"
)

type location struct {
	x, y int
}

type line struct {
	p1, p2 location
}

func (l *line) newLinePolar(loc location, length int, angle float64) {
	xpos := int(float64(length)*math.Cos(angle)) + loc.x
	ypos := int(float64(length)*math.Sin(angle)) + loc.y
	l.p1 = loc
	l.p2 = location{xpos, ypos}
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

func (l line) samDrawLine(s *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	op.ColorM.Scale(0, 230, 64, 1)
	l.p1.x += offset.x
	l.p1.y += offset.y
	l.p2.x += offset.x
	l.p2.y += offset.y

	x1 := float64(l.p1.x)
	x2 := float64(l.p2.x)
	y1 := float64(l.p1.y)
	y2 := float64(l.p2.y)

	imgToDraw := *images.empty
	ew, eh := imgToDraw.Size()
	length := math.Hypot(x2-x1, y2-y1)

	op.GeoM.Scale(
		length/float64(ew),
		2/float64(eh),
	)
	op.GeoM.Rotate(math.Atan2(y2-y1, x2-x1))
	op.GeoM.Translate(x1, y1)
	if err := s.DrawImage(&imgToDraw, &op); err != nil {
		log.Fatal(err)
	}
}

type shape struct {
	lines []line
}

func (s shape) collidesWith(os shape) bool {
	for _, slasheeLine := range s.lines {
		for _, bladeLine := range os.lines {
			if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
				return true
			}
		}
	}
	return false
}

func (s shape) normalcollides(exclude *bool) bool {
	for obj := range wepBlockers {
		if s.collidesWith(*obj) {
			return true
		}
	}
	for _, obj := range currentTShapes {
		if s.collidesWith(obj) {
			return true
		}
	}
	for obj := range slashers {
		if obj.locEnt.lSlasher.ent.collisionId == exclude {
			continue
		}
		if s.collidesWith(obj.locEnt.lSlasher.ent.rect.shape) {
			return true
		}
	}
	if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP > 0 {
		if myLocalPlayer.locEnt.lSlasher.ent.collisionId != exclude {
			if s.collidesWith(myLocalPlayer.locEnt.lSlasher.ent.rect.shape) {
				return true
			}
		}
	}
	for _, obj := range remotePlayers {
		if obj.rSlasher.ent.collisionId == exclude {
			continue
		}
		if s.collidesWith(obj.rSlasher.ent.rect.shape) {
			return true
		}
	}
	return false
}

type dimens struct {
	width, height int
}

type rectangle struct {
	location location
	dimens   dimens
	shape    shape
}

func (r *rectangle) refreshShape(newpoint location) {
	r.location = newpoint
	left := line{location{r.location.x, r.location.y}, location{r.location.x, r.location.y + r.dimens.height}}
	bottom := line{location{r.location.x, r.location.y + r.dimens.height}, location{r.location.x + r.dimens.width, r.location.y + r.dimens.height}}
	right := line{location{r.location.x + r.dimens.width, r.location.y + r.dimens.height}, location{r.location.x + r.dimens.width, r.location.y}}
	top := line{location{r.location.x + r.dimens.width, r.location.y}, location{r.location.x, r.location.y}}
	r.shape.lines = []line{left, bottom, right, top}
}
