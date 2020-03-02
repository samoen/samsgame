package main

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten"
)

func initEntities() {
	accelplayer := acceleratingEnt{
		&playerent{
			newRectangle(
				location{1, 1},
				dimens{20, 20},
			),
			moveSpeed{6, 6},
			directions{},
			func() directions {
				return directions{
					ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight),
					ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown),
					ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft),
					ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp),
				}
			},
		},
		momentum{},
		0.4,
		0.4,
	}

	playerSlasher := slasher{
		accelplayer.ent,
		false,
		false,
		&shape{
			[]line{line{}},
		},
		directions{},
	}

	playerMoveSystem.addBot(accelplayer.ent)
	renderingSystem.addShape(accelplayer.ent.rectangle.shape)
	renderingSystem.CenterOn = accelplayer.ent.rectangle
	collideSystem.addEnt(&accelplayer)
	slashSystem.slashers = append(slashSystem.slashers, &playerSlasher)

	for i := 1; i < 50; i++ {
		moveEnemy := &acceleratingEnt{
			&playerent{
				newRectangle(
					location{
						i * 30,
						1,
					},
					dimens{20, 20},
				),
				moveSpeed{9, 9},
				directions{},
				func() directions {
					return directions{
						rand.Intn(2) == 0,
						rand.Intn(2) == 0,
						rand.Intn(2) == 0,
						rand.Intn(2) == 0,
					}
				},
			},
			momentum{},
			0.4,
			0.4,
		}
		renderingSystem.addShape(moveEnemy.ent.rectangle.shape)
		collideSystem.addEnt(moveEnemy)
		botsMoveSystem.addBot(moveEnemy.ent)
		slashSystem.slashees = append(slashSystem.slashees, moveEnemy.ent)
	}
	mapBounds := newRectangle(
		location{0, 0},
		dimens{2000, 2000},
	)
	renderingSystem.addShape(mapBounds.shape)
	collideSystem.addSolid(mapBounds.shape)

	diagonalWall := shape{
		[]line{
			line{
				location{250, 310},
				location{600, 655},
			},
		},
	}
	renderingSystem.addShape(&diagonalWall)
	collideSystem.addSolid(&diagonalWall)

	lilRoom := newRectangle(
		location{45, 400},
		dimens{70, 20},
	)
	renderingSystem.addShape(lilRoom.shape)
	collideSystem.addSolid(lilRoom.shape)

	anotherRoom := newRectangle(location{900, 1200}, dimens{90, 150})
	renderingSystem.addShape(anotherRoom.shape)
	collideSystem.addSolid(anotherRoom.shape)
}
