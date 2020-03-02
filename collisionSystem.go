package main

import (
	"math"
)

var collideSystem = collisionSystem{}

type momentum struct {
	xaxis, yaxis float64
}
type acceleratingEnt struct {
	ent       *playerent
	moment    momentum
	tracktion float64
	agility   float64
}

type collisionSystem struct {
	movers []*acceleratingEnt
	solids []*shape
}

func (c *collisionSystem) addEnt(p *acceleratingEnt) {
	c.movers = append(c.movers, p)
}

func (r *collisionSystem) removeMover(s *playerent) {
	for i, renderable := range r.movers {
		if s == renderable.ent {
			if i < len(r.movers)-1 {
				copy(r.movers[i:], r.movers[i+1:])
			}
			r.movers[len(r.movers)-1] = nil
			r.movers = r.movers[:len(r.movers)-1]
			break
		}
	}
}

func (r *collisionSystem) removeSolid(s *shape) {
	for i, renderable := range r.solids {
		if s == renderable {
			if i < len(r.solids)-1 {
				copy(r.solids[i:], r.solids[i+1:])
			}
			r.solids[len(r.solids)-1] = nil // or the zero value of T
			r.solids = r.solids[:len(r.solids)-1]
		}
	}
}

func (c *collisionSystem) addSolid(s *shape) {
	c.solids = append(c.solids, s)
}
func (c *collisionSystem) work() {
	for i, p := range c.movers {
		correctedAgilityX := p.agility
		speedLimitx := float64(p.ent.moveSpeed.currentSpeed)
		if p.ent.directions.down || p.ent.directions.up {
			speedLimitx = speedLimitx * 0.707
			correctedAgilityX = p.agility * 0.707
		}
		correctedAgilityY := p.agility
		speedLimity := float64(p.ent.moveSpeed.currentSpeed)
		if p.ent.directions.right || p.ent.directions.left {
			speedLimity = speedLimity * 0.707
			correctedAgilityY = p.agility * 0.707
		}

		movedx := false

		if p.ent.directions.left {
			movedx = true
			// diagCorrected:=p.agility
			desired := p.moment.xaxis - correctedAgilityX
			if desired > -speedLimitx {
				p.moment.xaxis = desired
			} else {
				p.moment.xaxis = -speedLimitx
			}
		}
		if p.ent.directions.right {
			movedx = true
			desired := p.moment.xaxis + correctedAgilityX
			if desired < speedLimitx {
				p.moment.xaxis = desired
			} else {
				p.moment.xaxis = speedLimitx
			}
		}
		movedy := false
		if p.ent.directions.down {
			movedy = true
			desired := p.moment.yaxis + correctedAgilityY
			if desired < speedLimity {
				p.moment.yaxis = desired
			} else {
				p.moment.yaxis = speedLimity
			}
		}
		if p.ent.directions.up {
			movedy = true
			desired := p.moment.yaxis - correctedAgilityY
			if desired > -speedLimity {
				p.moment.yaxis = desired
			} else {
				p.moment.yaxis = -speedLimity
			}
		}
		if movedx && movedy {

		}
		// traction := float64(p.ent.moveSpeed.currentSpeed) / 50
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
			totalSolids = append(totalSolids, movingSolid.ent.rectangle.shape)
		}

		for i := 1; i < int(maxSpd+1); i++ {
			xcollided := false
			ycollided := false
			if int(absSpdx) > 0 {
				absSpdx--
				checklocx := p.ent.rectangle.location
				checklocx.x += unitmovex
				checkRect := newRectangle(checklocx, p.ent.rectangle.dimens)
				if !checkRect.shape.normalcollides(totalSolids) {
					p.ent.rectangle.refreshShape(checklocx)
				} else {
					p.moment.xaxis = 0
					xcollided = true
				}
			}

			if int(absSpdy) > 0 {
				absSpdy--
				checkrecty := *p.ent.rectangle
				checkrecty.shape = &shape{}
				checklocy := checkrecty.location
				checklocy.y += unitmovey
				checkrecty.refreshShape(checklocy)
				if !checkrecty.shape.normalcollides(totalSolids) {
					p.ent.rectangle.refreshShape(checklocy)
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
