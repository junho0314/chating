package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type MongoConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type JwtConfig struct {
	SignKey       string `mapstructure:"sign-key"`
	Realm         string `mapstructure:"realm"`
	ExpireMinutes int    `mapstructure:"expire-minutes"`
	RefreshDays   int    `mapstructure:"refresh-days"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file-path"`
	MaxSize    int    `mapstructure:"max-size"`
	MaxBackups int    `mapstructure:"max-backups"`
	MaxAge     int    `mapstructure:"max-age"`
}

type RdbConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DbName   string `mapstructure:"database"`
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

type AppConfig struct {
	Mongo MongoConfig `mapstructure:"mongo"`
	Jwt   JwtConfig   `mapstructure:"jwt"`
	Log   LogConfig   `mapstructure:"log"`
	Rdb   RdbConfig   `mapstructure:"rdb"`
	Redis RedisConfig `mapstructure:"redis"`
}

func LoadConfig() (AppConfig, error) {
	var config AppConfig

	viper.SetConfigName("configs")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Msgf("Error reading config file, %s", err.Error())
		return config, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Error().Msgf("Error reading config file, %s", err.Error())
		return config, err
	}

	return config, nil
}
