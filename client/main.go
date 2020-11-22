package main

import (
	"github.com/hajimehoshi/ebiten"
	"image/color"
	_ "image/png"
	"log"
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
	initTiles()
	prepMapChunks()

	go func() {
		time.Sleep(1500 * time.Millisecond)
		connectToServer()
	}()

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("sams cool game")
	ebiten.SetWindowResizable(true)
	ebiten.SetMaxTPS(40)
	samgame := &SamGame{}

	if err := ebiten.RunGame(samgame); err != nil {
		closeConn()
		log.Fatal(err)
		return
	}
	closeConn()
	log.Println("exited main")
}

func prepMapChunks() {
	stitchedChunks, _ = ebiten.NewImage(screenWidth, screenHeight, ebiten.FilterDefault)
	imW, imH := images.tile1.Size()
	for p := 0; p < worldWidth/chunkWidth; p++ {
		for k := 0; k < worldWidth/chunkWidth; k++ {
			//minichunk, _ := ebiten.NewImage(chunkWidth, chunkWidth, ebiten.FilterDefault)
			minichunk, _ := ebiten.NewImage(imW*tilesperChunk, imH*tilesperChunk, ebiten.FilterDefault)
			for i := 0; i <= tilesperChunk; i++ {
				for j := 0; j <= tilesperChunk; j++ {
					if im, ok := bgtilesNew[location{(tilesperChunk * p) + i, (tilesperChunk * k) + j}]; ok {
						opies := &ebiten.DrawImageOptions{}

						opies.GeoM.Translate(float64(i*imW), float64(j*imH))
						//opies.GeoM.Translate(float64(i*bgTileWidth), float64(j*bgTileWidth))
						//scaleToDimension(dimens{bgTileWidth + 1, bgTileWidth + 1}, im.sprite, opies, false)

						if err := minichunk.DrawImage(im.sprite, opies); err != nil {
							log.Fatal(err)
						}
					}
				}
			}
			mapChunks[location{p, k}] = minichunk
		}
	}
	//testbuf,_ := ebiten.NewImage(screenWidth,screenHeight,ebiten.FilterDefault)
	//go func() {
	//	for k,v := range mapChunks{
	//		ops := &ebiten.DrawImageOptions{}
	//		bs := &baseSprite{v, ops, 0}
	//		fullRenderOp(bs, location{k.x * chunkWidth, k.y * chunkWidth}, false, dimens{chunkWidth, chunkWidth}, 0, 0)
	//		if err := stitchedChunks.DrawImage(v, bs.bOps); err != nil {
	//			log.Fatal(err)
	//		}
	//	}
	//}()
}
