package gamecore

import (
	"math"
)

var movers = make(map[*entityid]*acceleratingEnt)
var remoteMovers = make(map[*entityid]*acceleratingEnt)
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
	destination location
	baseloc location
	endpoint location
}

func newControlledEntity() *acceleratingEnt {
	c := &acceleratingEnt{
		newRectangle(
			location{1, 1},
			dimens{20, 40},
		),
		Momentum{},
		3,
		4,
		100,
		Directions{},
		false,
		true,
		false,
		location{},
		location{},
		location{},
	}
	return c
}

func addMoveCollider(p *acceleratingEnt, id *entityid) {
	movers[id] = p
	id.systems = append(id.systems, moveCollider)
}
func addRemoteMover(p *acceleratingEnt, id *entityid) {
	remoteMovers[id] = p
	id.systems = append(id.systems, remoteMover)
}

func addSolid(s *shape, id *entityid) {
	solids[id] = s
	id.systems = append(id.systems, solidCollider)
}

const DEADRECKON = 10
var lagcompcount = DEADRECKON

func calcMomentum(p acceleratingEnt)Momentum{

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
		p.moment.Xaxis = int(float64(p.moment.Xaxis)*0.9)
		if	int(math.Abs(float64(p.moment.Xaxis)/10))<1{
			p.moment.Xaxis = 0
		}
		//p.moment.Xaxis += int(float64(-unitmovex)*(p.tracktion))
		//if math.Abs(float64(p.moment.Xaxis))<2{
		//	p.moment.Xaxis = 0
		//}
		//if unitmovex>0{
		//	if p.moment.Xaxis<0{
		//		p.moment.Xaxis = 0
		//	}
		//}else{
		//	if p.moment.Xaxis>0{
		//		p.moment.Xaxis = 0
		//	}
		//}
	}
	if !movedy {
		p.moment.Yaxis = int(float64(p.moment.Yaxis)*0.9)
		if	int(math.Abs(float64(p.moment.Yaxis)/10))<1{
			p.moment.Yaxis = 0
		}
		//p.moment.Yaxis += int(float64(-unitmovey)*(p.tracktion))
		//if math.Abs(float64(p.moment.Yaxis))<2{
		//	p.moment.Yaxis = 0
		//}
		//if unitmovey>0{
		//	if p.moment.Yaxis<0{
		//		p.moment.Yaxis = 0
		//	}
		//}else{
		//	if p.moment.Yaxis>0{
		//		p.moment.Yaxis = 0
		//	}
		//}
	}
	if p.moment.Xaxis < 0 {
		unitmovex = -1
	}
	if p.moment.Yaxis < 0 {
		unitmovey = -1
	}

	magnitude := math.Sqrt(math.Pow(float64(p.moment.Xaxis), 2) + math.Pow(float64(p.moment.Yaxis), 2))
	if magnitude > p.moveSpeed {
		if math.Abs(float64(p.moment.Xaxis))>p.moveSpeed*0.707{
			p.moment.Xaxis = int(p.moveSpeed * 0.707 * float64(unitmovex))
		}
		if math.Abs(float64(p.moment.Yaxis))>p.moveSpeed*0.707{
			p.moment.Yaxis = int(p.moveSpeed * 0.707 * float64(unitmovey))
		}
	}
	return p.moment
}

func moveCollide(p *acceleratingEnt, moverid *entityid){
	unitmovex := 1
	unitmovey := 1

	if p.moment.Xaxis < 0 {
		unitmovex = -1
	}
	if p.moment.Yaxis < 0 {
		unitmovey = -1
	}

	absSpdx := int(math.Abs(float64(p.moment.Xaxis)/10))
	absSpdy := int(math.Abs(float64(p.moment.Yaxis)/10))
	//fmt.Println("moving at speed:",absSpdx,absSpdy)
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

func collisionSystemWork() {
	for moverid, p := range movers {
		p.moment = calcMomentum(*p)
		moveCollide(p,moverid)
	}
}

func remoteMoversWork() {
	for id, p := range remoteMovers {

		if receiveCount<=(SENDRATE/2)+1{
			newplace := p.baseloc
			newplace.x += p.destination.x*receiveCount
			newplace.y += p.destination.y*receiveCount
			if receiveCount == (SENDRATE/2)+1{
				newplace = p.endpoint
			}
			checkrect := *p.rect
			checkrect.refreshShape(newplace)
			if !normalcollides(*checkrect.shape,solids,id){
				p.rect.refreshShape(newplace)
			}
			continue
		}

		p.moment = calcMomentum(*p)
		moveCollide(p,id)


		//newLocx := p.rect.location.x
		//newLocx+=int(p.moment.Xaxis/10)
		//checkrect := *p.rect
		//checkrect.refreshShape(location{newLocx,p.rect.location.y})
		//if !normalcollides(*checkrect.shape,solids,id){
		//	p.rect.refreshShape(checkrect.location)
		//}else{
		//	p.moment.Xaxis = 0
		//}
		//newLocy := p.rect.location
		//newLocy.y+=int(p.moment.Yaxis/10)
		//checkrecty := *p.rect
		//checkrecty.refreshShape(newLocy)
		//if !normalcollides(*checkrecty.shape,solids,id){
		//	p.rect.refreshShape(checkrecty.location)
		//}else{
		//	p.moment.Yaxis = 0
		//}
	}
}