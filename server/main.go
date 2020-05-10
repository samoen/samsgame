package main

import (
	"context"
	"fmt"
	"log"
	"mahgame/gamecore"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"sync"
)

var connections = make(map[*websocket.Conn]gamecore.ServerMessage)

func main() {
	conMutex := sync.Mutex{}
	fmt.Println("server go brr")
	http.Handle("/", http.FileServer(http.Dir(".")))
	//http.HandleFunc("/assets", func(w http.ResponseWriter, r *http.Request) {
	//	http.FileServer(http.Dir("/")).ServeHTTP(w,r)
	//})
	//http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "index.html")
	//})

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conno, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		//conMutex.Lock()
		log.Println("adding connection")
		//connections[conno] = gamecore.ServerMessage{}
		defer func() {
			delete(connections, conno)
			err = conno.Close(websocket.StatusInternalError, "handler defer, removed from connections")
			if err != nil {
				log.Println(err)
			}
		}()
		for {
			var v gamecore.ServerMessage
			err := wsjson.Read(context.Background(), conno, &v)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("received: ", v)
			conMutex.Lock()
			connections[conno] = v
			conMutex.Unlock()
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
				return
			}
			log.Println("sent message: ", toSend)
		}
	})
	http.Handle("/ws", hf)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
