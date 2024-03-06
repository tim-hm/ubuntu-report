package log

import (
	"log/slog"
	"os"
)

var globalLevel = &slog.LevelVar{}

func init() {
	logFile, err := os.OpenFile("/var/log/charm_workload.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	h := slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: globalLevel})
	slog.SetDefault(slog.New(h))
	globalLevel.Set(slog.LevelWarn)

	//h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: globalLevel})
	//slog.SetDefault(slog.New(h))
	//globalLevel.Set(slog.LevelWarn)
}

// SetLevel change global handler log level.
func SetLevel(l slog.Level) {
	globalLevel.Set(l)
}
