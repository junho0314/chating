package repo

import (
	"chating_service/internal/db"
	"chating_service/internal/model"

	"github.com/rs/zerolog/log"
)

func FetchChatingRoom(dbCtx *db.DbCtx) ([]model.ChatingRoom, error) {
	selectSQL := `
		SELECT 
			id,
			name
		FROM CHATING_ROOM
	`
	rows, err := dbCtx.DB.Query(selectSQL)
	if err != nil {
		log.Error().Msgf("Failed to fetch chating room: %v", err)
		return nil, err
	}
	defer rows.Close()

	var chatingRooms []model.ChatingRoom
	for rows.Next() {
		var chatingRoom model.ChatingRoom
		err := rows.Scan(
			&chatingRoom.RoomId,
			&chatingRoom.RoomName,
		)
		if err != nil {
			log.Error().Msgf("Failed to scan chating room: %v", err)
			return nil, err
		}
		chatingRooms = append(chatingRooms, chatingRoom)
	}

	return chatingRooms, nil
}
