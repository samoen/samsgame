package gamecore

import (
	"math"
)

var movers = make(map[*entityid]*acceleratingEnt)
var solids = make(map[*entityid]*shape)

type momentum struct {
	xaxis, yaxis float64
}
type acceleratingEnt struct {
	rect       *rectangle
	moment     momentum
	tracktion  float64
	agility    float64
	moveSpeed  float64
	directions directions
	atkButton  bool
}

func newControlledEntity() *acceleratingEnt {
	c := &acceleratingEnt{
		newRectangle(
			location{1, 1},
			dimens{20, 40},
		),
		momentum{},
		0.4,
		0.4,
		10,
		directions{},
		false,
	}
	return c
}

func addMoveCollider(p *acceleratingEnt, id *entityid) {
	movers[id] = p
	id.systems = append(id.systems, moveCollider)
}

func addSolid(s *shape, id *entityid) {
	solids[id] = s
	id.systems = append(id.systems, solidCollider)
}

func collisionSystemWork() {
	for moverid, p := range movers {
		xmov := 0
		ymov := 0

		if p.directions.left {
			xmov--
		}
		if p.directions.right {
			xmov++
		}
		if p.directions.up {
			ymov--
		}
		if p.directions.down {
			ymov++
		}

		movedx := xmov != 0
		movedy := ymov != 0

		correctedAgility := p.agility
		if movedx && movedy {
			correctedAgility = p.agility * 0.707
		}

		if xmov < 0 {
			p.moment.xaxis -= correctedAgility
		}
		if xmov > 0 {
			p.moment.xaxis += correctedAgility
		}
		if ymov > 0 {
			p.moment.yaxis += correctedAgility
		}
		if ymov < 0 {
			p.moment.yaxis -= correctedAgility
		}

		magnitude := math.Sqrt(math.Pow(p.moment.xaxis, 2) + math.Pow(p.moment.yaxis, 2))
		if magnitude > p.moveSpeed {
			p.moment.xaxis = (p.moment.xaxis / magnitude) * p.moveSpeed
			p.moment.yaxis = (p.moment.yaxis / magnitude) * p.moveSpeed
		}

		unitmovex := 1
		if p.moment.xaxis < 0 {
			unitmovex = -1
		}
		unitmovey := 1
		if p.moment.yaxis < 0 {
			unitmovey = -1
		}
		if !movedx {
			p.moment.xaxis += p.tracktion * -float64(unitmovex)
		}
		if !movedy {
			p.moment.yaxis += p.tracktion * -float64(unitmovey)
		}

		absSpdx := math.Abs(p.moment.xaxis)
		absSpdy := math.Abs(p.moment.yaxis)
		maxSpd := math.Max(absSpdx, absSpdy)

		for i := 1; i < int(maxSpd+1); i++ {
			xcollided := false
			ycollided := false
			if int(absSpdx) > 0 {
				absSpdx--
				checklocx := p.rect.location
				checklocx.x += unitmovex
				checkRect := newRectangle(checklocx, p.rect.dimens)
				if !normalcollides(*checkRect.shape, solids, moverid) {
					p.rect.refreshShape(checklocx)
				} else {
					p.moment.xaxis = 0
					xcollided = true
				}
			}

			if int(absSpdy) > 0 {
				absSpdy--
				checkrecty := *p.rect
				checkrecty.shape = newShape()
				checklocy := checkrecty.location
				checklocy.y += unitmovey
				checkrecty.refreshShape(checklocy)
				if !normalcollides(*checkrecty.shape, solids, moverid) {
					p.rect.refreshShape(checklocy)
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
