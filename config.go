package bint

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen   string         `yaml:"listen"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
}

func YamlConfig(path string) (cfg *Config, err error) {
	// 读取配置文件
	var buff []byte
	if buff, err = os.ReadFile(path); err != nil {
		return
	}
	// 解析为bint配置
	cfg = new(Config)
	if err = yaml.Unmarshal(buff, cfg); err != nil {
		return
	}
	return
}
