package main

import (
	"chating_service/internal/config"
	"chating_service/internal/controller"
	"chating_service/internal/db"
	"chating_service/internal/router"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	authMiddleware := controller.InitJwt(&config)

	engine := gin.New()
	setupLogConfig(&config)
	db.InitDbConnection(&config)
	db.InitRedisConnection(&config)

	router.InitRoute(engine, authMiddleware)

	err = engine.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func setupLogConfig(config *config.AppConfig) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = "timestamp"
	log.Logger = log.With().Caller().Logger()

	isConsoleOutput := config.Log.Output == "console"
	if isConsoleOutput {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		zerolog.CallerFieldName = "caller"
	} else {
		logFilePath := config.Log.FilePath
		if logFilePath == "" {
			log.Fatal().Msg("log file path empty.")
			os.Exit(-1)
		}

		log.Logger = log.Output(&lumberjack.Logger{
			Filename:   fmt.Sprintf(logFilePath, time.Now().Format("2006-01-02")),
			MaxSize:    config.Log.MaxSize,    // megabytes
			MaxAge:     config.Log.MaxAge,     //max no. of days to retain old log files
			MaxBackups: config.Log.MaxBackups, //max no. of old log files
		})
	}
}

func getZerologLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel // 기본값은 Info
	}
}
