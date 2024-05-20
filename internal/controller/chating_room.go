package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"chating_service/internal/service"
)

func GetChatingRoom(ctx *gin.Context) {
	localCtx := getLocalCtx(ctx)
	log.Info().Msgf("Get chating room")
	chatingRooms, err := service.GetChatingRoom(localCtx)
	if err != nil {
		log.Error().Msgf("Failed to get chating room: %v", err)
		FailureResponse(ctx, 500)
		return
	}

	ResponseWithData(ctx, chatingRooms)
}
