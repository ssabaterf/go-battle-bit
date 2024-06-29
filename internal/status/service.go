package status

import (
	"battlebit/internal/log"
	"context"
	"sync"
)

type GameStatus struct {
	Size        int
	Status      []byte
	HasStarted  bool
	HasFinished bool
	mutex       sync.Mutex
}

func NewGameStatus(sizeGame int) *GameStatus {
	return &GameStatus{
		Status:      make([]byte, (sizeGame+7)/8),
		Size:        sizeGame,
		HasStarted:  false,
		HasFinished: false,
	}
}
func (g *GameStatus) isBitOn(pos int) bool {
	return g.Status[pos>>3]&(1<<(pos&7)) != 0
}

func (g *GameStatus) isBitOff(pos int) bool {
	return g.Status[pos>>3]&(1<<(pos&7)) == 0
}

func (g *GameStatus) ToggleBit(ctx context.Context, pos int) {
	slog := log.GetLogger(ctx)

	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.HasStarted = true
	if g.isBitOn(pos) {
		return
	}
	g.Status[pos>>3] ^= (1 << (pos & 7))
	zeroes, ones := g.countBits()
	slog.Debug("Bit toggled", "pos", pos, "zeroes", zeroes, "ones", ones)
	if zeroes == 0 && ones == g.Size {
		g.HasFinished = true
		slog.Debug("Game finished")
	}
}
func (g *GameStatus) countBits() (int, int) {
	count0, count1 := 0, 0
	for i := 0; i < g.Size; i++ {
		if g.isBitOn(i) {
			count1++
		} else {
			count0++
		}
	}
	return count0, count1
}
