package bint

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Context 自定义上下文
type Context struct {
	Ctx   fiber.Ctx
	DB    *gorm.DB
	Cache *redis.Client
	Log   *slog.Logger
}

// Handler 目前仅用于Setup构建中间件
type Handler = func(Context) error
