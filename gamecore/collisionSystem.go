package gamecore

import (
	"math"
)

type Momentum struct {
	Xaxis int
	Yaxis int
}

type acceleratingEnt struct {
	rect        *rectangle
	moment      Momentum
	agility     float64
	moveSpeed   float64
	directions  Directions
	baseloc     location
	endpoint    location
}

func (p *acceleratingEnt)calcMomentum(){

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
		p.moment.Xaxis -= int(correctedAgility)
	}
	if xmov > 0 {
		p.moment.Xaxis += int(correctedAgility)
	}
	if ymov > 0 {
		p.moment.Yaxis += int(correctedAgility)
	}
	if ymov < 0 {
		p.moment.Yaxis -= int(correctedAgility)
	}

	unitmovex := 1
	unitmovey := 1
	if p.moment.Xaxis < 0 {
		unitmovex = -1
	}
	if p.moment.Yaxis < 0 {
		unitmovey = -1
	}
	if !movedx {
		p.moment.Xaxis = int(float64(p.moment.Xaxis) * 0.9)
		if int(math.Abs(float64(p.moment.Xaxis)/10)) < 1 {
			p.moment.Xaxis = 0
		}
	}
	if !movedy {
		p.moment.Yaxis = int(float64(p.moment.Yaxis) * 0.9)
		if int(math.Abs(float64(p.moment.Yaxis)/10)) < 1 {
			p.moment.Yaxis = 0
		}
	}
	if p.moment.Xaxis < 0 {
		unitmovex = -1
	}
	if p.moment.Yaxis < 0 {
		unitmovey = -1
	}

	magnitude := math.Sqrt(math.Pow(float64(p.moment.Xaxis), 2) + math.Pow(float64(p.moment.Yaxis), 2))
	if magnitude > p.moveSpeed {
		if math.Abs(float64(p.moment.Xaxis)) > p.moveSpeed*0.707 {
			p.moment.Xaxis = int(p.moveSpeed * 0.707 * float64(unitmovex))
		}
		if math.Abs(float64(p.moment.Yaxis)) > p.moveSpeed*0.707 {
			p.moment.Yaxis = int(p.moveSpeed * 0.707 * float64(unitmovey))
		}
	}
	//return p.moment
}

func (p *acceleratingEnt)moveCollide() {
	p.calcMomentum()
	unitmovex := 1
	unitmovey := 1

	if p.moment.Xaxis < 0 {
		unitmovex = -1
	}
	if p.moment.Yaxis < 0 {
		unitmovey = -1
	}

	absSpdx := int(math.Abs(float64(p.moment.Xaxis) / 10))
	absSpdy := int(math.Abs(float64(p.moment.Yaxis) / 10))
	maxSpd := absSpdx
	if absSpdy > absSpdx {
		maxSpd = absSpdy
	}
	for i := 1; i < maxSpd+1; i++ {
		xcollided := directionalCollide(&absSpdx,p,unitmovex, 0, &p.moment.Xaxis)
		ycollided := directionalCollide(&absSpdy,p,0,unitmovey,&p.moment.Yaxis)
		if xcollided && ycollided {
			break
		}
	}
}

func directionalCollide(absSpdx *int, p *acceleratingEnt, unitmovex int, unitmovey int, tozero *int)bool{
	if *absSpdx > 0 {
		*absSpdx--
		checklocx := p.rect.location
		checklocx.x += unitmovex
		checklocx.y += unitmovey
		checkRect := newRectangle(checklocx, p.rect.dimens)
		if !normalcollides(*checkRect.shape, p.rect.shape) {
			p.rect.refreshShape(checklocx)
		} else {
			*absSpdx = 0
			*tozero = 0
			return true
		}
	}
	return false
}

type remoteMoveState int

const (
	interpolating remoteMoveState = iota
	deadreckoning
	momentumOnly
)
const interpTime = 4
const deathreckTime = 4
