package gamecore

import (
	"context"
	"fmt"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"os"
)

type sockSelecter struct {
	ll   LocationList
	sock *websocket.Conn
}

type ServerMessage struct {
	Myloc    ServerLocation
	Mymom    Momentum
	Mydir    Directions
	Myaxe    Weapon
	Myhealth Hitpoints
	MyPNum   string
}

type MessageToServer struct {
	MyData    ServerMessage
	MyAnimals []ServerMessage
}

type Weapon struct {
	Swinging   bool
	Startangle float64
	IHit       []string
	Dmg        int
}

type LocationList struct {
	Locs     []ServerMessage
	YourPNum string
}

type ServerLocation struct {
	X int
	Y int
}

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

var host = "localhost"

func connectToServer() {
	args := os.Args
	if len(args) > 1 {
		host = args[1]
	}
	socketURL := "ws://" + host + ":8080/ws"
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

	socketURL = fmt.Sprintf("ws://"+host+":8080/ws?a=%s", myPNum)
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
				othersock = nil
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
				socketConnection = nil
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
	wepBlockers = make(map[*shape]bool)
	slashers = make(map[*localAnimal]bool)
	remotePlayers = make(map[string]*remotePlayer)
	myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP = -1
	placeMap()
	animal := slasher{}
	animal.defaultStats()
	animal.ent.moveSpeed = 50
	animal.ent.rect.refreshShape(location{70 + 50, 30})
	la := &localAnimal{}
	la.locEnt.lSlasher = animal
	slashers[la] = true
}

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
		for rpid, rp := range remotePlayers {
			for _, l := range msg.ll.Locs {
				if l.MyPNum == rp.servId {
					continue found
				}
			}
			delete(remotePlayers, rpid)
			rp.rSlasher.addDeathAnim()
		}

		for _, l := range msg.ll.Locs {
			if _, ok := remotePlayers[l.MyPNum]; !ok {
				log.Println("adding new player")
				remoteSlasher := slasher{}
				remoteSlasher.defaultStats()
				remoteSlasher.ent.rect.refreshShape(location{l.Myloc.X, l.Myloc.Y})
				remoteSlasher.deth.hp = l.Myhealth
				remoteP := &remotePlayer{}
				remoteP.rSlasher = remoteSlasher
				remoteP.servId = l.MyPNum
				remotePlayers[l.MyPNum] = remoteP
			}
			if rp, ok := remotePlayers[l.MyPNum]; ok {
				rp.rSlasher.ent.baseloc = rp.rSlasher.ent.rect.location
				rp.rSlasher.ent.endpoint = location{l.Myloc.X, l.Myloc.Y}
				rp.rSlasher.ent.directions = l.Mydir
				rp.rSlasher.ent.moment = l.Mymom
				rp.rSlasher.startangle = l.Myaxe.Startangle
				rp.rSlasher.atkButton = l.Myaxe.Swinging
				if rp.rSlasher.deth.skipHpUpdate > 0 {
					rp.rSlasher.deth.skipHpUpdate--
				} else {
					if l.Myhealth.CurrentHP < rp.rSlasher.deth.hp.CurrentHP {
						rp.rSlasher.deth.redScale = 10
					}
					rp.rSlasher.deth.hp = l.Myhealth
				}
				for _, hitid := range l.Myaxe.IHit {
					if hitid == myPNum {
						myLocalPlayer.locEnt.lSlasher.deth.redScale = 10
						myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP -= l.Myaxe.Dmg
						if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP < 1 {
							myLocalPlayer.locEnt.lSlasher.addDeathAnim()
						}
						break
					}
					for la, _ := range slashers {
						if hitid == myPNum+fmt.Sprintf("%p", la) {
							la.locEnt.lSlasher.deth.redScale = 10
							la.locEnt.lSlasher.deth.hp.CurrentHP -= l.Myaxe.Dmg
							la.checkRemove()
						}
					}
				}
			}
		}
		message := ServerMessage{}
		message.MyPNum = msg.ll.YourPNum
		message.Myloc = ServerLocation{myLocalPlayer.locEnt.lSlasher.ent.rect.location.x, myLocalPlayer.locEnt.lSlasher.ent.rect.location.y}
		message.Mymom = myLocalPlayer.locEnt.lSlasher.ent.moment
		message.Mydir = myLocalPlayer.locEnt.lSlasher.ent.directions
		messageWep := Weapon{}
		messageWep.Dmg = myLocalPlayer.locEnt.lSlasher.pivShape.damage
		messageWep.Swinging = myLocalPlayer.locEnt.lSlasher.swangSinceSend
		messageWep.Startangle = myLocalPlayer.locEnt.lSlasher.startangle
		messageWep.IHit = myLocalPlayer.locEnt.hitsToSend
		message.Myaxe = messageWep
		message.Myhealth = myLocalPlayer.locEnt.lSlasher.deth.hp

		myLocalPlayer.locEnt.hitsToSend = nil
		myLocalPlayer.locEnt.lSlasher.swangSinceSend = false

		var animalsToSend []ServerMessage
		for a, _ := range slashers {
			animessage := ServerMessage{}
			animessage.MyPNum = msg.ll.YourPNum + fmt.Sprintf("%p", a)
			animessage.Myloc = ServerLocation{a.locEnt.lSlasher.ent.rect.location.x, a.locEnt.lSlasher.ent.rect.location.y}
			animessage.Mymom = a.locEnt.lSlasher.ent.moment
			animessage.Mydir = a.locEnt.lSlasher.ent.directions
			messageWep := Weapon{}
			messageWep.Dmg = a.locEnt.lSlasher.pivShape.damage
			messageWep.Swinging = a.locEnt.lSlasher.swangSinceSend
			messageWep.Startangle = a.locEnt.lSlasher.startangle
			messageWep.IHit = a.locEnt.hitsToSend
			animessage.Myaxe = messageWep
			animessage.Myhealth = a.locEnt.lSlasher.deth.hp

			a.locEnt.hitsToSend = nil
			a.locEnt.lSlasher.swangSinceSend = false
			animalsToSend = append(animalsToSend, animessage)
		}
		mts := MessageToServer{}
		mts.MyData = message
		mts.MyAnimals = animalsToSend

		go func() {
			writeErr := wsjson.Write(context.Background(), msg.sock, mts)
			if writeErr != nil {
				log.Println(writeErr)
				closeConn()
				msg.sock = nil
				return
			}
			//log.Printf("sent my pos %+v", message)
		}()
	default:
	}
}
