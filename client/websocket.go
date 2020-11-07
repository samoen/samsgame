package main

import (
	"context"
	"fmt"
	"log"
	"mahgame/gamecore"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"os"
)

type sockSelecter struct {
	ll   gamecore.MessageToClient
	sock *websocket.Conn
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
	var v gamecore.MessageToClient
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
			var v gamecore.MessageToClient
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
			var v gamecore.MessageToClient
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
	localAnimals = make(map[*localAnimal]bool)
	remotePlayers = make(map[string]*remotePlayer)
	myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP = -1
	placeMap()
}

func socketReceive() {

	if socketConnection == nil || othersock == nil {
		return
	}

	receiveCount += 1
	receiveDebug = string(append([]byte(receiveDebug), '*'))
	select {
	case msg := <-receiveChan:
		//log.Printf("receiveChan: %+v", msg)
		interpTime = (receiveCount / 2)-1
		if interpTime < 0 {
			interpTime = 0
		}
		//log.Println(interpTime)
		receiveCount = 1
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
				remoteSlasher.rect.refreshShape(location{l.X, l.Y})
				remoteSlasher.deth.hp = hitpoints{l.CurrentHP, l.MaxHP}
				remoteP := &remotePlayer{}
				remoteP.rSlasher = remoteSlasher
				remoteP.servId = l.MyPNum
				remotePlayers[l.MyPNum] = remoteP
			}

			if rp, ok := remotePlayers[l.MyPNum]; ok {
				rp.rSlasher.baseloc = rp.rSlasher.rect.location
				rp.rSlasher.endpoint = location{l.X, l.Y}
				rp.rSlasher.directions = directions{l.Right, l.Down, l.Left, l.Up}
				rp.rSlasher.moment = momentum{l.Xaxis, l.Yaxis}
				rp.rSlasher.atkButton = l.NewSwing
				if l.NewSwing {
					rp.rSlasher.startangle = l.NewSwingAngle
				}else{
					rp.rSlasher.startangle = l.Heading
				}
				rp.rSlasher.swangin = l.Swangin
				if rp.rSlasher.deth.skipHpUpdate > 0 {
					rp.rSlasher.deth.skipHpUpdate--
				} else {
					if l.CurrentHP < rp.rSlasher.deth.hp.CurrentHP {
						rp.rSlasher.deth.redScale = 10
					}
					rp.rSlasher.deth.hp = hitpoints{l.CurrentHP, l.MaxHP}
				}
				for _, hitid := range l.IHit {
					if hitid == myPNum {
						myLocalPlayer.locEnt.lSlasher.deth.redScale = 10
						myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP -= l.Dmg
						if myLocalPlayer.locEnt.lSlasher.deth.hp.CurrentHP < 1 {
							myLocalPlayer.locEnt.lSlasher.addDeathAnim()
						}
						break
					}
					for la, _ := range localAnimals {
						if hitid == myPNum+fmt.Sprintf("%p", la) {
							la.locEnt.lSlasher.deth.redScale = 10
							la.locEnt.lSlasher.deth.hp.CurrentHP -= l.Dmg
							la.checkRemove()
						}
					}
				}
			}
		}
		message := myLocalPlayer.locEnt.toRemoteEnt(msg.ll.YourPNum)

		var animalsToSend []gamecore.EntityData
		for a, _ := range localAnimals {
			animessage := a.locEnt.toRemoteEnt(msg.ll.YourPNum + fmt.Sprintf("%p", a))
			animalsToSend = append(animalsToSend, animessage)
		}
		mts := gamecore.MessageToServer{}
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
