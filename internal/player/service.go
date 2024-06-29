package player

import "github.com/google/uuid"

type Player struct {
	PlayerId         string
	PlayerName       string
	PlayerConnection string
}

func NewPlayer(playerName string, playerConnection string) *Player {
	return &Player{
		PlayerId:         uuid.New().String(),
		PlayerName:       playerName,
		PlayerConnection: playerConnection,
	}
}
