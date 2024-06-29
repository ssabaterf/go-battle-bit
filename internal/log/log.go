package log

import (
	"battlebit/internal/contextkey"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
)

func SetLogs(level slog.Level) {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				s := a.Value.Any().(*slog.Source)
				s.File = path.Base(s.File)
				return slog.String("file", fmt.Sprintf("%s:%d", s.File, s.Line))
			}
			return a
		}})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(contextkey.SlogCtx).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return logger
}
