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
var conMutex = sync.Mutex{}

func main(){
	fmt.Println("server go brr")
	http.Handle("/", http.FileServer(http.Dir(".")))
	//http.HandleFunc("/assets", func(w http.ResponseWriter, r *http.Request) {
	//	http.FileServer(http.Dir("/")).ServeHTTP(w,r)
	//})
	//http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "index.html")
	//})

	hf := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		conno, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		//conMutex.Lock()
		log.Println("adding connection")
		connections[conno] = gamecore.ServerMessage{}
		go func(){
			defer func(){
				err = conno.Close(websocket.StatusInternalError, "couldn't read from socket, removing from connections")
				delete(connections,conno)
			}()
			for{

				var v gamecore.ServerMessage
				err := wsjson.Read(context.Background(), conno, &v)
				if err != nil {
					log.Println(err)
					return
					//continue
				}
				log.Println("received: ",v)
				//conMutex.Lock()
				connections[conno] = v
				//conMutex.Unlock()
				var locs []gamecore.LocWithPNum
				for subcon,loc := range connections{
					if subcon != conno{
						locWithP := gamecore.LocWithPNum{
							Loc: loc.Myloc,
							PNum: fmt.Sprintf("%p",subcon),
							HisMom: loc.Mymom,
							HisDir: loc.Mydir,
						}
						locs = append(locs,locWithP)
					}
				}
				toSend := gamecore.LocationList{Locs:locs}
				err = wsjson.Write(context.Background(),conno,toSend)
				if err != nil{
					log.Println(err)
					return
					//continue
				}
				log.Println("sent message: ",toSend)
			}
		}()
		//conMutex.Unlock()
		//defer func(){
		//	err = conno.Close(websocket.StatusInternalError, "closed from server defer")
		//	if err != nil {
		//		log.Println(err)
		//	}
		//}()
		//err = conno.Close(websocket.StatusNormalClosure, "ended normally")

	})
	//go func(){
	//	for{
	//
	//		conMutex.Lock()
	//		for conno,_ := range connections{
	//			var v gamecore.ServerMessage
	//			err := wsjson.Read(context.Background(), conno, &v)
	//			if err != nil {
	//				log.Println(err)
	//				err = conno.Close(websocket.StatusInternalError, "couldn't read from socket, removing from connections")
	//				delete(connections,conno)
	//				continue
	//			}
	//			log.Println("received: ",v)
	//			connections[conno] = v
	//		}
	//		conMutex.Unlock()
	//	}
	//}()


	//go func(){
	//	for{
	//		//time.Sleep(150 * time.Millisecond)
	//		conMutex.Lock()
	//		wg := sync.WaitGroup{}
	//		for conno,_ := range connections{
	//			conno:=conno
	//			wg.Add(1)
	//			go func(){
	//				var v gamecore.ServerMessage
	//				err := wsjson.Read(context.Background(), conno, &v)
	//				if err != nil {
	//					log.Println(err)
	//					err = conno.Close(websocket.StatusInternalError, "couldn't read from socket, removing from connections")
	//					delete(connections,conno)
	//					return
	//					//continue
	//				}
	//				log.Println("received: ",v)
	//				//conMutex.Lock()
	//				connections[conno] = v
	//				//conMutex.Unlock()
	//				wg.Done()
	//			}()
	//		}
	//		wg.Wait()
	//
	//
	//		for conno,_ := range connections{
	//			var locs []gamecore.LocWithPNum
	//			for subcon,loc := range connections{
	//				if subcon != conno{
	//				locWithP := gamecore.LocWithPNum{
	//					Loc: loc.Myloc,
	//					PNum: fmt.Sprintf("%p",subcon),
	//					HisMom: loc.Mymom,
	//					HisDir: loc.Mydir,
	//				}
	//				locs = append(locs,locWithP)
	//				}
	//			}
	//			toSend := gamecore.LocationList{Locs:locs}
	//			err := wsjson.Write(context.Background(),conno,toSend)
	//			if err != nil{
	//				log.Println(err)
	//				err = conno.Close(websocket.StatusInternalError, "couldn't write to socket, removing from connections")
	//				delete(connections,conno)
	//				continue
	//			}
	//			log.Println("sent message: ",toSend)
	//		}
	//		conMutex.Unlock()
	//		//time.Sleep(150 * time.Millisecond)
	//	}
	//}()
	http.Handle("/ws",hf)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
