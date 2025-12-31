package main

import (
	"flag"
	"log"

	"github.com/apissl/bint"
	"gorm.io/gorm"
)

type UserHandler struct{}

// Setup 配置根路径为 “/user”，第二个返回参数是 fiber 中间件
func (u *UserHandler) Setup() (string, []bint.Handler) {
	return "user", nil
}

// Get 根据id，查询一个用户（这里id会自动识别并生成路由：/user/:id）
func (u *UserHandler) Get(c bint.Context, id int) error {
	var user User
	if err := c.DB.First(&user, id).Error; err != nil {
		return err
	}
	return c.Ret(user)
}

// Post 写入一个用户
func (u *UserHandler) Post(c bint.Context) error {
	var user User
	if err := c.Ctx.Bind().JSON(&user); err != nil {
		return err
	}
	if err := c.DB.Create(&user).Error; err != nil {
		return err
	}
	return c.Ret()
}

type User struct {
	gorm.Model
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func main() {
	// 读取参数
	var path string
	flag.StringVar(&path, "config", "config.yaml", "config file path")
	flag.Parse()

	// 加载配置
	cfg, err := bint.YamlConfig(path)
	if err != nil {
		log.Fatal(err)
	}

	// 初始应用
	app := bint.
		NewBint(cfg).                   // 传入配置 *bint.Config
		SetControllers(&UserHandler{}). // 可以同时注册多个 handler
		SetModels(&User{})              // 可同时注册多个数据模型

	// 项目需要 postgis 插件
	if err := app.DB().Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`).Error; err != nil {
		log.Fatal(err)
	}

	// 启动应用，默认监听 :8080
	if err = app.Run(); err != nil {
		log.Fatal(err)
	}
}
