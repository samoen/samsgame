package main

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
	moveSpeed  moveSpeed
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
		moveSpeed{10},
		directions{},
		false,
	}
	return c
}

func addEnt(p *acceleratingEnt, id *entityid) {
	movers[id] = p
	id.systems = append(id.systems, moveCollider)
}

func addSolid(s *shape, id *entityid) {
	solids[id] = s
	id.systems = append(id.systems, solidCollider)
}

func (p *acceleratingEnt) drive() {
	correctedAgilityX := p.agility
	speedLimitx := float64(p.moveSpeed.currentSpeed)
	if p.directions.down || p.directions.up {
		correctedAgilityX = p.agility * 0.707
	}
	correctedAgilityY := p.agility
	speedLimity := float64(p.moveSpeed.currentSpeed)
	if p.directions.right || p.directions.left {
		correctedAgilityY = p.agility * 0.707
	}

	movedx := false
	if p.directions.left && !p.directions.right {
		movedx = true
		desired := p.moment.xaxis - correctedAgilityX
		if desired > -speedLimitx {
			p.moment.xaxis = desired
		} else {
			p.moment.xaxis = -speedLimitx
		}
	}
	if p.directions.right && !p.directions.left {
		movedx = true
		desired := p.moment.xaxis + correctedAgilityX
		if desired < speedLimitx {
			p.moment.xaxis = desired
		} else {
			p.moment.xaxis = speedLimitx
		}
	}
	movedy := false
	if p.directions.down && !p.directions.up {
		movedy = true
		desired := p.moment.yaxis + correctedAgilityY
		if desired < speedLimity {
			p.moment.yaxis = desired
		} else {
			p.moment.yaxis = speedLimity
		}
	}
	if p.directions.up && !p.directions.down {
		movedy = true
		desired := p.moment.yaxis - correctedAgilityY
		if desired > -speedLimity {
			p.moment.yaxis = desired
		} else {
			p.moment.yaxis = -speedLimity
		}
	}

	if !movedx {
		if p.moment.xaxis > 0 {
			p.moment.xaxis -= p.tracktion
		}
		if p.moment.xaxis < 0 {
			p.moment.xaxis += p.tracktion
		}
	}
	if !movedy {
		if p.moment.yaxis > 0 {
			p.moment.yaxis -= p.tracktion
		}
		if p.moment.yaxis < 0 {
			p.moment.yaxis += p.tracktion
		}
	}

	if math.Sqrt(math.Pow(p.moment.xaxis, 2)+math.Pow(p.moment.yaxis, 2)) > speedLimitx {
		p.moment.xaxis = p.moment.xaxis * 0.707
		p.moment.yaxis = p.moment.yaxis * 0.707
	}
}

func collisionSystemWork() {
	for moverid, p := range movers {
		p.drive()
		unitmovex := 1
		if p.moment.xaxis < 0 {
			unitmovex = -1
		}
		unitmovey := 1
		if p.moment.yaxis < 0 {
			unitmovey = -1
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
