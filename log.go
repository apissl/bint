package bint

import (
	"io"
	"log/slog"
	"path/filepath"
	"strings"
)

// LogLevel 日志等级
type LogLevel string

func (l LogLevel) GetLevel() slog.Level {
	switch strings.ToLower(string(l)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "fatal":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewJsonLogger Json格式日志
func NewJsonLogger(w io.Writer, level LogLevel) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       level.GetLevel(),
		ReplaceAttr: attrReplacer(),
		AddSource:   false,
	})
	logger := slog.New(h)
	slog.SetDefault(logger)
	return logger
}

// attrReplacer 属性替换函数
func attrReplacer() func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.TimeKey:
			//a.Value = slog.StringValue(a.Value.Time().Format(time.DateTime))
		case slog.SourceKey:
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
}
