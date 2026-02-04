// Package logger provides structured logging utilities for the Statigo framework.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/term"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const requestIDKey contextKey = "request_id"

// InitLogger initializes and returns a structured logger.
func InitLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	format := os.Getenv("LOG_FORMAT")

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if strings.ToUpper(format) == "JSON" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = NewBracketHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// GenerateRequestID generates a new UUID for request tracking.
func GenerateRequestID() string {
	return uuid.New().String()
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorOrange = "\033[38;5;214m"
)

// BracketHandler is a custom slog.Handler that formats logs in bracket notation.
type BracketHandler struct {
	writer       io.Writer
	opts         *slog.HandlerOptions
	useColors    bool
	levelColors  map[slog.Level]string
	bracketColor string
	keyColor     string
	timeColor    string
	messageColor string
}

// NewBracketHandler creates a new BracketHandler.
func NewBracketHandler(w io.Writer, opts *slog.HandlerOptions) *BracketHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	useColors := false
	if f, ok := w.(*os.File); ok {
		useColors = term.IsTerminal(int(f.Fd()))
	}

	if os.Getenv("NO_COLOR") != "" {
		useColors = false
	}

	return &BracketHandler{
		writer:       w,
		opts:         opts,
		useColors:    useColors,
		bracketColor: colorGray,
		keyColor:     colorCyan,
		timeColor:    colorGray,
		messageColor: colorOrange,
		levelColors: map[slog.Level]string{
			slog.LevelDebug: colorBlue,
			slog.LevelInfo:  colorGreen,
			slog.LevelWarn:  colorYellow,
			slog.LevelError: colorRed,
		},
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *BracketHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle formats and writes a log record in bracket notation.
func (h *BracketHandler) Handle(_ context.Context, r slog.Record) error {
	buf := make([]byte, 0, 1024)

	levelColor := ""
	if h.useColors {
		if color, ok := h.levelColors[r.Level]; ok {
			levelColor = color
		}
	}

	// Add timestamp
	if h.useColors {
		buf = append(buf, h.timeColor...)
	}
	buf = append(buf, '[')
	buf = r.Time.AppendFormat(buf, "06/01/02-15:04:05")
	buf = append(buf, ']')
	if h.useColors {
		buf = append(buf, colorReset...)
	}

	// Add level with color
	if h.useColors && levelColor != "" {
		buf = append(buf, levelColor...)
	}
	buf = append(buf, '[')
	buf = append(buf, r.Level.String()...)
	buf = append(buf, ']')
	if h.useColors {
		buf = append(buf, colorReset...)
	}

	// Add message
	buf = append(buf, '[')
	if h.useColors {
		buf = append(buf, h.messageColor...)
	}
	buf = append(buf, r.Message...)
	if h.useColors {
		buf = append(buf, colorReset...)
	}
	buf = append(buf, ']')

	// Add attributes
	r.Attrs(func(a slog.Attr) bool {
		if h.useColors {
			buf = append(buf, h.bracketColor...)
		}
		buf = append(buf, '[')
		if h.useColors {
			buf = append(buf, colorReset...)
			buf = append(buf, h.keyColor...)
		}
		buf = append(buf, a.Key...)
		if h.useColors {
			buf = append(buf, colorReset...)
			buf = append(buf, h.bracketColor...)
		}
		buf = append(buf, '=')
		if h.useColors {
			buf = append(buf, colorReset...)
		}
		buf = append(buf, fmt.Sprint(a.Value.Any())...)
		if h.useColors {
			buf = append(buf, h.bracketColor...)
		}
		buf = append(buf, ']')
		if h.useColors {
			buf = append(buf, colorReset...)
		}
		return true
	})

	buf = append(buf, '\n')
	_, err := h.writer.Write(buf)
	return err
}

// WithAttrs returns a new handler with additional attributes.
func (h *BracketHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns a new handler with a group name.
func (h *BracketHandler) WithGroup(name string) slog.Handler {
	return h
}
