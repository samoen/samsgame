package main

const worldWidth = 5000

func initEntities() {
	playerid := &entityid{}
	accelplayer := newControlledEntity()
	addPlayerControlled(accelplayer, playerid)
	collideSystem.addEnt(accelplayer, playerid)
	collideSystem.addSolid(accelplayer.rect.shape, playerid)
	addHitbox(accelplayer.rect.shape, playerid)
	centerOn = accelplayer.rect
	playerSlasher := newSlasher(accelplayer)
	addSlasher(playerid, playerSlasher)

	ps := &baseSprite{}
	ps.playerRect = accelplayer.rect
	ps.sprite = playerStandImage
	ps.redScale = new(int)
	ps.flip = &accelplayer.directions
	addBasicSprite(ps, playerid)

	for i := 1; i < 30; i++ {
		enemyid := &entityid{}
		moveEnemy := newControlledEntity()
		moveEnemy.rect.refreshShape(location{i*50 + 50, i * 30})
		enemySlasher := newSlasher(moveEnemy)
		addSlasher(enemyid, enemySlasher)
		addHitbox(moveEnemy.rect.shape, enemyid)
		collideSystem.addEnt(moveEnemy, enemyid)
		collideSystem.addSolid(moveEnemy.rect.shape, enemyid)
		eController := &enemyController{}
		eController.aEnt = moveEnemy
		addEnemyController(eController, enemyid)

		botDeathable := deathable{}
		botDeathable.currentHP = 3
		botDeathable.maxHP = 3
		botDeathable.deathableShape = moveEnemy.rect
		addDeathable(enemyid, &botDeathable)
		es := &baseSprite{}
		es.playerRect = moveEnemy.rect
		es.sprite = playerStandImage
		es.redScale = &botDeathable.redScale
		es.flip = &moveEnemy.directions
		addBasicSprite(es, enemyid)
	}

	worldBoundaryID := &entityid{}
	worldBoundRect := newRectangle(
		location{0, 0},
		dimens{worldWidth, worldWidth},
	)
	addHitbox(worldBoundRect.shape, worldBoundaryID)
	collideSystem.addSolid(worldBoundRect.shape, worldBoundaryID)
	pivotingSystem.addBlocker(worldBoundRect.shape, worldBoundaryID)

	diagonalWallID := &entityid{}
	diagonalWall := newShape()
	diagonalWall.lines = []line{
		{
			location{250, 310},
			location{600, 655},
		},
	}

	addHitbox(diagonalWall, diagonalWallID)
	collideSystem.addSolid(diagonalWall, diagonalWallID)
	pivotingSystem.addBlocker(diagonalWall, diagonalWallID)

	lilRoomID := &entityid{}
	lilRoom := newRectangle(
		location{45, 400},
		dimens{70, 20},
	)
	pivotingSystem.addBlocker(lilRoom.shape, lilRoomID)
	addHitbox(lilRoom.shape, lilRoomID)
	collideSystem.addSolid(lilRoom.shape, lilRoomID)

	anotherRoomID := &entityid{}
	anotherRoom := newRectangle(location{900, 1200}, dimens{90, 150})
	addHitbox(anotherRoom.shape, anotherRoomID)
	collideSystem.addSolid(anotherRoom.shape, anotherRoomID)
}
