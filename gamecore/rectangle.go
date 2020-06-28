package gamecore

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

type shape struct {
	lines []line
}
func (s shape)collidesWith(os shape)bool{
	for _, slasheeLine := range s.lines {
		for _, bladeLine := range os.lines {
			if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
				return true
			}
		}
	}
	return false
}

func normalcollides(checkcopy shape, exclude *shape) bool {
	for _, obj := range wepBlockers {
		if checkcopy.collidesWith(*obj){
			return true
		}
	}
	for _, obj := range currentTShapes {
		if checkcopy.collidesWith(obj){
			return true
		}
	}
	for _, obj := range slashers {
		if obj.locEnt.lSlasher.ent.rect.shape == exclude {
			continue
		}
		if checkcopy.collidesWith(*obj.locEnt.lSlasher.ent.rect.shape){
			return true
		}
	}
	if myLocalPlayer.locEnt.lSlasher.ent.rect.shape != exclude {
		if checkcopy.collidesWith(*myLocalPlayer.locEnt.lSlasher.ent.rect.shape){
			return true
		}
	}
	for _, obj := range remotePlayers {
		if obj.rSlasher.ent.rect.shape == exclude {
			continue
		}
		if checkcopy.collidesWith(*obj.rSlasher.ent.rect.shape){
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
	shape    *shape
}

func newRectangle(loc location, dims dimens) rectangle {
	r := rectangle{}
	r.dimens = dims
	r.shape = &shape{}
	r.refreshShape(loc)
	return r
}

func (r *rectangle) refreshShape(newpoint location) {
	r.location = newpoint
	left := line{location{r.location.x, r.location.y}, location{r.location.x, r.location.y + r.dimens.height}}
	bottom := line{location{r.location.x, r.location.y + r.dimens.height}, location{r.location.x + r.dimens.width, r.location.y + r.dimens.height}}
	right := line{location{r.location.x + r.dimens.width, r.location.y + r.dimens.height}, location{r.location.x + r.dimens.width, r.location.y}}
	top := line{location{r.location.x + r.dimens.width, r.location.y}, location{r.location.x, r.location.y}}
	r.shape.lines = []line{left, bottom, right, top}
}
