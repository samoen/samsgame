package main

const mapBoundWidth = 5000

func initEntities() {
	accelplayer := newControlledEntity()
	addPlayerControlled(accelplayer)
	collideSystem.addEnt(accelplayer)
	collideSystem.addSolid(accelplayer.rect.shape)
	renderingSystem.addShape(accelplayer.rect.shape)
	centerOn = accelplayer.rect
	playerSlasher := newSlasher(accelplayer)
	slashSystem.addSlasher(playerSlasher)
	// pivotingSystem.addSlashee(accelplayer.rect.shape)

	ps := &playerSprite{accelplayer.rect, playerStandImage}
	addPlayerSprite(ps)

	for i := 1; i < 30; i++ {
		moveEnemy := newControlledEntity()
		moveEnemy.rect.refreshShape(location{i*50 + 50, i * 30})
		enemySlasher := newSlasher(moveEnemy)
		slashSystem.addSlasher(enemySlasher)
		pivotingSystem.addSlashee(moveEnemy.rect.shape)
		renderingSystem.addShape(moveEnemy.rect.shape)
		collideSystem.addEnt(moveEnemy)
		collideSystem.addSolid(moveEnemy.rect.shape)
		botsMoveSystem.addEnemy(moveEnemy)
		es := &playerSprite{moveEnemy.rect, playerStandImage}
		addPlayerSprite(es)
	}

	mapBounds := newRectangle(
		location{0, 0},
		dimens{mapBoundWidth, mapBoundWidth},
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
