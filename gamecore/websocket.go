package gamecore

import (
	"context"
	"fmt"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var socketConnection *websocket.Conn
var othersock *websocket.Conn

func closeConn() {
	if socketConnection != nil {
		err := socketConnection.Close(websocket.StatusNormalClosure, "mainsock closeconn")
		if err != nil {
			log.Println(err)
		}
	}
	if othersock != nil {
		err := othersock.Close(websocket.StatusNormalClosure, "othersock closeconn")
		if err != nil {
			log.Println(err)
		}
	}
}

func connectToServer() {

	socketURL := "ws://localhost:8080/ws"
	var err error
	socketConnection, _, err = websocket.Dial(context.Background(), socketURL, nil)
	if err != nil {
		log.Println(err)
		return
	}
	var v LocationList
	err1 := wsjson.Read(context.Background(), socketConnection, &v)
	if err1 != nil {
		log.Println(err1)
		closeConn()
		socketConnection = nil
		return
	}
	myPNum = v.YourPNum
	clearEntities()

	socketURL = fmt.Sprintf("ws://localhost:8080/ws?a=%s", myPNum)
	othersock, _, err = websocket.Dial(context.Background(), socketURL, nil)
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
			ss := sockSelecter{}
			ss.sock = socketConnection
			ss.ll = v
			receiveChan <- ss
		}
	}()
	go func() {
		for {
			var v LocationList
			err1 := wsjson.Read(context.Background(), othersock, &v)
			if err1 != nil {
				log.Println(err1)
				closeConn()
				othersock = nil
				return
			}
			ss := sockSelecter{}
			ss.sock = othersock
			ss.ll = v
			receiveChan <- ss
		}
	}()
}

func clearEntities() {
	solids = make(map[*entityid]*shape)
	wepBlockers = make(map[*entityid]*shape)
	slashers = make(map[*entityid]*slasher)
	basicSprites = make(map[*entityid]*baseSprite)
	deathables = make(map[*entityid]*deathable)
	deathables = make(map[*entityid]*deathable)
	enemyControllers = make(map[*entityid]*enemyController)
	hitBoxes = make(map[*entityid]*shape)
	myDeathable.hp.CurrentHP = -1
	placeMap()
}

var myPNum string

func socketReceive() {

	if socketConnection == nil {
		return
	}
	receiveCount += 1
	receiveDebug = string(append([]byte(receiveDebug), '*'))
	select {
	case msg := <-receiveChan:
		log.Printf("receiveChan: %+v", msg)

		pingFrames = receiveCount
		receiveCount = 0
		receiveDebug = ""
	found:
		for pnum, rm := range otherPlayers {
			for _, l := range msg.ll.Locs {
				if l.PNum == pnum {
					continue found
				}
			}
			eliminate(rm)
			delete(otherPlayers, pnum)
		}

		for _, l := range msg.ll.Locs {
			//var remoteent *entityid
			if _, ok := otherPlayers[l.PNum]; !ok {

				log.Println("adding new player")
				newOtherPlayer := &entityid{}
				newOtherPlayer.remote = true
				otherPlayers[l.PNum] = newOtherPlayer
				addPlayerEntity(newOtherPlayer, location{l.Loc.X, l.Loc.Y}, l.ServMessage.Myhealth, false)

			}
			res := otherPlayers[l.PNum]
			//else {
			//diffx := l.Loc.X - movers[res].rect.location.x
			//diffy := l.Loc.Y - movers[res].rect.location.y
			//movers[res].destination = location{diffx / (pingFrames / 2), diffy / (pingFrames / 2)}
			slashers[res].ent.baseloc = slashers[res].ent.rect.location
			slashers[res].ent.endpoint = location{l.Loc.X, l.Loc.Y}
			slashers[res].ent.directions = l.ServMessage.Mydir
			slashers[res].ent.moment = l.ServMessage.Mymom
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
			//}
		}
		message := ServerMessage{}
		message.MyPNum = msg.ll.YourPNum
		message.Myloc = ServerLocation{mySlasher.ent.rect.location.x, mySlasher.ent.rect.location.y}
		message.Mymom = mySlasher.ent.moment
		message.Mydir = mySlasher.ent.directions
		messageWep := Weapon{}
		messageWep.Dmg = mySlasher.pivShape.damage
		messageWep.Swinging = mySlasher.swangSinceSend
		messageWep.Startangle = mySlasher.startangle
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
			writeErr := wsjson.Write(context.Background(), msg.sock, message)
			if writeErr != nil {
				log.Println(writeErr)
				closeConn()
				msg.sock = nil
				return
			}
			log.Printf("sent my pos %+v", message)
		}()
	default:
	}
}
