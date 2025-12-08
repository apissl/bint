package bint

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/binder"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/utils/v2"
)

type RequestLogger struct {
	log *slog.Logger
}

// NewRequestLogger 请求日志
func NewRequestLogger(logger *slog.Logger) *RequestLogger {
	return &RequestLogger{log: logger}
}

// Middleware 日志中间件
func (r *RequestLogger) Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		header := make(map[string]string)
		for key, value := range c.Request().Header.All() {
			header[string(key)] = string(value)
		}
		defer func() {
			r.log.Info("request_log",
				slog.Int("code", c.Response().StatusCode()),
				slog.String("rid", requestid.FromContext(c)),
				slog.Int64("latency", time.Since(start).Milliseconds()),
				slog.String("method", c.Method()),
				slog.String("host", c.Host()),
				slog.String("path", c.Path()),
				slog.String("ip", c.IP()),
				slog.Any("header", header),
				slog.Any("query", c.Queries()),
				slog.Any("req_data", getRequestData(c)),
				slog.Any("resp_data", getResponseData(c)),
			)
		}()
		return c.Next()
	}
}

func getRequestData(c fiber.Ctx) any {
	ctype := c.Get(fiber.HeaderContentType)
	ctype = binder.FilterFlags(utils.ParseVendorSpecificContentType(ctype))
	switch ctype {
	case "application/x-www-form-urlencoded", "multipart/form-data":
		form, err := c.MultipartForm()
		if err != nil {
			return nil
		}

		data := make(map[string]any, len(form.Value)+len(form.File))
		for k, v := range form.Value {
			data[k] = v[0]
		}
		for _, v := range form.File {
			f := v[0]
			data[f.Filename] = fmt.Sprintf("type:%s; size:%d;", f.Header.Get("Content-Type"), f.Size)
		}
		return data
	case "application/json":
		return json.RawMessage(c.Body())
	default:
		return struct{}{}
	}
}

func getResponseData(c fiber.Ctx) any {
	ctype := c.GetRespHeader(fiber.HeaderContentType)
	ctype = binder.FilterFlags(utils.ParseVendorSpecificContentType(ctype))
	switch ctype {
	case "application/json":
		return json.RawMessage(c.Response().Body())
	case "text/plain":
		return struct{}{}
	}
	return nil
}
