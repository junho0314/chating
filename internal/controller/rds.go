package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func RdsTest(ctx *gin.Context) {
	localCtx := getLocalCtx(ctx)

	err := localCtx.RedisCtx.Set("test", "success")
	if err != nil {
		log.Error().Msgf("Failed to set key: %v", err)
		FailureResponse(ctx, 500)
		return
	}

	val, err := localCtx.RedisCtx.Get("test")
	if err != nil {
		panic(err)
	}
	fmt.Println("test", val)

}
