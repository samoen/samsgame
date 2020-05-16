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

var connections = make(map[*websocket.Conn]gamecore.ServerMessage)

func main() {
	log.SetOutput(os.Stdout)
	conMutex := sync.Mutex{}
	log.Println("server go brr")

	//http.HandleFunc("/assets", func(w http.ResponseWriter, r *http.Request) {
	//	http.FileServer(http.Dir("/")).ServeHTTP(w,r)
	//})
	//http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "index.html")
	//})

	m := http.NewServeMux()
	m.Handle("/", http.FileServer(http.Dir(".")))
	servah := http.Server{Addr: ":8080", Handler: m}

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conno, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		//conMutex.Lock()
		log.Println("accepted connection")
		//connections[conno] = gamecore.ServerMessage{}
		removeConn := func() {
			delete(connections, conno)
			err = conno.Close(websocket.StatusInternalError, "removed from connections")
			if err != nil {
				log.Println(err)
			}
		}
		for {
			timer1 := time.NewTimer(166 * time.Millisecond)
			var locs []gamecore.LocWithPNum
			for subcon, loc := range connections {
				if subcon != conno && loc.Myloc.X != 0 {
					locWithP := gamecore.LocWithPNum{
						Loc:    loc.Myloc,
						PNum:   fmt.Sprintf("%p", subcon),
						HisMom: loc.Mymom,
						HisDir: loc.Mydir,
					}
					locs = append(locs, locWithP)
				}
			}
			toSend := gamecore.LocationList{Locs: locs}
			err = wsjson.Write(context.Background(), conno, toSend)
			if err != nil {
				log.Println(err)
				removeConn()
				return
			}
			log.Println("sent message: ", toSend)

			var v gamecore.ServerMessage
			err := wsjson.Read(context.Background(), conno, &v)
			if err != nil {
				log.Println(err)
				removeConn()
				return
			}
			log.Println("received: ", v)
			conMutex.Lock()
			connections[conno] = v
			conMutex.Unlock()
			<-timer1.C

		}
	})
	// http.Handle("/ws", hf)
	m.Handle("/ws", hf)
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal("ListenAndServe: ", err)
	// }
	go func() {
		if err := servah.ListenAndServe(); err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			servah.Shutdown(context.Background())
			log.Println("server stopped")
			break
		}
	}

	if scanner.Err() != nil {
		// handle error.
		log.Fatal("scannah error:")
	}
}
