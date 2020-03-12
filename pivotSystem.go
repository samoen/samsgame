package main

import (
	"time"
)

type pivotingShape struct {
	pivoterShape   *shape
	pivotPoint     *rectangle
	startangle     float64
	animationCount float64
	animating      bool
	doneAnimating  bool
}

func newPivotingShape(r *rectangle, heading float64) *pivotingShape {
	slashLine := newShape()
	slashLine.lines = []line{line{}}
	p := &pivotingShape{}
	p.pivoterShape = slashLine
	p.pivotPoint = r
	p.animationCount = heading + 1.6
	return p
}

var pivotingSystem = pivotSystem{}

type pivotSystem struct {
	pivoters []*pivotingShape
	blockers []*shape
	slashees []*shape
}

func (p *pivotSystem) addSlashee(s *shape) {
	p.slashees = append(p.slashees, s)
	s.removals = append(s.removals, func() {
		// s.removeSlasher(b.ent.rect.shape)
		p.removeSlashee(s)
	})
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
	s.pivoterShape.removals = append(s.pivoterShape.removals, func() {
		p.removePivoter(s.pivoterShape)
	})
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

func (p *pivotSystem) work() {
	toRemove := []*shape{}
	for _, bot := range p.pivoters {
		bot := bot

		keepOnPlayer := func() bool {
			midPlayer := bot.pivotPoint.location
			midPlayer.x += bot.pivotPoint.dimens.width / 2
			midPlayer.y += bot.pivotPoint.dimens.height / 2

			for i := 0; i < len(bot.pivoterShape.lines); i++ {
				rotLine := newLinePolar(midPlayer, 50, bot.animationCount+bot.startangle)
				bot.pivoterShape.lines[i] = rotLine
			}

			for _, blocker := range p.blockers {
				for _, blockerLine := range blocker.lines {
					for _, bladeLine := range bot.pivoterShape.lines {
						if _, _, intersected := bladeLine.intersects(blockerLine); intersected {
							return false
						}
					}
				}
			}

			return true
		}

		if bot.doneAnimating {
			toRemove = append(toRemove, bot.pivoterShape)
			continue
		}

		bot.animationCount -= 0.16
		notBlocked := keepOnPlayer()
		if !notBlocked {
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
		removeMe.sysPurge()
	}
}
