package gamecore

import (
	"context"
	"github.com/hajimehoshi/ebiten"
	"log"
	"math"
	"nhooyr.io/websocket/wsjson"
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
	lastflip   bool
	ignoreflip bool
	destination location
	baseloc     location
	endpoint    location
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
		interpolating := false
		if moverid.remote{
			if receiveCount > pingFrames+1 {
				p.directions.Down = false
				p.directions.Left = false
				p.directions.Right = false
				p.directions.Up = false
			}

			if receiveCount <= (pingFrames/2)+1 {
				interpolating = true
				newplace := p.baseloc
				newplace.x += p.destination.x * receiveCount
				newplace.y += p.destination.y * receiveCount
				if receiveCount == (pingFrames/2)+1 {
					newplace = p.endpoint
				}
				checkrect := *p.rect
				checkrect.refreshShape(newplace)
				if !normalcollides(*checkrect.shape, solids, moverid) {
					p.rect.refreshShape(newplace)
				}
			}
		}

		if !interpolating{
			p.moment = calcMomentum(*p)
			moveCollide(p, moverid)
		}
	}

}

func updateSprites(){
	offset:=renderOffset()
	for _, bs := range basicSprites {
		bs.bOps.ColorM.Reset()
		bs.bOps.GeoM.Reset()
	}
	for pid, bs := range basicSprites {

		if p,ok:=movers[pid];ok{
			if !p.ignoreflip {
				if p.directions.Left && !p.directions.Right {
					p.lastflip = true
				}
				if p.directions.Right && !p.directions.Left {
					p.lastflip = false}
			}

			if p.lastflip {
				invertGeom := ebiten.GeoM{}
				invertGeom.Scale(-1, 1)
				invertGeom.Translate(float64(p.rect.dimens.width), 0)
				bs.bOps.GeoM.Add(invertGeom)
			}

			scaleToDimension(p.rect.dimens, bs.sprite, bs.bOps)
			cameraShift(p.rect.location, offset, bs.bOps)
		}
		if mDeathable, ok := deathables[pid]; ok {
			bs.bOps.ColorM.Translate(float64(mDeathable.redScale), 0, 0, 0)
			if subbs, ok := basicSprites[mDeathable.hBarid]; ok {
				healthbarlocation := location{mDeathable.deathableShape.location.x, mDeathable.deathableShape.location.y - 10}
				healthbardimenswidth := mDeathable.hp.CurrentHP * mDeathable.deathableShape.dimens.width / mDeathable.hp.MaxHP
				scaleToDimension(dimens{healthbardimenswidth, 5}, emptyImage, subbs.bOps)
				cameraShift(healthbarlocation, offset, subbs.bOps)
			}
		}
		if bot,ok:=slashers[pid];ok{
			if bs, ok := basicSprites[bot.wepid]; ok {
				_, imH := bs.sprite.Size()
				ownerCenter := rectCenterPoint(*bot.ent.rect)
				cameraShift(ownerCenter, offset, bs.bOps)
				addOp := ebiten.GeoM{}
				hRatio := float64(bot.pivShape.bladeLength+bot.pivShape.bladeLength/4) / float64(imH)
				addOp.Scale(hRatio, hRatio)
				addOp.Translate(-float64(bot.ent.rect.dimens.width)/2, 0)
				addOp.Rotate(bot.pivShape.animationCount - (math.Pi / 2))
				bs.bOps.GeoM.Add(addOp)
			}
		}
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
				newOtherPlayer.remote = true
				otherPlayers[l.PNum] = newOtherPlayer
				addPlayerEntity(newOtherPlayer,location{l.Loc.X, l.Loc.Y},l.ServMessage.Myhealth,false)

			} else {
				diffx := l.Loc.X - movers[res].rect.location.x
				diffy := l.Loc.Y - movers[res].rect.location.y
				movers[res].baseloc = movers[res].rect.location
				movers[res].destination = location{diffx / (pingFrames / 2), diffy / (pingFrames / 2)}
				movers[res].endpoint = location{l.Loc.X, l.Loc.Y}
				movers[res].directions = l.ServMessage.Mydir
				movers[res].moment = l.ServMessage.Mymom
				slashers[res].startangle = l.ServMessage.Myaxe.Startangle
				slashers[res].ent.atkButton = l.ServMessage.Myaxe.Swinging
				if deathables[res].skipHpUpdate > 0 {
					deathables[res].skipHpUpdate--
				} else {
					if l.ServMessage.Myhealth.CurrentHP < deathables[res].hp.CurrentHP {
						deathables[res].gotHit = true
					}
					deathables[res].hp = l.ServMessage.Myhealth
				}

				if l.YouCopped {
					myDeathable.gotHit = true
					myDeathable.hp.CurrentHP--
				}
			}
		}
		message := ServerMessage{}
		message.Myloc = ServerLocation{myAccelEnt.rect.location.x, myAccelEnt.rect.location.y}
		message.Mymom = myAccelEnt.moment
		message.Mydir = myAccelEnt.directions
		messageWep := Weapon{}
		messageWep.Swinging = mySlasher.swangSinceSend
		messageWep.Startangle = mySlasher.pivShape.startCount
		var hitlist []string
		for serverid, localid := range otherPlayers {
			for _, hitlocal := range mySlasher.hitsToSend {
				if localid == hitlocal {
					hitlist = append(hitlist, serverid)
				}
			}
		}
		messageWep.IHit = hitlist
		message.Myaxe = messageWep
		message.Myhealth = myDeathable.hp

		mySlasher.hitsToSend = nil
		mySlasher.swangSinceSend = false

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
	default:
	}
	receiveCount++
}