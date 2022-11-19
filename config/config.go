package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Id      string `mapstructure:"id"`
		Version string `mapstructure:"version"`
	} `mapstructure:"app"`
	Stream struct {
		HeaderLength int64 `mapstructure:"headerLength"` // header length of the expected msg
	} `mapstructure:"stream"`
	OrderBook struct {
		Depth int // depth of the printed market depth
	}
}

func NewConfig() *Config {
	env := strings.ToLower(os.Getenv("ENV"))
	if env == "" {
		// default to dev
		env = "dev"
	}

	viper.SetConfigName(fmt.Sprintf("config-%s", env))
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("unable to initialize config", err)
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("unable to initialize config", err)
	}

	return &config
}
