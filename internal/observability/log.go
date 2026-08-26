package observability

import (
	"encoding/json"
	"log/slog"
	"os"
)

func NewLogger(level string) *slog.Logger {
	var l slog.Level
	if level == "debug" {
		l = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}

var lastJSONPayload string

func JSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return lastJSONPayload
	}
	lastJSONPayload = string(b)
	return lastJSONPayload
}
