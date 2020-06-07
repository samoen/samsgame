package gamecore

import (
	"math"
)

type entityid struct {
	systems []sysIndex
}

type pivotingShape struct {
	pivoterShape   *shape
	pivotPoint     *rectangle
	animationCount float64
	alreadyHit     map[*entityid]bool
	startCount     float64
	bladeLength    int
	damage         int
}

func (p *pivotingShape) makeAxe() {
	midPlayer := p.pivotPoint.location
	midPlayer.x += p.pivotPoint.dimens.width / 2
	midPlayer.y += p.pivotPoint.dimens.height / 2
	rotLine := newLinePolar(midPlayer, p.bladeLength, p.animationCount)
	crossLine := newLinePolar(rotLine.p2, p.bladeLength/3, p.animationCount+math.Pi/2)
	frontCrossLine := newLinePolar(rotLine.p2, p.bladeLength/3, p.animationCount-math.Pi/2)
	p.pivoterShape.lines = []line{rotLine, crossLine, frontCrossLine}
}

var wepBlockers = make(map[*entityid]*shape)

func addBlocker(b *shape, id *entityid) {
	wepBlockers[id] = b
	id.systems = append(id.systems, weaponBlocker)
}

func checkBlocker(sh shape) (int,int,bool) {
	for _, blocker := range wepBlockers {
		for _, blockerLine := range blocker.lines {
			for _, bladeLine := range sh.lines {
				if x, y, intersected := bladeLine.intersects(blockerLine); intersected {
					return x,y,true
				}
			}
		}
	}
	return 0,0,false
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

func checkSlashee(bot *pivotingShape, ownerid *entityid, slashables map[*entityid]*slasher) (bool, *slasher, *entityid) {
	for slasheeid, slashee := range slashables {
		if slasheeid == ownerid {
			continue
		}
		if _, ok := bot.alreadyHit[slasheeid]; ok {
			continue
		}
		for _, slasheeLine := range slashee.deth.deathableShape.shape.lines {
			for _, bladeLine := range bot.pivoterShape.lines {
				if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
					return true, slashee, slasheeid
				}
			}
		}
	}
	return false, nil, nil
}
