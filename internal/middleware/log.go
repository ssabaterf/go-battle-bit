package middleware

import (
	"battlebit/internal/contextkey"
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var RequestidHeader = "X-Request-Id"

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		ipAddress := getIpAddress(r)
		ctx := r.Context()
		requestID := r.Header.Get(RequestidHeader)
		if requestID == "" {
			requestID = GenerateRequestID()
		}

		r.Header.Set(RequestidHeader, requestID)
		w.Header().Set(RequestidHeader, requestID)

		reqLogger := slog.With(slog.String("requestId", requestID))
		ctx = context.WithValue(ctx, contextkey.ReqIdCtx, requestID)
		ctx = context.WithValue(ctx, contextkey.SlogCtx, reqLogger)
		r = r.WithContext(ctx)

		wrapper := &statusTrackingResponseWriter{w, http.StatusOK}

		defer func() {
			slog.Info("", slog.String("requestId", requestID), slog.String("ipaddress", ipAddress), slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Int("status", wrapper.statusCode), slog.Duration("duration", time.Duration(time.Since(start).Microseconds())))
		}()

		next.ServeHTTP(wrapper, r)

	})
}

func getIpAddress(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

type statusTrackingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusTrackingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusTrackingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func GenerateRequestID() string {
	return uuid.New().String()
}
