package main

import (
	"battlebit/internal/hub"
	"battlebit/internal/log"
	"battlebit/internal/middleware"
	"battlebit/internal/server"
	"fmt"
	"log/slog"
	"net/http"
)

func main() {
	log.SetLogs(slog.LevelInfo)
	slog.Info("Starting services...")
	hub := hub.NewHub()
	slog.Info("Hub created")
	gs := server.NewGameServer(hub)
	slog.Info("GameServer created")

	mux := http.NewServeMux()
	setupRoutes(mux, gs)

	muxLR := middleware.LogMiddleware(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", "8080"),
		Handler: muxLR,
	}

	slog.Info("Server started", slog.String("port", "8080"))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Error starting server", slog.String("error", err.Error()))
	}
}

func setupRoutes(mux *http.ServeMux, gs *server.GameServer) {
	mux.HandleFunc("/", gs.HomePage)
	mux.HandleFunc("/ws", gs.WsEndpoint)
}
