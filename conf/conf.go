package conf

import (
	"fmt"

	"github.com/spf13/viper"
)

type confInfo struct {
	Name string
	Type string
	Path string
}

// 对viper再次进行包装,方便增加属性
type config struct {
	viper *viper.Viper
}

var (
	Conf *config
)

func init() {
	ci := confInfo{
		Name: "confs",
		Type: "yaml",
		Path: ".",
	}
	Conf = &config{viper: getConf(ci)}
}

func getConf(ci confInfo) *viper.Viper {
	v := viper.New()
	v.SetConfigName(ci.Name)
	v.SetConfigType(ci.Type)
	if ci.Path == "" {
		v.AddConfigPath(".")
	} else {
		v.AddConfigPath(ci.Path)
	}

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	return v
}

func (c *config) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *config) GetInt(key string) int {
	return c.viper.GetInt(key)
}

func (c *config) GetInt64(key string) int64 {
	return c.viper.GetInt64(key)
}
