package gamecore

import (
	"context"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var socketConnection *websocket.Conn

func closeConn() {
	if socketConnection != nil {
		err := socketConnection.Close(websocket.StatusNormalClosure, "closed from client defer")
		if err != nil {
			log.Println(err)
		}
	}
}

func connectToServer() {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//defer cancel()

	var err error
	socketConnection, _, err = websocket.Dial(context.Background(), "ws://localhost:8080/ws", nil)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			var v LocationList
			err1 := wsjson.Read(context.Background(), socketConnection, &v)
			if err1 != nil {
				log.Println(err1)
				closeConn()
				socketConnection = nil
				return
			}


			//go func(){

			//select {
			//case
			receiveChan <- v
			//:
			//default:
			//}

			//}()
		}
	}()
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
				addPlayerEntity(newOtherPlayer, location{l.Loc.X, l.Loc.Y}, l.ServMessage.Myhealth, false)

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
					myDeathable.hp.CurrentHP -= l.ServMessage.Myaxe.Dmg
					//myDeathable.hp.CurrentHP--
				}
			}
		}
		message := ServerMessage{}
		message.Myloc = ServerLocation{myAccelEnt.rect.location.x, myAccelEnt.rect.location.y}
		message.Mymom = myAccelEnt.moment
		message.Mydir = myAccelEnt.directions
		messageWep := Weapon{}
		messageWep.Dmg = mySlasher.pivShape.damage
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