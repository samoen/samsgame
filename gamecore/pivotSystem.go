package gamecore

import (
	"math"
)

type entityid struct {
	systems []sysIndex
	remote  bool
}

type pivotingShape struct {
	pivoterShape   *shape
	pivotPoint     *rectangle
	animationCount float64
	alreadyHit     map[*entityid]bool
	startCount     float64
	bladeLength    int
}

func (p *pivotingShape) makeAxe(heading float64, centerRect rectangle) {
	midPlayer := centerRect.location
	midPlayer.x += centerRect.dimens.width / 2
	midPlayer.y += centerRect.dimens.height / 2
	rotLine := newLinePolar(midPlayer, p.bladeLength, heading)
	crossLine := newLinePolar(rotLine.p2, p.bladeLength/3, heading+math.Pi/2)
	frontCrossLine := newLinePolar(rotLine.p2, p.bladeLength/3, heading-math.Pi/2)
	p.pivoterShape.lines = []line{rotLine, crossLine, frontCrossLine}
}

var wepBlockers = make(map[*entityid]*shape)

func addBlocker(b *shape, id *entityid) {
	wepBlockers[id] = b
	id.systems = append(id.systems, weaponBlocker)
}

func checkBlocker(sh shape) bool {
	for _, blocker := range wepBlockers {
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

func checkSlashee(bot *pivotingShape, ownerid *entityid) (bool, *deathable, *entityid) {
	for slasheeid, slashee := range deathables {
		if slasheeid == ownerid {
			continue
		}
		if _, ok := bot.alreadyHit[slasheeid]; ok {
			continue
		}
		for _, slasheeLine := range slashee.deathableShape.shape.lines {
			for _, bladeLine := range bot.pivoterShape.lines {
				if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
					return true, slashee, slasheeid
				}
			}
		}
	}
	return false, nil, nil
}
