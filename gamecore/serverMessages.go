package gamecore

type EntityData struct {
	X             int
	Y             int
	Xaxis         int
	Yaxis         int
	Right         bool
	Down          bool
	Left          bool
	Up            bool
	NewSwing      bool
	NewSwingAngle float64
	Heading       float64
	Swangin       bool
	IHit          []string
	Dmg           int
	CurrentHP     int
	MaxHP         int
	MyPNum        string
}

type MessageToServer struct {
	MyData    EntityData
	MyAnimals []EntityData
}

type MessageToClient struct {
	Locs     []EntityData
	YourPNum string
}
