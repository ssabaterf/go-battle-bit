package hub

import (
	"battlebit/internal/bb"
	"battlebit/internal/log"
	"battlebit/internal/status"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Hub struct {
	LimitGames int
	Games      map[string]*bb.Game
}

func NewHub() *Hub {
	limitGame := getLimitGame()
	slog.Debug("Created Hub", "limitGames", limitGame)
	return &Hub{
		LimitGames: limitGame,
		Games:      make(map[string]*bb.Game),
	}
}

func (h *Hub) CreateNewGame(ctx context.Context, ng CrateNewGame) *bb.Game {
	slog := log.GetLogger(ctx)

	status := status.NewGameStatus(ng.Size)
	game := bb.NewGame(status, ng.Autopilots)
	h.Games[game.GameId] = game
	slog.Debug("Game created", "gameId", game.GameId, "size", ng.Size)
	return game
}

func (h *Hub) ListGames(ctx context.Context) []*bb.GameMetrics {
	slog := log.GetLogger(ctx)

	games := make([]*bb.GameMetrics, 0)
	for _, g := range h.Games {
		games = append(games, g.Metrics(ctx))
	}
	slog.Debug("List games", "games", len(games))
	return games
}

func (h *Hub) GetGame(ctx context.Context, gameId GameId) (*bb.Game, error) {
	slog := log.GetLogger(ctx)
	g, ok := h.Games[gameId.ID]
	if !ok {
		slog.Debug("Game not found", "gameId", gameId.ID)
		return nil, fmt.Errorf("game not found")
	}
	slog.Debug("Game found", "gameId", gameId.ID)
	return g, nil
}

func (h *Hub) RemoveGame(ctx context.Context, gameId GameId) {
	slog := log.GetLogger(ctx)
	delete(h.Games, gameId.ID)
	slog.Debug("Game removed", "gameId", gameId.ID)
}

func getLimitGame() int {
	limitGameString, ok := os.LookupEnv("BB_LIMIT_GAMES")
	if ok {
		limitGameInt, err := strconv.Atoi(limitGameString)
		if err != nil {
			slog.Error("Error parsing BB_LIMIT_GAMES", "error", err.Error())
			limitGameInt = 5
			slog.Debug("Using default value", "BB_LIMIT_GAMES", limitGameInt)
		}
		return limitGameInt
	}
	return 5
}
