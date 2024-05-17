package model

import "chating_service/internal/db"

type LocalCtx struct {
	AccountId int64     // user account Id
	RdbCtx    *db.DbCtx // db connection
	RedisCtx  *db.RedisCtx
}
