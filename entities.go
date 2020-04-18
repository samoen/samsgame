package main

const worldWidth = 5000

func initEntities() {
	playerid := &entityid{}
	accelplayer := newControlledEntity()
	addPlayerControlled(accelplayer, playerid)
	addEnt(accelplayer, playerid)
	addSolid(accelplayer.rect.shape, playerid)
	addHitbox(accelplayer.rect.shape, playerid)
	centerOn = accelplayer.rect
	playerSlasher := newSlasher(accelplayer)
	addSlasher(playerid, playerSlasher)
	pDeathable := deathable{}
	pDeathable.currentHP = 6
	pDeathable.maxHP = 6
	pDeathable.deathableShape = accelplayer.rect
	addDeathable(playerid, &pDeathable)

	ps := &baseSprite{}
	ps.redScale = new(int)
	ps.sprite = playerStandImage
	ps.owner = accelplayer
	addBasicSprite(ps, playerid)

	for i := 1; i < 30; i++ {
		enemyid := &entityid{}
		moveEnemy := newControlledEntity()
		moveEnemy.rect.refreshShape(location{i*50 + 50, i * 30})
		enemySlasher := newSlasher(moveEnemy)
		addSlasher(enemyid, enemySlasher)
		addHitbox(moveEnemy.rect.shape, enemyid)
		addEnt(moveEnemy, enemyid)
		addSolid(moveEnemy.rect.shape, enemyid)
		eController := &enemyController{}
		eController.aEnt = moveEnemy
		addEnemyController(eController, enemyid)

		botDeathable := deathable{}
		botDeathable.currentHP = 3
		botDeathable.maxHP = 3
		botDeathable.deathableShape = moveEnemy.rect
		addDeathable(enemyid, &botDeathable)
		es := &baseSprite{}
		es.redScale = &botDeathable.redScale
		es.sprite = playerStandImage
		es.owner = moveEnemy
		addBasicSprite(es, enemyid)
	}

	worldBoundaryID := &entityid{}
	worldBoundRect := newRectangle(
		location{0, 0},
		dimens{worldWidth, worldWidth},
	)
	addHitbox(worldBoundRect.shape, worldBoundaryID)
	addSolid(worldBoundRect.shape, worldBoundaryID)
	addBlocker(worldBoundRect.shape, worldBoundaryID)

	diagonalWallID := &entityid{}
	diagonalWall := newShape()
	diagonalWall.lines = []line{
		{
			location{250, 310},
			location{600, 655},
		},
	}

	addHitbox(diagonalWall, diagonalWallID)
	addSolid(diagonalWall, diagonalWallID)
	addBlocker(diagonalWall, diagonalWallID)

	lilRoomID := &entityid{}
	lilRoom := newRectangle(
		location{45, 400},
		dimens{70, 20},
	)
	addBlocker(lilRoom.shape, lilRoomID)
	addHitbox(lilRoom.shape, lilRoomID)
	addSolid(lilRoom.shape, lilRoomID)

	anotherRoomID := &entityid{}
	anotherRoom := newRectangle(location{900, 1200}, dimens{90, 150})
	addHitbox(anotherRoom.shape, anotherRoomID)
	addSolid(anotherRoom.shape, anotherRoomID)
}
