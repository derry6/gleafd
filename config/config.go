package config

import (
	"fmt"
)

const (
	envPrefix = "GLEAFD_"
)

type SegmentConfig struct {
	Enable bool   `yaml:"enable"`
	DBHost string `yaml:"db_host"`
	DBName string `yaml:"db_name"`
	DBUser string `yaml:"db_user"`
	DBPass string `yaml:"db_pass"`
}

func (c *SegmentConfig) DBUrl() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",
		c.DBUser, c.DBPass, c.DBHost, c.DBName)
}

type SnowflakeConfig struct {
	Enable        bool   `yaml:"enable"`
	RedisAddresss string `yaml:"redis_addr"`
}

type Config struct {
	Name      string          `yaml:"name"`
	Addr      string          `yaml:"addr"`
	Log       string          `yaml:"log"`
	Segment   SegmentConfig   `yaml:"segment"`
	Snowflake SnowflakeConfig `yaml:"snowflake"`
}

func newConfig() *Config {
	return &Config{
		Name: "gleafd0",
		Addr: ":9060",
		Log:  "info",
		Segment: SegmentConfig{
			Enable: true,
			DBHost: "127.0.0.1:5506",
			DBName: "gleafd",
			DBUser: "gleafd",
			DBPass: "123456",
		},
		Snowflake: SnowflakeConfig{
			Enable:        true,
			RedisAddresss: "127.0.0.1:8379",
		},
	}
}
