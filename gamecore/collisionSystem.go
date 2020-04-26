package gamecore

import (
	"math"
)

var movers = make(map[*entityid]*acceleratingEnt)
var solids = make(map[*entityid]*shape)

type Momentum struct {
	Xaxis int `json:"Xaxis"`
	Yaxis int `json:"Yaxis"`
}
type acceleratingEnt struct {
	rect       *rectangle
	moment     Momentum
	tracktion  float64
	agility    float64
	moveSpeed  float64
	directions Directions
	atkButton  bool
	collides   bool
}

func newControlledEntity() *acceleratingEnt {
	c := &acceleratingEnt{
		newRectangle(
			location{1, 1},
			dimens{20, 40},
		),
		Momentum{},
		0.4,
		0.4,
		10,
		Directions{},
		false,
		true,
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

		if p.directions.Left {
			xmov--
		}
		if p.directions.Right {
			xmov++
		}
		if p.directions.Up {
			ymov--
		}
		if p.directions.Down {
			ymov++
		}

		movedx := xmov != 0
		movedy := ymov != 0

		correctedAgility := p.agility
		if movedx && movedy {
			correctedAgility = p.agility * 0.707
		}

		if xmov < 0 {
			p.moment.Xaxis -= int(correctedAgility*10)
		}
		if xmov > 0 {
			p.moment.Xaxis += int(correctedAgility*10)
		}
		if ymov > 0 {
			p.moment.Yaxis += int(correctedAgility*10)
		}
		if ymov < 0 {
			p.moment.Yaxis -= int(correctedAgility*10)
		}

		magnitude := math.Sqrt(math.Pow(float64(p.moment.Xaxis)/10, 2) + math.Pow(float64(p.moment.Yaxis)/10, 2))
		if magnitude > p.moveSpeed {
			p.moment.Xaxis = int((float64(p.moment.Xaxis) / magnitude) * p.moveSpeed)
			p.moment.Yaxis = int((float64(p.moment.Yaxis) / magnitude) * p.moveSpeed)
		}

		unitmovex := 1
		if p.moment.Xaxis < 0 {
			unitmovex = -1
		}
		unitmovey := 1
		if p.moment.Yaxis < 0 {
			unitmovey = -1
		}
		if !movedx {
			p.moment.Xaxis += int(10*(p.tracktion * -float64(unitmovex)))
		}
		if !movedy {
			p.moment.Yaxis += int(10*(p.tracktion * -float64(unitmovey)))
		}

		if !p.collides{
			newLoc := p.rect.location
			newLoc.x+=int(p.moment.Xaxis)
			newLoc.y+=int(p.moment.Yaxis)
			p.rect.refreshShape(newLoc)
			continue
		}

		absSpdx := math.Abs(float64(p.moment.Xaxis)/10)
		absSpdy := math.Abs(float64(p.moment.Yaxis)/10)
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
					p.moment.Xaxis = 0
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
					p.moment.Yaxis = 0
					ycollided = true
				}
			}

			if xcollided && ycollided {
				break
			}
		}
	}
}
