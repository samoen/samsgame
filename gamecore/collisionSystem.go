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
	remote bool
}

func newControlledEntity() *acceleratingEnt {
	c := &acceleratingEnt{
		newRectangle(
			location{1, 1},
			dimens{20, 40},
		),
		Momentum{},
		0.1,
		1,
		10,
		Directions{},
		false,
		true,
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

const DEADRECKON = 8
var lagcompcount = 0
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
		if p.remote{
			if lagcompcount>1{
				lagcompcount--
			}else{
				p.directions.Left = false
				p.directions.Right = false
				p.directions.Up = false
				p.directions.Down = false
			}
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
		//if !p.remote{
			magnitude := math.Sqrt(math.Pow(float64(p.moment.Xaxis)/10, 2) + math.Pow(float64(p.moment.Yaxis)/10, 2))
			if magnitude > p.moveSpeed {
				p.moment.Xaxis = int((float64(p.moment.Xaxis) / magnitude) * p.moveSpeed)
				p.moment.Yaxis = int((float64(p.moment.Yaxis) / magnitude) * p.moveSpeed)
			}
		//}
		unitmovex := 1
		unitmovey := 1
		if p.moment.Xaxis < 0 {
			unitmovex = -1
		}
		if p.moment.Yaxis < 0 {
			unitmovey = -1
		}
		//if !movedx {
		p.moment.Xaxis += int(float64(-p.moment.Xaxis)*(p.tracktion))
		if math.Abs(float64(p.moment.Xaxis))<2{
			p.moment.Xaxis = 0
		}
		//if unitmovex>0{
		//	if p.moment.Xaxis<0{
		//		p.moment.Xaxis = 0
		//	}
		//}else{
		//	if p.moment.Xaxis>0{
		//		p.moment.Xaxis = 0
		//	}
		//}
		//}
		//if !movedy {
		p.moment.Yaxis += int(float64(-p.moment.Yaxis)*(p.tracktion))
		if math.Abs(float64(p.moment.Yaxis))<2{
			p.moment.Yaxis = 0
		}
		//if unitmovey>0{
		//	if p.moment.Yaxis<0{
		//		p.moment.Yaxis = 0
		//	}
		//}else{
		//	if p.moment.Yaxis>0{
		//		p.moment.Yaxis = 0
		//	}
		//}
		//}

		if p.moment.Xaxis < 0 {
			unitmovex = -1
		}
		if p.moment.Yaxis < 0 {
			unitmovey = -1
		}

		if !p.collides{
			newLoc := p.rect.location
			newLoc.x+=int(p.moment.Xaxis)
			newLoc.y+=int(p.moment.Yaxis)
			p.rect.refreshShape(newLoc)
			continue
		}

		absSpdx := int(math.Abs(float64(p.moment.Xaxis)/10))
		absSpdy := int(math.Abs(float64(p.moment.Yaxis)/10))
		maxSpd := absSpdx
		if absSpdy>absSpdx{
			maxSpd = absSpdy
		}
		for i := 1; i < maxSpd+1; i++ {
			xcollided := false
			ycollided := false
			if absSpdx > 0 {
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

			if absSpdy > 0 {
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
