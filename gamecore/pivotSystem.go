package gamecore

import (
	"math"
)

type entityid struct {
	systems []sysIndex
	linked []*entityid
}

type pivotingShape struct {
	pivoterShape   *shape
	pivotPoint     *rectangle
	ownerid        *entityid
	animationCount float64
	alreadyHit     map[*entityid]bool
	startCount     float64
	swangin *bool
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

var swordLength = 45

var pivoters = make(map[*entityid]*pivotingShape)
var wepBlockers = make(map[*entityid]*shape)

func addPivoter(eid *entityid, s *pivotingShape) {
	pivoters[eid] = s
	eid.systems = append(eid.systems, pivotingHitbox)

	//if ok, _, _ := checkSlashee(s); !ok {
		for i := 1; i < 20; i++ {
			if !checkBlocker(*s.pivoterShape) {
				break
			} else {
				s.animationCount -= 0.2
				s.pivoterShape.lines = makeAxe(s.animationCount, *s.pivotPoint)
			}
		}
	//}
	s.startCount = s.animationCount
}

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

func pivotSystemWork() {
	for id, bot := range pivoters {

		bot.animationCount -= 0.12
		bot.pivoterShape.lines = makeAxe(bot.animationCount, *bot.pivotPoint)
		blocked := checkBlocker(*bot.pivoterShape)

		if ok, slashee, slasheeid := checkSlashee(bot); ok {
			slashee.gotHit = true
			bot.alreadyHit[slasheeid] = true
		}
		if blocked {
			*bot.swangin = false
			eliminate(id)
			continue
		} else if math.Abs(bot.startCount-bot.animationCount) > 2 {
			*bot.swangin = false
			eliminate(id)
			continue
		}
	}
}

func checkSlashee(bot *pivotingShape) (bool, *deathable, *entityid) {
	for slasheeid, slashee := range deathables {
		if slasheeid == bot.ownerid {
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

