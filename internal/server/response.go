package server

import (
	"battlebit/internal/log"
	"context"
	"encoding/json"

	"github.com/gorilla/websocket"
)

func responseError(req *JSONRPCRequest, code int, err error) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: err.Error(),
		},
		ID: req.ID,
	}
}

func responseResult(req *JSONRPCRequest, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}
}
func sendMsg(ctx context.Context, conn *websocket.Conn, resp *JSONRPCResponse) {
	log := log.GetLogger(ctx)

	p, err := json.Marshal(resp)
	if err != nil {
		log.Error("Failed to marshal JSONRPCResponse", "error", err.Error())
		subResp := &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    500,
				Message: err.Error(),
			},
			ID: resp.ID,
		}
		p, _ = json.Marshal(subResp)
	}
	if err := conn.WriteMessage(1, p); err != nil {
		log.Error("Failed to write message", "error", err.Error())
		return
	}
}
