package server

import (
	"battlebit/internal/bb"
	"battlebit/internal/hub"
	"battlebit/internal/log"
	"battlebit/internal/player"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

type GameServer struct {
	hub      *hub.Hub
	upgrader websocket.Upgrader
}

func NewGameServer(h *hub.Hub) *GameServer {
	return &GameServer{
		hub: h,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}

func (gs *GameServer) HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home HTTP")
}
func (gs *GameServer) WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := log.GetLogger(ctx)
	ws, err := gs.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("Failed to upgrade to websocket", "error", err.Error())
		return
	}
	log.Info("Client connected", "remoteAddr", ws.RemoteAddr())

	mssgBytes := []byte("Hi Client!")
	err = ws.WriteMessage(1, mssgBytes)
	if err != nil {
		log.Error("Failed to write message to client", "error", err.Error())
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	gs.messageProcessor(ctx, ws)
	slog.Info("Client disconnected", slog.String("remoteAddr", ws.RemoteAddr().String()))
	err = ws.Close()
	if err != nil {
		log.Error("Failed to close connection", "error", err.Error())
	}
}

func (gs *GameServer) messageProcessor(ctx context.Context, conn *websocket.Conn) {
	log := log.GetLogger(ctx)
	counterReceived := 0
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Error("Failed to read message", "error", err.Error(), "counterReceived", counterReceived, "messageType", messageType)
			continue
		}
		// print out that message for clarity
		msgTxt := string(p)
		log.Info("Received message", "message", msgTxt)
		counterReceived++

		jsonRPCRequest := new(JSONRPCRequest)
		err = json.Unmarshal(p, jsonRPCRequest)
		if err != nil {
			log.Error("Failed to unmarshal JSONRPCRequest", "error", err.Error(), "counterReceived", counterReceived, "messageType", messageType)
			resp := responseError(jsonRPCRequest, 400, err)
			sendMsg(ctx, conn, resp)
			continue
		} else {
			resp := gs.processRequest(ctx, jsonRPCRequest, conn)
			sendMsg(ctx, conn, resp)
		}
	}
}
func (gs *GameServer) processRequest(ctx context.Context, req *JSONRPCRequest, conn *websocket.Conn) *JSONRPCResponse {
	switch req.Method {
	case METHOD_CREATE_GAME, METHOD_LIST_GAMES, METHOD_GET_GAME, METHOD_REMOVE_GAME:
		return gs.hubRouter(ctx, req)
	case METHOD_JOIN_GAME, METHOD_LEAVE_GAME, METHOD_PLAYER_MOVE:
		return gs.gameRouter(ctx, req, conn)
	case METHOD_GAME_METRICS:
		return gs.metricRouter(ctx, req)
	default:
		return responseResult(req, map[string]string{"message": "method not found"})
	}
}

func (gs *GameServer) hubRouter(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	log := log.GetLogger(ctx)
	switch req.Method {
	case METHOD_CREATE_GAME:
		log.Info("Creating new game", "params", string(req.Params))
		ng := new(hub.CrateNewGame)
		err := json.Unmarshal(req.Params, ng)
		if err != nil {
			log.Error("Failed to unmarshal CrateNewGame", "error", err.Error())
			return responseError(req, 400, err)
		}
		game := gs.hub.CreateNewGame(ctx, *ng)
		started := game.StartGame(ctx)
		return responseResult(req, started)
	case METHOD_LIST_GAMES:
		log.Info("Listing games")
		return responseResult(req, gs.hub.ListGames(ctx))
	case METHOD_GET_GAME:
		log.Info("Getting game", "params", string(req.Params))
		gId := new(hub.GameId)
		err := json.Unmarshal(req.Params, gId)
		if err != nil {
			log.Error("Failed to unmarshal GetGame", "error", err.Error())
			return responseError(req, 400, err)
		}
		g, err := gs.hub.GetGame(ctx, *gId)
		if err != nil {
			log.Error("Failed to get game", "error", err.Error())
			return responseError(req, 404, err)
		}
		return responseResult(req, g)
	case METHOD_REMOVE_GAME:
		log.Info("Removing game", "params", string(req.Params))
		gId := new(hub.GameId)
		err := json.Unmarshal(req.Params, gId)
		if err != nil {
			log.Error("Failed to unmarshal RemoveGame", "error", err.Error())
			return responseError(req, 400, err)
		}
		gs.hub.RemoveGame(ctx, *gId)
		return responseResult(req, map[string]string{"message": "game removed"})
	default:
		log.Info("Method not found", "method", req.Method)
		return responseResult(req, map[string]string{"message": "method not found"})
	}
}

func (gs *GameServer) gameRouter(ctx context.Context, req *JSONRPCRequest, conn *websocket.Conn) *JSONRPCResponse {
	log := log.GetLogger(ctx)
	switch req.Method {
	case METHOD_JOIN_GAME:
		log.Info("Joining game", "params", string(req.Params))
		pj := new(bb.PlayerJoin)
		err := json.Unmarshal(req.Params, pj)
		if err != nil {
			log.Error("Failed to unmarshal PlayerJoin", "error", err.Error())
			return responseError(req, 400, err)
		}
		game, err := gs.hub.GetGame(ctx, hub.GameId{ID: pj.GameId})
		if err != nil {
			log.Error("Failed to get game", "error", err.Error())
			return responseError(req, 404, err)
		}
		player := player.NewPlayer(pj.PlayerName, conn.RemoteAddr().String())
		pa, err := game.AddPlayer(ctx, player)
		if err != nil {
			log.Error("Failed to add player", "error", err.Error())
			return responseError(req, 400, err)
		}
		return responseResult(req, pa)
	case METHOD_LEAVE_GAME:
		log.Info("Leaving game", "params", string(req.Params))
		pj := new(bb.PlayerLeave)
		err := json.Unmarshal(req.Params, pj)
		if err != nil {
			log.Error("Failed to unmarshal PlayerJoin", "error", err.Error())
			return responseError(req, 400, err)
		}
		game, err := gs.hub.GetGame(ctx, hub.GameId{ID: pj.GameId})
		if err != nil {
			log.Error("Failed to get game", "error", err.Error())
			return responseError(req, 404, err)
		}
		return responseResult(req, game.RemovePlayer(ctx, pj.PlayerId))
	case METHOD_PLAYER_MOVE:
		log.Info("Player move", "params", string(req.Params))
		pm := new(bb.PlayerMove)
		err := json.Unmarshal(req.Params, pm)
		if err != nil {
			log.Error("Failed to unmarshal PlayerMove", "error", err.Error())
			return responseError(req, 400, err)
		}
		game, err := gs.hub.GetGame(ctx, hub.GameId{ID: pm.GameId})
		if err != nil {
			log.Error("Failed to get game", "error", err.Error())
			return responseError(req, 404, err)
		}
		return responseResult(req, game.PlayerMove(ctx, pm.PlayerId, pm.Index))
	default:
		log.Info("Method not found", "method", req.Method)
		return responseResult(req, map[string]string{"message": "method not found"})
	}
}

func (gs *GameServer) metricRouter(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	log := log.GetLogger(ctx)
	switch req.Method {
	case METHOD_GAME_METRICS:
		log.Info("Game metrics", "params", string(req.Params))
		ga := new(hub.GameId)
		err := json.Unmarshal(req.Params, ga)
		if err != nil {
			log.Error("Failed to unmarshal GameMetrics", "error", err.Error())
			return responseError(req, 400, err)
		}
		game, err := gs.hub.GetGame(ctx, hub.GameId{ID: ga.ID})
		if err != nil {
			log.Error("Failed to get game", "error", err.Error())
			return responseError(req, 404, err)
		}
		return responseResult(req, game.Metrics(ctx))
	default:
		log.Info("Method not found", "method", req.Method)
		return responseResult(req, map[string]string{"message": "method not found"})
	}
}
