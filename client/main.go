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
			ttype := blank
			if j > tilesAcross-1 || i > tilesAcross-1 || j < 0 || i < 0 {
				ttype = offworld
				tileImage = images.empty
			} else if 90 < rand.Intn(100) {
				ttype = rocky
				tileImage = images.tile2
			}
			bgl := &bgLoading{}
			bgl.tiletyp = ttype
			bgl.ops = &ebiten.DrawImageOptions{}
			bgtiles[location{i, j}] = bgl
			bgtilesNew[location{i,j}] = &baseSprite{tileImage,&ebiten.DrawImageOptions{},0}
		}
	}

	ttshapes[blank] = shape{lines: []line{line{location{180, 5}, location{140, 60}}}}
	ttshapes[rocky] = shape{lines: []line{line{location{80, 20}, location{80, 120}}}}

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
