package hub

type CrateNewGame struct {
	Size       int `json:"size"`
	Autopilots int `json:"autopilots"`
}

type GameId struct {
	ID string `json:"id"`
}
