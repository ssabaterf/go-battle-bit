package server

import "encoding/json"

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"` // Params can be anything, so use json.RawMessage
	ID      interface{}     `json:"id"`               // ID can be string, number, or null
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"` // Result can be anything, so use interface{}
	Error   *JSONRPCError `json:"error,omitempty"`  // Error is a pointer to JSONRPCError
	ID      interface{}   `json:"id"`               // ID can be string, number, or null
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // Data can be anything, so use interface{}
}
