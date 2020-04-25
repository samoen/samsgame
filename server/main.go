package main

import (
	"fmt"
	"log"

	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)
type serverMessage struct{
	Myloc serverLocation `json:"myloc"`
}
type serverLocation struct{
	X int `json:"x"`
	Y int `json:"y"`
}
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
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer func(){
			err = c.Close(websocket.StatusInternalError, "closed from server defer")
			if err != nil {
				log.Println(err)
			}
		}()

		go func(){
			//ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
			//defer cancel()
			for{
				var v serverMessage
				err = wsjson.Read(r.Context(), c, &v)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("received: ",v.Myloc.X,v.Myloc.Y)
			}
		}()
		for{
			toSend := serverMessage{serverLocation{30,30}}
			err = wsjson.Write(r.Context(),c,toSend)
			//_,err = w.Write([]byte("ahoy from server"))
			if err != nil{
				log.Println(err)
				return
			}
			log.Println("sent message: ",toSend)
			time.Sleep(500 * time.Millisecond)
		}

		//err = c.Close(websocket.StatusNormalClosure, "ended normally")

	})
	http.Handle("/ws",hf)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
