package main

import (
	"math"
)

type momentum struct {
	Xaxis int
	Yaxis int
}

type acceleratingEnt struct {
	collisionId *bool
	rect        rectangle
	moment      momentum
	agility     float64
	moveSpeed   float64
	directions  directions
	baseloc     location
	endpoint    location
}

func (a *acceleratingEnt) spawnSafe() {
	for {
		if a.rect.shape.normalcollides(a.collisionId) {
			a.rect.refreshShape(location{a.rect.location.x, a.rect.location.y + 20})
		} else {
			break
		}
	}
}

func (a *acceleratingEnt) calcMomentum() {

	xmov := 0
	ymov := 0

	if a.directions.Left {
		xmov--
	}
	if a.directions.Right {
		xmov++
	}
	if a.directions.Up {
		ymov--
	}
	if a.directions.Down {
		ymov++
	}

	movedx := xmov != 0
	movedy := ymov != 0

	correctedAgility := a.agility
	if movedx && movedy {
		correctedAgility = a.agility * 0.707
	}

	if xmov < 0 {
		a.moment.Xaxis -= int(correctedAgility)
	}
	if xmov > 0 {
		a.moment.Xaxis += int(correctedAgility)
	}
	if ymov > 0 {
		a.moment.Yaxis += int(correctedAgility)
	}
	if ymov < 0 {
		a.moment.Yaxis -= int(correctedAgility)
	}

	unitmovex := 1
	unitmovey := 1
	if a.moment.Xaxis < 0 {
		unitmovex = -1
	}
	if a.moment.Yaxis < 0 {
		unitmovey = -1
	}
	if !movedx {
		a.moment.Xaxis = int(float64(a.moment.Xaxis) * 0.9)
		if int(math.Abs(float64(a.moment.Xaxis)/10)) < 1 {
			a.moment.Xaxis = 0
		}
	}
	if !movedy {
		a.moment.Yaxis = int(float64(a.moment.Yaxis) * 0.9)
		if int(math.Abs(float64(a.moment.Yaxis)/10)) < 1 {
			a.moment.Yaxis = 0
		}
	}
	if a.moment.Xaxis < 0 {
		unitmovex = -1
	}
	if a.moment.Yaxis < 0 {
		unitmovey = -1
	}

	magnitude := math.Sqrt(math.Pow(float64(a.moment.Xaxis), 2) + math.Pow(float64(a.moment.Yaxis), 2))
	if magnitude > a.moveSpeed {
		if math.Abs(float64(a.moment.Xaxis)) > a.moveSpeed*0.707 {
			a.moment.Xaxis = int(a.moveSpeed * 0.707 * float64(unitmovex))
		}
		if math.Abs(float64(a.moment.Yaxis)) > a.moveSpeed*0.707 {
			a.moment.Yaxis = int(a.moveSpeed * 0.707 * float64(unitmovey))
		}
	}
	//return a.moment
}

func (a *acceleratingEnt) moveCollide() {
	a.calcMomentum()
	unitmovex := 1
	unitmovey := 1

	if a.moment.Xaxis < 0 {
		unitmovex = -1
	}
	if a.moment.Yaxis < 0 {
		unitmovey = -1
	}

	absSpdx := int(math.Abs(float64(a.moment.Xaxis) / 10))
	absSpdy := int(math.Abs(float64(a.moment.Yaxis) / 10))
	maxSpd := absSpdx
	if absSpdy > absSpdx {
		maxSpd = absSpdy
	}
	for i := 1; i < maxSpd+1; i++ {
		xcollided := a.directionalCollide(&absSpdx, unitmovex, 0, &a.moment.Xaxis)
		ycollided := a.directionalCollide(&absSpdy, 0, unitmovey, &a.moment.Yaxis)
		if xcollided && ycollided {
			break
		}
	}
}

func (a *acceleratingEnt) directionalCollide(absSpdx *int, unitmovex int, unitmovey int, tozero *int) bool {
	if *absSpdx > 0 {
		*absSpdx--
		checkloc := a.rect.location
		checkloc.x += unitmovex
		checkloc.y += unitmovey
		checkRect := a.rect
		checkRect.refreshShape(checkloc)
		if !checkRect.shape.normalcollides(a.collisionId) {
			a.rect.refreshShape(checkloc)
		} else {
			*absSpdx = 0
			*tozero = 0
			return true
		}
	}
	return false
}
