package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"mahgame/gamecore"
	"net/http"
	"os"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var connections = make(map[*bool]*ServerEntity)

type ServerEntity struct {
	entconn   *websocket.Conn
	otherconn *websocket.Conn
	sm        gamecore.ServerMessage
	busy      bool
}

func updateServerEnt(mapid *bool, conno *websocket.Conn) {
	closeit := func() {
		log.Println("removeConn called")
		conMutex.Lock()
		delete(connections, mapid)
		conMutex.Unlock()
		err := conno.Close(websocket.StatusInternalError, "sam closing")
		if err != nil {
			log.Println(err)
		}
	}
	timer1 := time.NewTimer(166 * time.Millisecond)
	var locs []gamecore.LocWithPNum
	for subcon, loc := range connections {
		if subcon == mapid {
			continue
		}
		if loc.sm.Myloc.X == 0 {
			continue
		}
		if loc.sm.Myhealth.CurrentHP < 1 {
			continue
		}
		locWithP := gamecore.LocWithPNum{}
		locWithP.Loc = loc.sm.Myloc
		locWithP.PNum = fmt.Sprintf("%p", subcon)
		locWithP.ServMessage = loc.sm
		locWithP.ServMessage.Myaxe.IHit = nil
		for _, hitid := range loc.sm.Myaxe.IHit {
			if hitid == fmt.Sprintf("%p", mapid) {
				locWithP.YouCopped = true
			}
		}

		locs = append(locs, locWithP)
	}
	toSend := gamecore.LocationList{}
	toSend.Locs = locs
	toSend.YourPNum = fmt.Sprintf("%p", mapid)
	err := wsjson.Write(context.Background(), conno, toSend)
	if err != nil {
		log.Println(err)
		closeit()
		return
	}
	//log.Println("sent message: ", toSend)

	var v gamecore.ServerMessage
	err = wsjson.Read(context.Background(), conno, &v)
	if err != nil {
		log.Println(err)
		closeit()
		return
	}
	//log.Println("received: ", v)
	conMutex.Lock()
	connections[mapid].sm = v
	conMutex.Unlock()
	<-timer1.C
}

var conMutex = sync.Mutex{}

func main() {
	log.SetOutput(os.Stdout)

	log.Println("server go brr")

	m := http.NewServeMux()

	fs := http.FileServer(http.Dir("./assets"))
	m.Handle("/assets/", http.StripPrefix("/assets/", fs))

	wfs2 := http.FileServer(http.Dir("./website/."))
	m.Handle("/", wfs2)

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conno, err := websocket.Accept(w, r, nil)
		q := r.URL.Query()
		param := q.Get("a")
		log.Println(param)
		if err != nil {
			log.Fatal(err)
			return
		}
		conMutex.Lock()
		for id, serveEnt := range connections {
			if fmt.Sprintf("%p", id) == param {
				fmt.Println("found existing")
				serveEnt.otherconn = conno
				conMutex.Unlock()
				return
			}
		}
		conMutex.Unlock()
		t := true
		mapid := &t
		//log.Println("accepted connection")
		servEnt := &ServerEntity{}
		servEnt.entconn = conno
		servEnt.sm.Myhealth.CurrentHP = -2
		conMutex.Lock()
		connections[mapid] = servEnt
		conMutex.Unlock()
	})
	m.Handle("/ws", hf)
	servah := http.Server{Addr: ":8080", Handler: m}
	go func() {
		if err := servah.ListenAndServe(); err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	go func() {
		for {
			conMutex.Lock()
			for id, serveEnt := range connections {
				conMutex.Unlock()
				id := id
				serveEnt := serveEnt
				if !serveEnt.busy {
					//conMutex.Lock()
					serveEnt.busy = true
					//conMutex.Unlock()
					go func() {
						updateServerEnt(id, serveEnt.entconn)
						//updateServerEnt(id, serveEnt.otherconn)
						//conMutex.Lock()
						serveEnt.busy = false
						//conMutex.Unlock()
					}()
				}
				conMutex.Lock()
			}
			conMutex.Unlock()
			//time.Sleep(100 * time.Millisecond)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			err := servah.Shutdown(context.Background())
			if err != nil {
				log.Fatal(err)
				return
			}
			log.Println("server stopped")
			break
		}
	}

	if scanner.Err() != nil {
		log.Fatal("scannah error")
	}
}
