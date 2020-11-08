package main

import (
	"github.com/hajimehoshi/ebiten"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"time"
)

func main() {
	images = imagesStruct{}
	images.newImages()

	if err := images.empty.Fill(color.White); err != nil {
		log.Fatal(err)
	}

	myLocalPlayer = localPlayer{}
	myLocalPlayer.locEnt.lSlasher.defaultStats()
	myLocalPlayer.placePlayer()

	for i := 1; i < 10; i++ {
		animal := slasher{}
		animal.defaultStats()
		animal.moveSpeed = 50
		animal.rect.refreshShape(location{i*50 + 50, i * 30})
		la := &localAnimal{}
		la.locEnt.lSlasher = animal
		localAnimals[la] = true
	}

	placeMap()

	tilesAcross := worldWidth / bgTileWidth
	for i := -1; i < tilesAcross+1; i++ {
		for j := -1; j < tilesAcross+1; j++ {
			tileImage:=images.tile1
			passable := true
			if j > tilesAcross-1 || i > tilesAcross-1 || j < 0 || i < 0 {
				tileImage = images.tile2
			} else if 98 < rand.Intn(100) {
				passable = false
				tileImage = images.tile2
			}

			bgBs := baseSprite{tileImage,&ebiten.DrawImageOptions{},0}
			bgtilesNew[location{i,j}] = &backgroundTile{bgBs,passable}
		}
	}

	tileRenderBuffer, _ = ebiten.NewImage(screenWidth,screenHeight,ebiten.FilterDefault)

	go func() {
		time.Sleep(1500 * time.Millisecond)
		connectToServer()
	}()

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("sams cool game")
	ebiten.SetWindowResizable(true)
	//ebiten.SetMaxTPS(35)
	samgame := &SamGame{}

	if err := ebiten.RunGame(samgame); err != nil {
		closeConn()
		log.Fatal(err)
		return
	}
	closeConn()
	log.Println("exited main")
}
