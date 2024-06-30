package bb

import (
	"battlebit/internal/log"
	"battlebit/internal/player"
	"battlebit/internal/status"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	GameId           string
	SizeGame         int
	Game             *status.GameStatus
	Players          []*player.Player
	LastMoveTime     time.Time
	LastMoveBy       string
	playerMutex      sync.Mutex
	InitTimer        time.Time
	NumberAutoPilots int
	DelayAutoPilots  int
	autoPilotBreak   chan struct{}
	totalIterations  uint64
	iterarations     int
	WinnerId         string
	WinnerName       string
}

func NewGame(status *status.GameStatus, NumberPilots int) *Game {
	return &Game{
		GameId:           uuid.New().String(),
		Game:             status,
		Players:          make([]*player.Player, 0),
		NumberAutoPilots: NumberPilots,
		DelayAutoPilots:  0,
		autoPilotBreak:   make(chan struct{}, 1),
	}
}

func (g *Game) AddPlayer(ctx context.Context, player *player.Player) (*PlayerAdded, error) {
	slog := log.GetLogger(ctx)

	g.playerMutex.Lock()
	defer g.playerMutex.Unlock()
	if len(g.Players) >= 10 {
		slog.Debug("Game is full", "gameId", g.GameId, "size", g.Game.Size)
		return nil, fmt.Errorf("game is full")
	}
	_, err := g.GetPlayerById(ctx, player.PlayerId)
	if err == nil {
		slog.Debug("Player already added", "gameId", g.GameId, "playerId", player.PlayerId)
		return nil, fmt.Errorf("player already added")
	}
	g.Players = append(g.Players, player)
	slog.Debug("Player added", "gameId", g.GameId, "playerId", player.PlayerId, "playerName", player.PlayerName)
	return &PlayerAdded{
		GameId:     g.GameId,
		PlayerId:   player.PlayerId,
		PlayerName: player.PlayerName,
	}, nil
}
func (g *Game) RemovePlayer(ctx context.Context, playerId string) *PlayerRemoved {
	slog := log.GetLogger(ctx)

	g.playerMutex.Lock()
	defer g.playerMutex.Unlock()
	for i, player := range g.Players {
		if player.PlayerId == playerId {
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			slog.Debug("Player removed", "gameId", g.GameId, "playerId", player.PlayerId)
			return &PlayerRemoved{
				GameId:   g.GameId,
				PlayerId: player.PlayerId,
			}
		}
	}
	slog.Debug("Player not found", "gameId", g.GameId, "playerId", playerId)
	return &PlayerRemoved{}
}
func (g *Game) PlayerMove(ctx context.Context, playerId string, index int) *PlayerMoved {
	slog := log.GetLogger(ctx)

	player, err := g.GetPlayerById(ctx, playerId)
	if err != nil {
		return &PlayerMoved{}
	}
	if index >= g.Game.Size {
		slog.Debug("Index out of range", "gameId", g.GameId, "playerId", player.PlayerId, "index", index)
		return &PlayerMoved{}
	}
	g.Game.ToggleBit(ctx, index)
	g.playerMutex.Lock()
	defer g.playerMutex.Unlock()
	g.LastMoveTime = time.Now()
	g.LastMoveBy = player.PlayerId
	if g.Game.HasFinished {
		g.WinnerId = player.PlayerId
		g.WinnerName = player.PlayerName
		g.FinishGame(ctx)
	}
	slog.Debug("Player moved", "gameId", g.GameId, "playerId", player.PlayerId, "index", index, "timeMove", g.LastMoveTime)
	return &PlayerMoved{
		GameId:   g.GameId,
		PlayerId: player.PlayerId,
		Index:    index,
		TimeMove: g.LastMoveTime,
		GameStatus: GameStatus{
			IsInProcess: g.Game.HasStarted && !g.Game.HasFinished,
			IsFinished:  g.Game.HasFinished,
		},
	}
}

func (g *Game) GetPlayerById(ctx context.Context, playerId string) (*player.Player, error) {
	slog := log.GetLogger(ctx)

	for _, player := range g.Players {
		if player.PlayerId == playerId {
			slog.Debug("Player found", "gameId", g.GameId, "playerId", player.PlayerId)
			return player, nil
		}
	}
	slog.Debug("Player not found", "gameId", g.GameId, "playerId", playerId)
	return nil, fmt.Errorf("player not found")
}

func (g *Game) GenerateGaussianRandomInt(mean, stddev, max int) int {
	for {
		randFloat := rand.NormFloat64()*float64(stddev) + float64(mean)
		randInt := int(randFloat)
		if randInt >= 0 && randInt < max {
			return randInt
		}
	}
}

func (g *Game) StartGame(ctx context.Context) *GameStarted {
	slog := log.GetLogger(ctx)

	g.InitTimer = time.Now()
	g.Game.HasStarted = true
	g.StartAutoPilots(ctx)

	slog.Debug("Game started", "gameId", g.GameId, "size", g.Game.Size, "players", len(g.Players))
	return &GameStarted{
		GameId:           g.GameId,
		SizeGame:         g.Game.Size,
		InitTime:         g.InitTimer,
		NumberAutoPilots: g.NumberAutoPilots,
	}
}

func (g *Game) FinishGame(ctx context.Context) *GameFinished {
	slog := log.GetLogger(ctx)
	g.autoPilotBreak <- struct{}{}
	g.Game.HasFinished = true

	slog.Info("Game finished", "gameId", g.GameId, "size", g.Game.Size, "players", len(g.Players), "duration", time.Since(g.InitTimer).String())
	slog.Info("Winner", "playerId", g.WinnerId, "playerName", g.WinnerName)
	return &GameFinished{
		GameId:           g.GameId,
		SizeGame:         g.Game.Size,
		InitTime:         g.InitTimer,
		NumberAutoPilots: g.NumberAutoPilots,
		WinnerId:         g.WinnerId,
		WinnerName:       g.WinnerName,
		Duration:         time.Since(g.InitTimer)}
}

func (g *Game) StartAutoPilots(ctx context.Context) {
	slog := log.GetLogger(ctx)

	autoPilots := make([]*player.Player, 0)
	for i := 0; i < g.NumberAutoPilots; i++ {
		autoPilot := player.NewPlayer(fmt.Sprintf("Autopilot %d", i), fmt.Sprintf("Connection %d", i))
		_, err := g.AddPlayer(ctx, autoPilot)
		if err != nil {
			slog.Error("Error adding autopilot", "error", err.Error())
			continue
		}
		autoPilots = append(autoPilots, autoPilot)
	}

	timeDelay := time.Duration(g.DelayAutoPilots) * time.Millisecond
	if g.DelayAutoPilots == 0 {
		timeDelay = 1 * time.Nanosecond
	}
	delay := time.NewTicker(timeDelay)
	go g.AutopilotGame(autoPilots, delay, g.autoPilotBreak)
	slog.Debug("AutoPilots started", "number", g.NumberAutoPilots, "delay", g.DelayAutoPilots)
}

func (g *Game) AutopilotGame(autoPilots []*player.Player, delay *time.Ticker, finisher chan struct{}) {
	ctx := context.Background()
	for !g.Game.HasFinished {
		randIndex1 := g.GenerateGaussianRandomInt(g.iterarations, 5, g.Game.Size)
		for _, autoPilot := range autoPilots {
			g.PlayerMove(ctx, autoPilot.PlayerId, randIndex1)
		}
		g.iterarations++
		if g.iterarations > g.Game.Size {
			g.totalIterations += uint64(g.iterarations)
			g.iterarations = 0
			slog.Debug("RESTARTING ITERATOR. Game is still running", "gameId", g.GameId, "size", g.Game.Size, "players", len(g.Players), "duration", time.Since(g.InitTimer).String())
		}
		<-delay.C

		select {
		case <-finisher:
			slog.Debug("AutoPilots Breaking", "iterations", g.totalIterations)
			return
		default:
			continue
		}
	}
}

func (g *Game) Metrics(ctx context.Context) *GameMetrics {
	log := log.GetLogger(ctx)
	log.Debug("Getting metrics", "gameId", g.GameId)
	return &GameMetrics{
		GameId:                g.GameId,
		SizeGame:              g.Game.Size,
		NumberAutoPilots:      g.NumberAutoPilots,
		Players:               len(g.Players),
		AutoPilots:            g.NumberAutoPilots,
		AutoPilotTotalIters:   g.totalIterations,
		AutoPilotCurrentIters: g.iterarations,
		AutoPilotMoves:        (g.totalIterations + uint64(g.iterarations)) * uint64(g.NumberAutoPilots),
		CurrentDuration:       time.Since(g.InitTimer).String(),
		GameStatus: GameStatus{
			IsInProcess: g.Game.HasStarted && !g.Game.HasFinished,
			IsFinished:  g.Game.HasFinished,
		},
	}
}
