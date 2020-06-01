package gamecore

import (
	"math"
)

var movers = make(map[*entityid]*acceleratingEnt)
var solids = make(map[*entityid]*shape)

type Momentum struct {
	Xaxis int
	Yaxis int
}

type acceleratingEnt struct {
	rect         *rectangle
	moment       Momentum
	tracktion    float64
	agility      float64
	moveSpeed    float64
	directions   Directions
	atkButton    bool
	lastflip     bool
	ignoreflip   bool
	destination  location
	baseloc      location
	endpoint     location
}

func addMoveCollider(p *acceleratingEnt, id *entityid) {
	movers[id] = p
	id.systems = append(id.systems, moveCollider)
}

func addSolid(s *shape, id *entityid) {
	solids[id] = s
	id.systems = append(id.systems, solidCollider)
}

func calcMomentum(p acceleratingEnt) Momentum {

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
	return p.moment
}

func moveCollide(p *acceleratingEnt, moverid *entityid) {
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
				absSpdx = 0
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
				absSpdy = 0
				ycollided = true
			}
		}

		if xcollided && ycollided {
			break
		}
	}
}
type remoteMoveState int
const (
	interpolating remoteMoveState = iota
	deadreckoning
	momentumOnly

)
const interpTime = 4
const deathreckTime = 4
func collisionSystemWork() {
	rms := interpolating
	if receiveCount > interpTime {
		rms = deadreckoning
	}
	if receiveCount > interpTime + deathreckTime{
		rms = momentumOnly
	}
	for moverid, p := range movers {
		if moverid.remote {
			switch rms {
			case interpolating:
				var newplace location
				if receiveCount == interpTime {
					newplace = p.endpoint
				}else{
					diffx := (p.endpoint.x - p.baseloc.x)/interpTime
					diffy := (p.endpoint.y - p.baseloc.y)/interpTime
					newplace = p.rect.location
					newplace.x += diffx
					newplace.y += diffy
				}
				checkrect := newRectangle(newplace, p.rect.dimens)
				if !normalcollides(*checkrect.shape, solids, moverid) {
					p.rect.refreshShape(newplace)
				}
			case deadreckoning:
				p.moment = calcMomentum(*p)
				moveCollide(p, moverid)
			case momentumOnly:
				//if receiveCount > pingFrames {
					p.directions.Down = false
					p.directions.Left = false
					p.directions.Right = false
					p.directions.Up = false
				//}
				p.moment = calcMomentum(*p)
				moveCollide(p, moverid)
			}

		}
		if !moverid.remote{
			p.moment = calcMomentum(*p)
			moveCollide(p, moverid)
		}
	}
}
