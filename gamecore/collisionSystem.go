package gamecore

import (
	"context"
	"log"
	"math"
	"nhooyr.io/websocket/wsjson"
)

var movers = make(map[*entityid]*acceleratingEnt)
var remoteMovers = make(map[*entityid]*RemoteMover)
var solids = make(map[*entityid]*shape)

//Momentum hi
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
}

func newControlledEntity() *acceleratingEnt {
	c := &acceleratingEnt{}
	c.rect = newRectangle(
		location{50, 50},
		dimens{20, 40},
	)
	c.tracktion = 3
	c.agility = 4
	c.moveSpeed = 100
	return c
}

func addMoveCollider(p *acceleratingEnt, id *entityid) {
	movers[id] = p
	id.systems = append(id.systems, moveCollider)
}
func addRemoteMover(p *RemoteMover, id *entityid) {
	remoteMovers[id] = p
	id.systems = append(id.systems, remoteMover)
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
		p.moment.Yaxis = int(float64(p.moment.Yaxis) * 0.9)
		if int(math.Abs(float64(p.moment.Yaxis)/10)) < 1 {
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
	//fmt.Println("moving at speed:",absSpdx,absSpdy)
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

func collisionSystemWork() {
	for moverid, p := range movers {
		p.moment = calcMomentum(*p)
		moveCollide(p, moverid)
	}
}

func socketReceive() {
	if socketConnection == nil {
		return
	}

	select {
	case msg := <-receiveChan:
		log.Printf("receiveChan: %+v", msg)
		pingFrames = receiveCount
		receiveCount = 1
	found:
		for pnum, rm := range otherPlayers {
			for _, l := range msg.Locs {
				if l.PNum == pnum {
					continue found
				}
			}
			eliminate(rm)
			delete(otherPlayers, pnum)
		}

		for _, l := range msg.Locs {
			if res, ok := otherPlayers[l.PNum]; !ok {
				log.Println("adding new player")
				newOtherPlayer := &entityid{}
				otherPlayers[l.PNum] = newOtherPlayer
				accelEnt := newControlledEntity()
				accelEnt.rect.refreshShape(location{l.Loc.X, l.Loc.Y})
				addHitbox(accelEnt.rect.shape, newOtherPlayer)
				addSolid(accelEnt.rect.shape, newOtherPlayer)
				otherPlay := &RemoteMover{}
				otherPlay.baseloc = accelEnt.rect.location
				otherPlay.endpoint = accelEnt.rect.location
				otherPlay.accelEnt = accelEnt
				addRemoteMover(otherPlay, newOtherPlayer)
				remoteSlasher := newSlasher(accelEnt)
				remoteSlasher.remote = true
				addSlasher(newOtherPlayer, remoteSlasher)
				pDeathable := deathable{}
				pDeathable.currentHP = 6
				pDeathable.maxHP = 6
				pDeathable.deathableShape = accelEnt.rect
				pDeathable.remote = true
				addDeathable(newOtherPlayer, &pDeathable)
			} else {
				diffx := l.Loc.X - remoteMovers[res].accelEnt.rect.location.x
				diffy := l.Loc.Y - remoteMovers[res].accelEnt.rect.location.y
				remoteMovers[res].baseloc = remoteMovers[res].accelEnt.rect.location
				remoteMovers[res].destination = location{diffx / (pingFrames / 2), diffy / (pingFrames / 2)}
				remoteMovers[res].endpoint = location{l.Loc.X, l.Loc.Y}
				remoteMovers[res].accelEnt.directions = l.HisDir
				remoteMovers[res].accelEnt.moment = l.HisMom
				slashers[res].startangle = l.HisAxe.Startangle
				slashers[res].ent.atkButton = l.HisAxe.Swinging
				if l.YouCopped{
					myDeathable.gotHit = true
				}
			}
		}
		message := ServerMessage{}
		message.Myloc = ServerLocation{myAccelEnt.rect.location.x, myAccelEnt.rect.location.y}
		message.Mymom = myAccelEnt.moment
		message.Mydir = myAccelEnt.directions
		messageWep := Weapon{}
		messageWep.Swinging = mySlasher.swangin
		messageWep.Startangle = mySlasher.pivShape.startCount
		var hitlist []string
		for serverid,localid := range otherPlayers{
			for _,hitlocal := range mySlasher.hitsToSend {
				if localid == hitlocal{
					hitlist = append(hitlist,serverid)
				}
			}
		}
		messageWep.IHit = hitlist
		mySlasher.hitsToSend = nil
		message.Myaxe = messageWep
		go func() {
			writeErr := wsjson.Write(context.Background(), socketConnection, message)
			if writeErr != nil {
				log.Println(writeErr)
				closeConn()
				socketConnection = nil
				return
			}
			log.Printf("sent my pos %+v", message)
		}()
		// netbusy = false
	default:
		// if receiveCount < SENDRATE+1 {
		// receiveCount++
		// }
	}
	receiveCount++
}

func remoteMoversWork() {
	for id, p := range remoteMovers {
		if receiveCount > pingFrames+1 {
			p.accelEnt.directions.Down = false
			p.accelEnt.directions.Left = false
			p.accelEnt.directions.Right = false
			p.accelEnt.directions.Up = false
		}

		if receiveCount <= (pingFrames/2)+1 {
			newplace := p.baseloc
			newplace.x += p.destination.x * receiveCount
			newplace.y += p.destination.y * receiveCount
			if receiveCount == (pingFrames/2)+1 {
				newplace = p.endpoint
			}
			checkrect := *p.accelEnt.rect
			checkrect.refreshShape(newplace)
			if !normalcollides(*checkrect.shape, solids, id) {
				p.accelEnt.rect.refreshShape(newplace)
			}
			continue
		}

		p.accelEnt.moment = calcMomentum(*p.accelEnt)
		moveCollide(p.accelEnt, id)
	}
}
