package main

import (
	_ "image/png"
	"log"
	"mahgame/gamecore"

	"github.com/hajimehoshi/ebiten"
)





func main() {

	//emptyImage.Fill(color.White)

	//gamecore.initEntities()

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowSize(gamecore.ScreenWidth, gamecore.ScreenHeight)
	ebiten.SetWindowTitle("sams cool game")
	ebiten.SetWindowResizable(true)

	samgame := &gamecore.SamGame{}

	if err := ebiten.RunGame(samgame); err != nil {
		log.Fatal(err)
	}
}


