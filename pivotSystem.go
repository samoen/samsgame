package main

import (
	"math"
	"time"
)

type entityid struct {
	systems []sysIndex
}

type pivotingShape struct {
	pivoterShape   *shape
	pivotPoint     *rectangle
	ownerid        *entityid
	animationCount float64
	animating      bool
	doneAnimating  bool
	alreadyHit     map[*entityid]bool
}

func makeAxe(heading float64, centerRect rectangle) []line {
	midPlayer := centerRect.location
	midPlayer.x += centerRect.dimens.width / 2
	midPlayer.y += centerRect.dimens.height / 2
	rotLine := newLinePolar(midPlayer, swordLength, heading)
	crossLine := newLinePolar(rotLine.p2, swordLength/3, heading+math.Pi/2)
	frontCrossLine := newLinePolar(rotLine.p2, swordLength/3, heading-math.Pi/2)
	return []line{rotLine, crossLine, frontCrossLine}
}

func newPivotingShape(owner *entityid, r *rectangle, heading float64) *pivotingShape {
	p := &pivotingShape{}
	p.pivotPoint = r
	p.ownerid = owner
	p.animationCount = heading + 1.6
	p.pivoterShape = newShape()
	p.pivoterShape.lines = makeAxe(p.animationCount, *r)
	p.alreadyHit = make(map[*entityid]bool)
	return p
}

var pivotingSystem = newPivotSystem()

var swordLength = 45

type pivotSystem struct {
	pivoters map[*entityid]*pivotingShape
	blockers map[*entityid]*shape
	// slashees map[*entityid]*shape
}

func newPivotSystem() pivotSystem {
	p := pivotSystem{}
	p.pivoters = make(map[*entityid]*pivotingShape)
	// p.slashees = make(map[*entityid]*shape)
	p.blockers = make(map[*entityid]*shape)
	return p
}

func (p *pivotSystem) addPivoter(eid *entityid, s *pivotingShape) {
	p.pivoters[eid] = s
	eid.systems = append(eid.systems, pivotingHitbox)

	for i := 1; i < 15; i++ {
		if !pivotingSystem.checkBlocker(*s.pivoterShape) {
			break
		} else {
			s.animationCount -= 0.4
			s.pivoterShape.lines = makeAxe(s.animationCount, *s.pivotPoint)
		}
	}

	s.animating = true
	animChan := time.NewTimer(310 * time.Millisecond).C
	go func() {
		select {
		case <-animChan:
			s.doneAnimating = true
		}
	}()
}

func (p *pivotSystem) addBlocker(b *shape, id *entityid) {
	p.blockers[id] = b
	id.systems = append(id.systems, weaponBlocker)
}

func (p *pivotSystem) checkBlocker(sh shape) bool {
	for _, blocker := range p.blockers {
		for _, blockerLine := range blocker.lines {
			for _, bladeLine := range sh.lines {
				if _, _, intersected := bladeLine.intersects(blockerLine); intersected {
					return true
				}
			}
		}
	}
	return false
}
func newLinePolar(loc location, length int, angle float64) line {
	xpos := int(float64(length)*math.Cos(angle)) + loc.x
	ypos := int(float64(length)*math.Sin(angle)) + loc.y
	return line{loc, location{xpos, ypos}}
}

// func rotateAround(center location, point location, angle float64) location {
// 	result := location{}
// 	rotatedX := math.Cos(angle)*float64(point.x-center.x) - math.Sin(angle)*float64(point.y-center.y) + float64(center.x)
// 	rotatedY := math.Sin(angle)*float64(point.x-center.x) + math.Cos(angle)*float64(point.y-center.y) + float64(center.y)
// 	result.x = int(rotatedX)
// 	result.y = int(rotatedY)
// 	return result
// }

func (p *pivotSystem) work() {
	for id, bot := range p.pivoters {

		if bot.doneAnimating {
			// deathables[id] = deathable{}
			eliminate(id)
			continue
		}

		bot.animationCount -= 0.16
		bot.pivoterShape.lines = makeAxe(bot.animationCount, *bot.pivotPoint)
		blocked := p.checkBlocker(*bot.pivoterShape)
		if blocked {
			eliminate(id)
			// deathables[id] = deathable{}
			continue
		} else {
		foundSlashee:
			for slasheeid, slashee := range deathables {
				if slasheeid == bot.ownerid {
					continue foundSlashee
				}
				if _, ok := bot.alreadyHit[slasheeid]; ok {
					continue foundSlashee
				}
				for _, slasheeLine := range slashee.deathableShape.lines {
					for _, bladeLine := range bot.pivoterShape.lines {
						if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
							// for pivID, ps := range p.pivoters {
							// 	if ps.ownerid == slasheeid {
							// 		deathables[pivID] = deathable{}
							// 		// eliminate(pivID)
							// 		break
							// 	}
							// }
							slashee.gotHit = true
							bot.alreadyHit[slasheeid] = true
							// deathables[slasheeid] = deathable{}
							// eliminate(slasheeid)
							break foundSlashee
						}
					}
				}
			}
		}
	}
}
func addDeathable(id *entityid, d *deathable) {
	id.systems = append(id.systems, hurtable)
	deathables[id] = d

	// healthbar := &entityid{}
	// sprite := &weaponSprite{}
	// addWeaponSprite()
}
