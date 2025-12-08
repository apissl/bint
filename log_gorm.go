package bint

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type gormLogger struct {
	log *slog.Logger
}

func NewGormLogger(log *slog.Logger) logger.Interface {
	return &gormLogger{
		log: log.With("type", "gorm"), // 添加类型
	}
}

func (g *gormLogger) LogMode(_ logger.LogLevel) logger.Interface {
	return g
}

func (g *gormLogger) Info(ctx context.Context, s string, i ...any) {
	g.log.InfoContext(ctx, fmt.Sprintf(s, i...))
}

func (g *gormLogger) Warn(ctx context.Context, s string, i ...any) {
	g.log.WarnContext(ctx, fmt.Sprintf(s, i...))
}

func (g *gormLogger) Error(ctx context.Context, s string, i ...any) {
	g.log.ErrorContext(ctx, fmt.Sprintf(s, i...))
}

func (g *gormLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 执行耗时
	latency := time.Since(begin).Milliseconds()
	sql, rows := fc()
	if err != nil {
		g.log.Error("sql_trace", "latency", latency, "sql", sql, "rows", rows, "code_line", utils.FileWithLineNum(), "err", err)
	} else {
		g.log.Debug("sql_trace", "latency", latency, "sql", sql, "rows", rows, "code_line", utils.FileWithLineNum())
	}
}
