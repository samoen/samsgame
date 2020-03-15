package main

func initEntities() {

	accelplayer := newControlledEntity()
	playerMoveSystem.addPlayer(accelplayer)
	collideSystem.addEnt(accelplayer)
	renderingSystem.addShape(accelplayer.rect.shape)
	renderingSystem.CenterOn = accelplayer.rect
	weaponRenderingSystem.CenterOn = accelplayer.rect
	playerSlasher := newSlasher(accelplayer)
	slashSystem.addSlasher(playerSlasher)
	// pivotingSystem.addSlashee(accelplayer.rect.shape)

	for i := 1; i < 20; i++ {
		moveEnemy := newControlledEntity()
		moveEnemy.rect.refreshShape(location{i * 50, i * 30})
		enemySlasher := newSlasher(moveEnemy)
		slashSystem.addSlasher(enemySlasher)
		pivotingSystem.addSlashee(moveEnemy.rect.shape)
		renderingSystem.addShape(moveEnemy.rect.shape)
		collideSystem.addEnt(moveEnemy)
		botsMoveSystem.addEnemy(moveEnemy)
	}

	mapBounds := newRectangle(
		location{0, 0},
		dimens{5000, 5000},
	)
	renderingSystem.addShape(mapBounds.shape)
	collideSystem.addSolid(mapBounds.shape)
	pivotingSystem.addBlocker(mapBounds.shape)

	diagonalWall := newShape()
	diagonalWall.lines = []line{
		line{
			location{250, 310},
			location{600, 655},
		},
	}

	renderingSystem.addShape(diagonalWall)
	collideSystem.addSolid(diagonalWall)
	pivotingSystem.addBlocker(diagonalWall)

	lilRoom := newRectangle(
		location{45, 400},
		dimens{70, 20},
	)
	pivotingSystem.addBlocker(lilRoom.shape)
	renderingSystem.addShape(lilRoom.shape)
	collideSystem.addSolid(lilRoom.shape)

	anotherRoom := newRectangle(location{900, 1200}, dimens{90, 150})
	renderingSystem.addShape(anotherRoom.shape)
	collideSystem.addSolid(anotherRoom.shape)
}
