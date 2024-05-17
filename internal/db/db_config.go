package db

import (
	"context"
	"database/sql"
	"fmt"

	"chating_service/internal/config"

	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"

	"time"
)

var dbPool *sql.DB

const (
	MaxOpenConnections    = 20
	MaxIdleConnections    = 10
	ConnectionMaxIdleTime = 5 * time.Minute // minutes
)

func InitDbConnection(appConfig *config.AppConfig) {

	dbAddress := fmt.Sprintf("%s:%d", appConfig.Rdb.Host, appConfig.Rdb.Port)

	cfg := mysql.Config{
		User:                 appConfig.Rdb.User,
		Passwd:               appConfig.Rdb.Password,
		Net:                  "tcp",
		Addr:                 dbAddress,
		DBName:               appConfig.Rdb.DbName,
		AllowNativePasswords: true,
	}

	dbPool, _ = sql.Open(appConfig.Rdb.Driver, cfg.FormatDSN())
	//if err != nil {
	//	log.Fatal("unable to use data source name"+ err.Error())
	//}

	dbPool.SetMaxOpenConns(MaxOpenConnections)
	dbPool.SetMaxIdleConns(MaxIdleConnections)
	dbPool.SetConnMaxIdleTime(ConnectionMaxIdleTime)

	dbPool.Ping()
	//if error != nil {
	//	log.Error().Msg("InitDbConnection:: error connecting to the database. "+ err)
	//	return
	//}
	log.Info().Msg("InitDbConnection:: database connected successfully!!!")
}

func GetDbConnection(ctx context.Context) DbCtx {
	return DbCtx{
		DB:  dbPool,
		Ctx: ctx,
	}
}

func InitTestDbConnection(appConfig *config.AppConfig, connStr string) {
	dbPool, _ = sql.Open(appConfig.Rdb.Driver, connStr)

	dbPool.SetMaxOpenConns(MaxOpenConnections)
	dbPool.SetMaxIdleConns(MaxIdleConnections)
	dbPool.SetConnMaxIdleTime(ConnectionMaxIdleTime)

	dbPool.Ping()
	//if error != nil {
	//	log.Error().Msg("InitDbConnection:: error connecting to the database. "+ err.Error())
	//	return
	//}
	//log.Info("InitDbConnection:: database connected successfully!!!")
}
