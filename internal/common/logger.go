package common

import (
	"fmt"
	"io"
	"log/slog"
)

func NewConsoleLogger(level slog.Level) *slog.Logger {
	return slog.New(NewLogHandler(
		func(message string) { fmt.Println(message) },
		&slog.HandlerOptions{Level: level}))
}

func NewStructuredLogger(writer io.Writer, level slog.Level) *slog.Logger {
	return slog.New(NewLogHandler(
		func(message string) { writer.Write([]byte(message + "\n")) },
		&slog.HandlerOptions{Level: level}))
}
