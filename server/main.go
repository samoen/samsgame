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
	sm        gamecore.EntityData
	animals []gamecore.EntityData
	busy      bool
	ping      time.Duration
}

func updateServerEnt(mapid *bool, conno *websocket.Conn) error {
	timer1 := time.NewTimer(300 * time.Millisecond)
	var locs []gamecore.EntityData
	conMutex.Lock()
	for subcon, loc := range connections {
		if subcon == mapid {
			continue
		}
		if loc.sm.X == 0 {
			continue
		}
		if loc.sm.CurrentHP > 0 {
			locs = append(locs, loc.sm)
		}

		for _,an := range loc.animals{
			locs = append(locs, an)
		}
	}
	conMutex.Unlock()
	toSend := gamecore.MessageToClient{}
	toSend.Locs = locs
	toSend.YourPNum = fmt.Sprintf("%p", mapid)
	err := wsjson.Write(context.Background(), conno, toSend)
	if err != nil {
		log.Println(err)
		return err
	}
	//log.Println("sent message: ", toSend)

	var v gamecore.MessageToServer
	err = wsjson.Read(context.Background(), conno, &v)
	if err != nil {
		log.Println(err)
		return err
	}
	//log.Println("received: ", v)
	conMutex.Lock()
	if se, ok := connections[mapid]; ok {
		se.sm = v.MyData
		se.animals = v.MyAnimals
	}
	conMutex.Unlock()
	<-timer1.C
	return nil
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
		if err != nil {
			log.Fatal(err)
			return
		}
		q := r.URL.Query()
		param := q.Get("a")
		log.Println(param)
		found := false
		var otherconn *websocket.Conn
		var myid *bool
		conMutex.Lock()
		for id, serveEnt := range connections {
			if fmt.Sprintf("%p", id) == param {
				log.Println("found existing")
				found = true
				serveEnt.otherconn = conno
				otherconn = serveEnt.entconn
				myid = id
			}
		}
		conMutex.Unlock()

		if !found {
			t := true
			mapid := &t
			//log.Println("accepted connection")
			servEnt := &ServerEntity{}
			servEnt.ping = 400
			servEnt.entconn = conno
			servEnt.sm.CurrentHP = -2
			conMutex.Lock()
			connections[mapid] = servEnt
			conMutex.Unlock()

			toSend := gamecore.MessageToClient{}
			toSend.YourPNum = fmt.Sprintf("%p", mapid)
			err := wsjson.Write(context.Background(), conno, toSend)
			if err != nil {
				log.Println(err)
			}
			return
		}
		pingmeasure := time.Duration(500 * time.Millisecond)
		wg2 := sync.WaitGroup{}
		exit1 := false
		exit2 := false
		for {
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				measure := time.Now()
				err := updateServerEnt(myid, conno)
				if err != nil {
					exit1 = true
				}
				pingmeasure = time.Since(measure)
				wg.Done()
			}()
			halfping := pingmeasure / 2
			time.Sleep(halfping)
			wg2.Wait()
			if exit2 == true {
				break
			}
			wg2.Add(1)
			go func() {
				err := updateServerEnt(myid, otherconn)
				if err != nil {
					exit2 = true
				}
				wg2.Done()
			}()
			time.Sleep(halfping)
			wg.Wait()
			if exit1 == true || exit2 == true {
				log.Println("exit called")
				conMutex.Lock()
				delete(connections, myid)
				conMutex.Unlock()
				err := conno.Close(websocket.StatusNormalClosure, "server closing a secondary")
				if err != nil {
					log.Println(err)
				}
				err = otherconn.Close(websocket.StatusNormalClosure, "server closing a primary")
				if err != nil {
					log.Println(err)
				}
				break
			}
		}

	})
	m.Handle("/ws", hf)
	servah := http.Server{Addr: ":8080", Handler: m}
	go func() {
		if err := servah.ListenAndServe(); err != nil {
			log.Fatal("ListenAndServe: ", err)
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
