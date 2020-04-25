package main

import (
	"fmt"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

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
			err = c.Close(websocket.StatusInternalError, "closed from defer")
			if err != nil {
				log.Println(err)
			}
		}()


		for{
			//ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
			//defer cancel()

			var v interface{}
			err = wsjson.Read(r.Context(), c, &v)
			if err != nil {
				log.Println(err)
				return
			}
			log.Printf("received: %v", v)
			//_,err = w.Write([]byte("ahoy from server"))

		}

		//err = c.Close(websocket.StatusNormalClosure, "ended normally")

	})
	http.Handle("/ws",hf)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
