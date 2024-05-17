package controller

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"chating_service/internal/config"
	"chating_service/internal/constants"
	"chating_service/internal/model"
	"chating_service/internal/repo"
	"chating_service/internal/utils"
)

// InitJwt Initialize JWT middleware
func InitJwt(appConfig *config.AppConfig) (authMiddleware *jwt.GinJWTMiddleware) {

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Key:            []byte(appConfig.Jwt.SignKey),
		Timeout:        time.Duration(appConfig.Jwt.ExpireMinutes) * time.Minute,
		IdentityKey:    "id",
		TokenLookup:    "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:  "Bearer",
		TimeFunc:       time.Now,
		CookieSameSite: http.SameSiteDefaultMode,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if account, ok := data.(*model.Account); ok {
				encryptedAccountId := utils.EncryptAES(strconv.FormatInt(account.Id, 10))
				return jwt.MapClaims{
					"id": encryptedAccountId,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			decryptedAccountId := utils.DecryptAES(claims["id"].(string))
			accountId, _ := strconv.ParseInt(decryptedAccountId, 10, 64)
			c.Set(constants.AccountIdField, accountId)
			return &model.Account{
				Id: accountId,
			}
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if account, ok := data.(*model.Account); ok {
				c.Set(constants.AccountIdField, account.Id)

				return true
			}
			log.Error().Msg(
				"authorization error: " + data.(string))
			return false
		},
		LoginResponse: func(c *gin.Context, code int, token string, exp time.Time) {
			log.Debug().Int("LoginResponse : Status - ", code).Msg(", token:[" + token + "]")

			LoginResponse(c, code, token, exp)

		},
		LogoutResponse: func(c *gin.Context, code int) {
			log.Debug().Int("LogoutResponse : ", code)
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginForm model.UserLogin
			err := c.ShouldBindJSON(&loginForm)
			if err != nil {
				// missing userId or password
				setLoginFailureResponseCode(c, constants.InvalidCredentials, constants.Error)
				return "", jwt.ErrMissingLoginValues
			}
			var token string
			authHeader := c.GetHeader("Authorization") // "Authorization" 헤더에서 값을 가져옵니다.
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ") // "Bearer " 접두어를 제거합니다.
				// token 변수를 사용하여 로직을 계속 진행합니다.
			}

			account, err := authenticateAccount(c, appConfig, loginForm.UserId, loginForm.Password, token)
			if err != nil {
				log.Error().Msg(
					"auth error. " + err.Error())
				return nil, err
			}
			c.Set("userAccount", &account)

			return &account, nil
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			log.Debug().Int("Unauthorized", code).Msg(message)

			// If the login request is invalid
			if _, exists := c.Keys[constants.ResponseCodeKey]; exists {
				responseCode := c.Keys[constants.ResponseCodeKey].(int)
				domainCodeValue := c.Keys[constants.DomainCodeKey].(int)

				LoginFailureResponse(c, responseCode, domainCodeValue)
				return
			}

			TokenExpireLoginResponse(c, code, message)
		},
	})

	if err != nil {
		log.Error().Msg(
			"JWT Error:" + err.Error())
	}

	return authMiddleware
}

/**
* @desc Authenticate login account
*
* @param
* @error
* @return
 */
func authenticateAccount(ginCtx *gin.Context, appConfig *config.AppConfig, userId string, password string, pushToken string) (model.Account, error) {
	localCtx := getLocalCtx(ginCtx)
	account, err := repo.GetUserAccount(localCtx.RdbCtx, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Error().Msgf("user with userId : %s not found", userId)
			setLoginFailureResponseCode(ginCtx, constants.InvalidUserId, constants.Error)
			return model.Account{}, errors.New("user not found")
		}
		return model.Account{}, err
	}

	// // check account_status_code
	// if account.AccountStatus != constants.AccountStatusActive {
	// 	setLoginFailureResponseCode(ginCtx, constants.InvalidAccountStatus, account.AccountStatus)
	// 	return account, errors.New("invalid account status")
	// }

	// if the password is wrong
	if utils.IsInvalidPassword(password, account.Password) {

		setLoginFailureResponseCode(ginCtx, constants.InvalidPassword, constants.Error)

		return account, errors.New("incorrect password")
	}

	// update pushToken
	repo.InsertRefreshToken(account.Id, pushToken, time.Duration(appConfig.Jwt.RefreshDays*24)*time.Hour, localCtx)

	// push message
	//utils.SendPushNotification(account.Id, account.UserId)

	return account, nil
}

func setLoginFailureResponseCode(ginCtx *gin.Context, responseCode int, domainCodeValue int) {
	ginCtx.Set(constants.ResponseCodeKey, responseCode)
	ginCtx.Set(constants.DomainCodeKey, domainCodeValue)
}
