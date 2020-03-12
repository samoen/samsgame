package main

import (
	"math"
)

var collideSystem = collisionSystem{}

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
			dimens{20, 20},
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

type collisionSystem struct {
	movers []*acceleratingEnt
	solids []*shape
}

func (c *collisionSystem) addEnt(p *acceleratingEnt) {
	c.movers = append(c.movers, p)
	p.rect.shape.removals = append(p.rect.shape.removals, func() {
		c.removeMover(p.rect.shape)
	})
}

func (c *collisionSystem) removeMover(s *shape) {
	for i, renderable := range c.movers {
		if s == renderable.rect.shape {
			if i < len(c.movers)-1 {
				copy(c.movers[i:], c.movers[i+1:])
			}
			c.movers[len(c.movers)-1] = nil
			c.movers = c.movers[:len(c.movers)-1]
			break
		}
	}
}

func (c *collisionSystem) removeSolid(s *shape) {
	for i, renderable := range c.solids {
		if s == renderable {
			if i < len(c.solids)-1 {
				copy(c.solids[i:], c.solids[i+1:])
			}
			c.solids[len(c.solids)-1] = nil
			c.solids = c.solids[:len(c.solids)-1]
		}
	}
}

func (c *collisionSystem) addSolid(s *shape) {
	c.solids = append(c.solids, s)
}

func (c *collisionSystem) work() {
	for i, p := range c.movers {
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
		if p.directions.left {
			movedx = true
			desired := p.moment.xaxis - correctedAgilityX
			if desired > -speedLimitx {
				p.moment.xaxis = desired
			} else {
				p.moment.xaxis = -speedLimitx
			}
		}
		if p.directions.right {
			movedx = true
			desired := p.moment.xaxis + correctedAgilityX
			if desired < speedLimitx {
				p.moment.xaxis = desired
			} else {
				p.moment.xaxis = speedLimitx
			}
		}
		movedy := false
		if p.directions.down {
			movedy = true
			desired := p.moment.yaxis + correctedAgilityY
			if desired < speedLimity {
				p.moment.yaxis = desired
			} else {
				p.moment.yaxis = speedLimity
			}
		}
		if p.directions.up {
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

		var totalSolids []*shape
		for _, sol := range c.solids {
			totalSolids = append(totalSolids, sol)
		}
		for j, movingSolid := range c.movers {

			if i == j {
				continue
			}
			totalSolids = append(totalSolids, movingSolid.rect.shape)
		}

		for i := 1; i < int(maxSpd+1); i++ {
			xcollided := false
			ycollided := false
			if int(absSpdx) > 0 {
				absSpdx--
				checklocx := p.rect.location
				checklocx.x += unitmovex
				checkRect := newRectangle(checklocx, p.rect.dimens)
				if !checkRect.shape.normalcollides(totalSolids) {
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
				if !checkrecty.shape.normalcollides(totalSolids) {
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
