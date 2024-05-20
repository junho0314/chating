package repo

import (
	"chating_service/internal/constants"
	"chating_service/internal/model"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

func InsertRefreshToken(accountId int64, token string, expiration time.Duration, localCtx *model.LocalCtx) error {

	refreshToken, _, err := GetRefreshToken(accountId, localCtx)
	if err == nil && refreshToken != "" {
		return nil
	}

	log.Info().Msgf("refresh token: %v", refreshToken)

	accountIdStr := strconv.FormatInt(accountId, 10)

	err = localCtx.RedisCtx.SetWithExpire(constants.RefreshTokenKey+accountIdStr, token, expiration)
	if err != nil {
		log.Error().Msgf("Failed to set key: %v", err)
		return err
	}

	log.Info().Msgf("Refresh token inserted: %v", token)

	return nil
}

func GetRefreshToken(accountId int64, localCtx *model.LocalCtx) (string, time.Time, error) {

	log.Info().Msgf("GetRefreshToken: %v", accountId)
	accountIdToString := strconv.FormatInt(accountId, 10)

	token, err := localCtx.RedisCtx.Get(constants.RefreshTokenKey + accountIdToString)
	if err != nil {
		return "", time.Now(), err
	}

	exp, err := localCtx.RedisCtx.GetTtl(constants.RefreshTokenKey + accountIdToString)
	if err != nil {
		return "", time.Now(), err
	}
	expire := time.Now().Add(exp)
	return token, expire, nil
}

func CheckRefreshToken(accountId int64, userToken string, localCtx *model.LocalCtx) (bool, error) {

	accountIdToString := strconv.FormatInt(accountId, 10)

	token, err := localCtx.RedisCtx.Get(constants.RefreshTokenKey + accountIdToString)
	if err != nil {
		return false, err
	}

	if token == "" {
		return false, nil
	}

	if token != userToken {
		return false, nil
	}

	return true, nil
}
