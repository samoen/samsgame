package gamecore

type entityid struct {
	nothing bool
}

type pivotingShape struct {
	pivoterShape   shape
	animationCount float64
	alreadyHit     map[*shape]bool
	startCount     float64
	bladeLength    int
	damage         int
}


// func rotateAround(center location, point location, angle float64) location {
// 	result := location{}
// 	rotatedX := math.Cos(angle)*float64(point.x-center.x) - math.Sin(angle)*float64(point.y-center.y) + float64(center.x)
// 	rotatedY := math.Sin(angle)*float64(point.x-center.x) + math.Cos(angle)*float64(point.y-center.y) + float64(center.y)
// 	result.x = int(rotatedX)
// 	result.y = int(rotatedY)
// 	return result
// }
