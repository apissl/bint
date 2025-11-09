package bint

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v3"
)

type Controller interface {
	Setup() (string, []Handler)
}

func autoGenerateRoutes(bint *Bint, controllers ...Controller) {
	// 注册群组、中间件
	for _, ctrl := range controllers {
		val := reflect.ValueOf(ctrl)
		typ := val.Type()

		// 创建路由组
		prefix, middlewares := ctrl.Setup()
		group := bint.app.Group(prefix)
		for _, handler := range middlewares {
			group.Use(func(c fiber.Ctx) error {
				return handler(Context{
					Ctx:   c,
					DB:    bint.DB(),
					Cache: bint.Cache(),
					Log:   bint.Log(),
				})
			})
		}

		// 注册路由方法
		for i := 0; i < val.NumMethod(); i++ {
			method := typ.Method(i)
			methodVal := val.Method(i)

			if !method.IsExported() || method.Name == "Setup" {
				continue
			}

			// 解析方法签名
			httpMethod, routePath, params, ok := parseMethod(method)
			if !ok {
				continue
			}

			handler := createSmartHandler(bint, methodVal, method.Type, params)
			registerHandler(group, httpMethod, routePath, handler)
		}
	}
}

// 方法解析
func parseMethod(method reflect.Method) (httpMethod, routePath string, params []string, ok bool) {
	name := method.Name
	// 支持的HTTP方法列表
	for _, m := range []string{"Get", "Post", "Put", "Delete", "Patch", "Head"} {
		if strings.HasPrefix(name, m) {
			// 解析请求方法及路径
			httpMethod = strings.ToUpper(m)
			basePath := camelToPath(name[len(m):])
			// 解析方法参数信息
			for i := 1; i < method.Type.NumIn(); i++ {
				pt := method.Type.In(i)
				// bint.Context 必须在首位
				if i == 1 {
					if pt != reflect.TypeOf((*Context)(nil)).Elem() {
						panic("method must take bint.Context as the first parameter")
					}
					continue
				}
				// 扫描到结构体了，退出
				if pt.Kind() == reflect.Struct {
					break
				}
				// 新增路径绑定参数
				params = append(params, fmt.Sprintf("param%d", i-1))
			}
			routePath = buildRoutePath(basePath, params)
			return httpMethod, routePath, params, true
		}
	}
	return
}

// 构建路由路径
func buildRoutePath(basePath string, params []string) string {
	if len(params) == 0 {
		return basePath
	}
	return fmt.Sprintf("%s/:%s", basePath, strings.Join(params, "/:"))
}

func camelToPath(s string) string {
	if s == "" {
		return ""
	}
	var buf strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			// 处理连续大写（如HTTP→HTTP）
			nextLower := false
			if i+1 < len(s) {
				nextLower = unicode.IsLower(rune(s[i+1]))
			}

			if !(unicode.IsUpper(rune(s[i-1])) && nextLower) {
				buf.WriteByte('-')
			}
		}
		buf.WriteRune(unicode.ToLower(r))
	}
	return "/" + buf.String()
}

// 创建带参数解析的处理函数
func createSmartHandler(bint *Bint, methodVal reflect.Value, methodType reflect.Type, params []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 准备参数列表
		args := make([]reflect.Value, 0, methodType.NumIn())

		// 第一个参数是接收器（已经通过methodVal绑定）
		// 从第二个参数开始处理
		for i := 1; i < methodType.NumIn(); i++ {
			paramType := methodType.In(i)

			// 处理上下文参数
			if paramType == reflect.TypeOf((*Context)(nil)).Elem() {
				args = append(args, reflect.ValueOf(Context{
					Ctx:   c,
					DB:    bint.DB(),
					Cache: bint.Cache(),
					Log:   bint.Log(),
				}))
				continue
			}

			// 参数解析逻辑
			if i > 1 {
				// 自动绑定请求体（须前置）
				if paramType.Kind() == reflect.Struct {
					req := reflect.New(paramType).Interface()
					if err := c.Bind().Body(req); err != nil {
						return c.Status(fiber.StatusBadRequest).SendString("invalid request body")
					}
					args = append(args, reflect.ValueOf(req).Elem())
					continue
				}
				// 自动绑定路径参数
				paramName := params[i-2]
				paramValue, err := parseParam(c, paramName, paramType)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).SendString(err.Error())
				}
				args = append(args, paramValue)
				continue
			}

			// 无法处理的参数类型
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("unsupported parameter: %s", paramType.String()))
		}

		// 调用方法
		results := methodVal.Call(args)

		// 处理返回值
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok {
				return err
			}
		}

		return nil
	}
}

// 参数解析辅助函数
func parseParam(c fiber.Ctx, name string, typ reflect.Type) (reflect.Value, error) {
	strValue := c.Params(name)
	if strValue == "" {
		return reflect.Value{}, fmt.Errorf("missing required parameter: %s", name)
	}

	switch typ.Kind() {
	case reflect.String:
		return reflect.ValueOf(strValue), nil

	case reflect.Int, reflect.Int64:
		intValue, err := strconv.Atoi(strValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid integer value %s", name)
		}
		return reflect.ValueOf(intValue), nil

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(strValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid boolean value %s", name)
		}
		return reflect.ValueOf(boolValue), nil

	default:
		return reflect.Value{}, fmt.Errorf("unsupported parameter type: %s", typ.Name())
	}
}

func registerHandler(r fiber.Router, method string, path string, handler fiber.Handler) {
	switch method {
	case "GET":
		r.Get(path, handler)
	case "POST":
		r.Post(path, handler)
	case "PUT":
		r.Put(path, handler)
	case "DELETE":
		r.Delete(path, handler)
	case "PATCH":
		r.Patch(path, handler)
	case "HEAD":
		r.Head(path, handler)
	default:
		panic(fmt.Sprintf("unsupported method: %s", method))
	}
}
