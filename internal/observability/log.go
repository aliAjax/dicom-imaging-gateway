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
func JSON(v any) string { b, _ := json.Marshal(v); return string(b) }
