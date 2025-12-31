package bint

import (
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Bint struct {
	app   *fiber.App
	db    *gorm.DB
	cache *redis.Client
	log   *slog.Logger
	cfg   *Config
}

func NewBint(cfg *Config) *Bint {
	logger := NewJsonLogger(os.Stdout, "debug")
	instance := &Bint{
		app: fiber.New(),
		log: logger,
		cfg: cfg,
	}
	instance.app.Use(
		cors.New(),
		requestid.New(),
		NewRequestLogger(instance.log).Middleware(),
		// 拦截错误并返回
		func(c fiber.Ctx) error {
			if err := c.Next(); err != nil {
				return c.JSON(Response{Code: -1, Msg: err.Error()})
			}
			return nil
		},
	)
	if p, err := initPostgres(cfg.Postgres, NewGormLogger(instance.log)); err != nil {
		log.Fatal(err)
	} else {
		instance.db = p
	}
	if r, err := initRedis(cfg.Redis); err != nil {
		log.Fatal(err)
	} else {
		instance.cache = r
	}
	return instance
}

func (b *Bint) SetControllers(controllers ...Controller) *Bint {
	autoGenerateRoutes(b, controllers...)
	return b
}

func (b *Bint) SetModels(models ...any) *Bint {
	_ = b.db.AutoMigrate(models...)
	return b
}

func (b *Bint) SetMiddlewares(middlewares ...Handler) *Bint {
	for _, fn := range middlewares {
		b.app.Use(func(c fiber.Ctx) error {
			return fn(Context{Ctx: c, DB: b.DB(), Cache: b.Cache()})
		})
	}
	return b
}

func (b *Bint) DB() *gorm.DB {
	return b.db
}

func (b *Bint) Cache() *redis.Client {
	return b.cache
}

func (b *Bint) Log() *slog.Logger {
	return b.log
}

func (b *Bint) App() *fiber.App {
	return b.app
}

func (b *Bint) Config() *Config {
	return b.cfg
}

func (b *Bint) Run() error {
	// 打印路由
	routes := b.app.GetRoutes(true)
	b.log.Debug("bint.Run()", "routes", routes, "routes_len", len(routes))
	// 监听地址
	listen := ":8080"
	if b.cfg.Listen != "" {
		listen = b.cfg.Listen
	}
	// 启动监听
	return b.app.Listen(listen, fiber.ListenConfig{DisableStartupMessage: true})
}
