package main

import (
	"math"
)

var collideSystem = collisionSystem{}

type momentum struct {
	xaxis, yaxis float64
}
type acceleratingEnt struct {
	ent       playerent
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

// func (c *collisionSystem) addMover(p *playerent) {
// 	a := acceleratingEnt{p, momentum{}}
// 	c.movers = append(c.movers, &a)
// }
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

		var totalSolids []shape
		for _, sol := range c.solids {
			totalSolids = append(totalSolids, *sol)
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
				checkrect := p.ent.rectangle
				checklocx := checkrect.location
				checklocx.x += unitmovex
				checkrect.refreshShape(checklocx)
				if !checkrect.shape.normalcollides(totalSolids) {
					p.ent.rectangle.refreshShape(checklocx)
				} else {
					p.moment.xaxis = 0
					xcollided = true
				}
			}
			if int(absSpdy) > 0 {
				absSpdy--
				checkrecty := p.ent.rectangle
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
