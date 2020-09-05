package gamecore

type EntityData struct {
	X          int
	Y          int
	Xaxis      int
	Yaxis      int
	Right      bool
	Down       bool
	Left       bool
	Up         bool
	Swinging   bool
	Startangle float64
	IHit       []string
	Dmg        int
	CurrentHP  int
	MaxHP      int
	MyPNum     string
}

type MessageToServer struct {
	MyData    EntityData
	MyAnimals []EntityData
}

type MessageToClient struct {
	Locs     []EntityData
	YourPNum string
}
