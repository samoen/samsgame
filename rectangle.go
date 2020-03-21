package main

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

type sysIndex int

const (
	hitBoxRenderable sysIndex = iota
	spriteRenderable
	moveCollider
	solidCollider
	enemyControlled
	playerControlled
	abilityActivator
	hurtable
	rotatingSprite
	weaponBlocker
	pivotingHitbox
)

func eliminate(s *shape) {
	for _, sys := range s.systems {
		switch sys {
		case spriteRenderable:
			for i, renderable := range playerSprites {
				if s == renderable.playerRect.shape {
					if i < len(playerSprites)-1 {
						copy(playerSprites[i:], playerSprites[i+1:])
					}
					playerSprites[len(playerSprites)-1] = nil
					playerSprites = playerSprites[:len(playerSprites)-1]
					break
				}
			}
		case hitBoxRenderable:
			renderingSystem.removeShape(s)
		case moveCollider:
			collideSystem.removeMover(s)
		case solidCollider:
			collideSystem.removeSolid(s)
		case enemyControlled:
			botsMoveSystem.removeEnemyMover(s)
		case abilityActivator:
			slashSystem.removeSlasher(s)
		case hurtable:
			pivotingSystem.removeSlashee(s)
		case pivotingHitbox:
			pivotingSystem.removePivoter(s)
		case rotatingSprite:
			weaponRenderingSystem.removeWeaponSprite(s)
		case playerControlled:
			for i, renderable := range playerControllables {
				if s == renderable.rect.shape {
					if i < len(playerControllables)-1 {
						copy(playerControllables[i:], playerControllables[i+1:])
					}
					playerControllables[len(playerControllables)-1] = nil
					playerControllables = playerControllables[:len(playerControllables)-1]
					break
				}
			}
		case weaponBlocker:
			for i, renderable := range pivotingSystem.blockers {
				if s == renderable {
					if i < len(pivotingSystem.blockers)-1 {
						copy(pivotingSystem.blockers[i:], pivotingSystem.blockers[i+1:])
					}
					pivotingSystem.blockers[len(pivotingSystem.blockers)-1] = nil
					pivotingSystem.blockers = pivotingSystem.blockers[:len(pivotingSystem.blockers)-1]
					break
				}
			}
		}

	}
}

type shape struct {
	lines   []line
	systems []sysIndex
}

func newShape() *shape {
	s := &shape{}
	return s
}

func normalcollides(checkcopy shape, entities []*shape, exclude *shape) bool {
	for _, li := range checkcopy.lines {
		for _, obj := range entities {
			if obj == exclude {
				continue
			}
			for _, subline := range obj.lines {
				if _, _, intersects := subline.intersects(li); intersects {
					return true
				}
			}
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

func newRectangle(loc location, dims dimens) *rectangle {
	r := rectangle{}
	r.dimens = dims
	r.shape = newShape()
	r.refreshShape(loc)
	return &r
}

func (r *rectangle) refreshShape(newpoint location) {
	r.location = newpoint
	left := line{location{r.location.x, r.location.y}, location{r.location.x, r.location.y + r.dimens.height}}
	bottom := line{location{r.location.x, r.location.y + r.dimens.height}, location{r.location.x + r.dimens.width, r.location.y + r.dimens.height}}
	right := line{location{r.location.x + r.dimens.width, r.location.y + r.dimens.height}, location{r.location.x + r.dimens.width, r.location.y}}
	top := line{location{r.location.x + r.dimens.width, r.location.y}, location{r.location.x, r.location.y}}
	r.shape.lines = []line{left, bottom, right, top}
}
