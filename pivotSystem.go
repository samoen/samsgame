package main

import (
	"math"
	"time"
)

type pivotingShape struct {
	pivoterShape   *shape
	pivotPoint     *rectangle
	animationCount float64
	animating      bool
	doneAnimating  bool
}

func (p *pivotingShape) makeAxe(heading float64) {
	p.animationCount -= heading
	midPlayer := p.pivotPoint.location
	midPlayer.x += p.pivotPoint.dimens.width / 2
	midPlayer.y += p.pivotPoint.dimens.height / 2
	rotLine := newLinePolar(midPlayer, swordLength, p.animationCount)
	crossLine := newLinePolar(rotLine.p2, swordLength/3, p.animationCount+math.Pi/2)
	frontCrossLine := newLinePolar(rotLine.p2, swordLength/3, p.animationCount-math.Pi/2)
	p.pivoterShape.lines = []line{rotLine, crossLine, frontCrossLine}
}

func newPivotingShape(r *rectangle, heading float64) *pivotingShape {
	p := &pivotingShape{}
	p.pivotPoint = r
	p.animationCount = heading + 1.6
	p.pivoterShape = newShape()
	p.makeAxe(0)
	return p
}

var pivotingSystem = pivotSystem{}
var swordLength = 45

type pivotSystem struct {
	pivoters []*pivotingShape
	blockers []*shape
	slashees []*shape
}

func (p *pivotSystem) addSlashee(s *shape) {
	p.slashees = append(p.slashees, s)
	s.systems = append(s.systems, hurtable)
}

func (p *pivotSystem) removeSlashee(sh *shape) {
	for i, renderable := range p.slashees {
		if sh == renderable {
			if i < len(p.slashees)-1 {
				copy(p.slashees[i:], p.slashees[i+1:])
			}
			p.slashees[len(p.slashees)-1] = nil
			p.slashees = p.slashees[:len(p.slashees)-1]
			break
		}
	}
}

func (p *pivotSystem) addPivoter(s *pivotingShape) {
	p.pivoters = append(p.pivoters, s)
	s.pivoterShape.systems = append(s.pivoterShape.systems, pivotingHitbox)

	for i := 1; i < 7; i++ {
		if !pivotingSystem.checkBlocker(s.pivoterShape) {
			break
		} else {
			s.makeAxe(0.5)
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

func (p *pivotSystem) addBlocker(b *shape) {
	p.blockers = append(p.blockers, b)
	b.systems = append(b.systems, weaponBlocker)
}

func (p *pivotSystem) removePivoter(sh *shape) {
	for i, subslasher := range p.pivoters {
		if sh == subslasher.pivoterShape {
			if i < len(p.pivoters)-1 {
				copy(p.pivoters[i:], p.pivoters[i+1:])
			}
			p.pivoters[len(p.pivoters)-1] = nil
			p.pivoters = p.pivoters[:len(p.pivoters)-1]
			break
		}
	}
}

func (p *pivotSystem) checkBlocker(sh *shape) bool {
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
	toRemove := []*shape{}
	for _, bot := range p.pivoters {

		if bot.doneAnimating {
			toRemove = append(toRemove, bot.pivoterShape)
			continue
		}

		bot.makeAxe(0.16)
		blocked := p.checkBlocker(bot.pivoterShape)
		if blocked {
			toRemove = append(toRemove, bot.pivoterShape)
			continue
		} else {
		foundSlashee:
			for _, slashee := range p.slashees {
				if slashee == bot.pivotPoint.shape {
					continue foundSlashee
				}
				for _, slasheeLine := range slashee.lines {
					for _, bladeLine := range bot.pivoterShape.lines {
						if _, _, intersected := bladeLine.intersects(slasheeLine); intersected {
							toRemove = append(toRemove, slashee)
							for _, ps := range p.pivoters {
								if ps.pivotPoint.shape == slashee {
									toRemove = append(toRemove, ps.pivoterShape)
									break
								}
							}
							break foundSlashee
						}
					}
				}
			}
		}
	}
	for _, removeMe := range toRemove {
		eliminate(removeMe)
	}
}
