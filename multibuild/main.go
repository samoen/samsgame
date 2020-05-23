package main

import (
	"log"
	"os/exec"
	"time"
)

//
//import (
//	"mahgame/client"
//)
//
func main() {
	//server.Servite()
	//t :=time.NewTimer(1000 * time.Millisecond)
	//<-t.C
	//client.DoEet()
	//t =time.NewTimer(1000 *time.Millisecond)
	//<-t.C
	//client.DoEet()
	var err error
	err = exec.Command("go", "build", "mahgame/server").Run()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err = exec.Command("server.exe").Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(1000 * time.Millisecond)
	err = exec.Command("go", "build", "mahgame/client").Run()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1000 * time.Millisecond)
	go func() {
		err = exec.Command("client.exe").Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(3000 * time.Millisecond)
	//go func() {
	err = exec.Command("client.exe").Run()
	if err != nil {
		log.Fatal(err)
	}
	//}()

}
