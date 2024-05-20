package service

import (
	"chating_service/internal/model"
	"chating_service/internal/repo"
)

func GetChatingRoom(localCtx *model.LocalCtx) ([]model.ChatingRoom, error) {
	chatingRooms, err := repo.FetchChatingRoom(localCtx.RdbCtx)
	if err != nil {
		return nil, err
	}
	return chatingRooms, nil
}
