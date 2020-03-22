package main

const mapBoundWidth = 5000

func initEntities() {
	playerid := &entityid{}
	accelplayer := newControlledEntity()
	addPlayerControlled(accelplayer, playerid)
	collideSystem.addEnt(accelplayer, playerid)
	collideSystem.addSolid(accelplayer.rect.shape, playerid)
	renderingSystem.addShape(accelplayer.rect.shape, playerid)
	centerOn = accelplayer.rect
	playerSlasher := newSlasher(accelplayer)
	slashSystem.addSlasher(playerid, playerSlasher)
	// pivotingSystem.addSlashee(accelplayer.rect.shape)

	ps := &playerSprite{accelplayer.rect, playerStandImage}
	addPlayerSprite(ps, playerid)

	for i := 1; i < 30; i++ {
		enemyid := &entityid{}
		moveEnemy := newControlledEntity()
		moveEnemy.rect.refreshShape(location{i*50 + 50, i * 30})
		enemySlasher := newSlasher(moveEnemy)
		slashSystem.addSlasher(enemyid, enemySlasher)
		pivotingSystem.addSlashee(moveEnemy.rect.shape, enemyid)
		renderingSystem.addShape(moveEnemy.rect.shape, enemyid)
		collideSystem.addEnt(moveEnemy, enemyid)
		collideSystem.addSolid(moveEnemy.rect.shape, enemyid)
		botsMoveSystem.addEnemy(moveEnemy, enemyid)
		es := &playerSprite{moveEnemy.rect, playerStandImage}
		addPlayerSprite(es, enemyid)
	}

	mapBoundsID := &entityid{}
	mapBounds := newRectangle(
		location{0, 0},
		dimens{mapBoundWidth, mapBoundWidth},
	)
	renderingSystem.addShape(mapBounds.shape, mapBoundsID)
	collideSystem.addSolid(mapBounds.shape, mapBoundsID)
	pivotingSystem.addBlocker(mapBounds.shape, mapBoundsID)

	diagonalWallID := &entityid{}
	diagonalWall := newShape()
	diagonalWall.lines = []line{
		line{
			location{250, 310},
			location{600, 655},
		},
	}

	renderingSystem.addShape(diagonalWall, diagonalWallID)
	collideSystem.addSolid(diagonalWall, diagonalWallID)
	pivotingSystem.addBlocker(diagonalWall, diagonalWallID)

	lilRoomID := &entityid{}
	lilRoom := newRectangle(
		location{45, 400},
		dimens{70, 20},
	)
	pivotingSystem.addBlocker(lilRoom.shape, lilRoomID)
	renderingSystem.addShape(lilRoom.shape, lilRoomID)
	collideSystem.addSolid(lilRoom.shape, lilRoomID)

	anotherRoomID := &entityid{}
	anotherRoom := newRectangle(location{900, 1200}, dimens{90, 150})
	renderingSystem.addShape(anotherRoom.shape, anotherRoomID)
	collideSystem.addSolid(anotherRoom.shape, anotherRoomID)
}
