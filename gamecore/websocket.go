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
	wepBlockers = make(map[*entityid]*shape)
	slashers = make(map[*entityid]*slasher)
	remotePlayers = make(map[*entityid]*slasher)
	enemyControllers = make(map[*entityid]*enemyController)
	mySlasher.deth.hp.CurrentHP = -1
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
				if l.MyPNum == pnum {
					continue found
				}
			}
			eliminate(rm)
			delete(otherPlayers, pnum)
		}

		for _, l := range msg.ll.Locs {
			if _, ok := otherPlayers[l.MyPNum]; !ok {
				log.Println("adding new player")
				newOtherPlayer := &entityid{}
				otherPlayers[l.MyPNum] = newOtherPlayer
				remoteP := addPlayerEntity(newOtherPlayer, location{l.Myloc.X, l.Myloc.Y}, l.Myhealth)
				addRemotePlayer(newOtherPlayer,remoteP)
			}
			res := otherPlayers[l.MyPNum]
			remotePlayers[res].ent.baseloc = remotePlayers[res].ent.rect.location
			remotePlayers[res].ent.endpoint = location{l.Myloc.X, l.Myloc.Y}
			remotePlayers[res].ent.directions = l.Mydir
			remotePlayers[res].ent.moment = l.Mymom
			remotePlayers[res].startangle = l.Myaxe.Startangle
			remotePlayers[res].ent.atkButton = l.Myaxe.Swinging
			if remotePlayers[res].deth.skipHpUpdate > 0 {
				remotePlayers[res].deth.skipHpUpdate--
			} else {
				if l.Myhealth.CurrentHP < remotePlayers[res].deth.hp.CurrentHP {
					remotePlayers[res].deth.redScale = 10
				}
				remotePlayers[res].deth.hp = l.Myhealth
			}
			for _, hitid := range l.Myaxe.IHit {
				if hitid == myPNum {
					mySlasher.deth.redScale = 10
					mySlasher.deth.hp.CurrentHP -= l.Myaxe.Dmg
					if mySlasher.deth.hp.CurrentHP < 1 {
						eliminate(myId)
					}
					break
				}
			}
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
		message.Myhealth = mySlasher.deth.hp

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
