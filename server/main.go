package main

import (
	"fmt"
	"log"
	"net/http"
)

func main(){
	fmt.Println("server go brr")
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("assets"))))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
