package bb

import "time"

type GameStarted struct {
	GameId           string    `json:"gameId"`
	SizeGame         int       `json:"sizeGame"`
	InitTime         time.Time `json:"initTime"`
	NumberAutoPilots int       `json:"numberAutoPilots"`
}

type GameFinished struct {
	GameId           string        `json:"gameId"`
	SizeGame         int           `json:"sizeGame"`
	InitTime         time.Time     `json:"initTime"`
	NumberAutoPilots int           `json:"numberAutoPilots"`
	WinnerId         string        `json:"winnerId"`
	WinnerName       string        `json:"winnerName"`
	Duration         time.Duration `json:"duration"`
}

type PlayerAdded struct {
	GameId     string `json:"gameId"`
	PlayerId   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

type PlayerRemoved struct {
	GameId   string `json:"gameId"`
	PlayerId string `json:"playerId"`
}

type PlayerMoved struct {
	GameId     string     `json:"gameId"`
	PlayerId   string     `json:"playerId"`
	Index      int        `json:"index"`
	TimeMove   time.Time  `json:"timeMove"`
	GameStatus GameStatus `json:"gameStatus"`
}

type GameStatus struct {
	IsInProcess bool `json:"isInProcess"`
	IsFinished  bool `json:"isFinished"`
}

type GameMetrics struct {
	GameId                string     `json:"gameId"`
	SizeGame              int        `json:"sizeGame"`
	NumberAutoPilots      int        `json:"numberAutoPilots"`
	Players               int        `json:"players"`
	AutoPilots            int        `json:"autoPilots"`
	AutoPilotTotalIters   uint64     `json:"autoPilotTotalIters"`
	AutoPilotCurrentIters int        `json:"autoPilotCurrentIters"`
	AutoPilotMoves        uint64     `json:"autoPilotMoves"`
	CurrentDuration       string     `json:"currentDuration"`
	GameStatus            GameStatus `json:"gameStatus"`
}

type PlayerJoin struct {
	GameId     string `json:"gameId"`
	PlayerName string `json:"playerName"`
}

type PlayerLeave struct {
	GameId   string `json:"gameId"`
	PlayerId string `json:"playerName"`
}
type PlayerMove struct {
	GameId   string `json:"gameId"`
	PlayerId string `json:"playerName"`
	Index    int    `json:"index"`
}
