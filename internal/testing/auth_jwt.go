package controller

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
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
		MaxRefresh:     time.Duration(appConfig.Jwt.RefreshDays) * time.Hour * 24,
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

			account, err := authenticateAccount(c, loginForm.UserId, loginForm.Password)
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
func authenticateAccount(ginCtx *gin.Context, userId string, password string) (model.Account, error) {
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

	// if the password is wrong
	if utils.IsInvalidPassword(password, account.Password) {

		setLoginFailureResponseCode(ginCtx, constants.InvalidPassword, constants.Error)

		return account, errors.New("incorrect password")
	}

	// push message
	//utils.SendPushNotification(account.Id, account.UserId)

	return account, nil
}

func setLoginFailureResponseCode(ginCtx *gin.Context, responseCode int, domainCodeValue int) {
	ginCtx.Set(constants.ResponseCodeKey, responseCode)
	ginCtx.Set(constants.DomainCodeKey, domainCodeValue)
}

func LoginHandler(c *gin.Context, authMiddleware *jwt.GinJWTMiddleware) {
	localCtx := getLocalCtx(c)

	var loginForm model.UserLogin
	if err := c.ShouldBindJSON(&loginForm); err != nil {
		setLoginFailureResponseCode(c, constants.InvalidCredentials, constants.Error)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing login values"})
		return
	}

	account, err := authenticateAccount(c, loginForm.UserId, loginForm.Password)
	if err != nil {
		log.Error().Msg("auth error. " + err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect username or password"})
		return
	}

	// 액세스 토큰 생성
	claims := authMiddleware.PayloadFunc(account)
	accessToken, accessTokenExpire, err := authMiddleware.TokenGenerator(claims)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to generate access token"})
		return
	}

	// 리프레시 토큰 생성
	refreshClaims := jwt.MapClaims{
		"id":  utils.EncryptAES(account.UserId),
		"exp": time.Now().Add(authMiddleware.MaxRefresh).Unix(),
	}
	refreshToken, refreshTokenExpire, err := authMiddleware.TokenGenerator(refreshClaims)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to generate refresh token"})
		return
	}

	// Redis에 리프레시 토큰 저장
	userID := account.Id
	repo.InsertRefreshToken(strconv.FormatInt(userID, 10), refreshToken, authMiddleware.MaxRefresh, localCtx)
	c.JSON(http.StatusOK, gin.H{
		"access_token":   accessToken,
		"refresh_token":  refreshToken,
		"access_expire":  accessTokenExpire,
		"refresh_expire": refreshTokenExpire,
	})
}

func TokenRefreshHandler(c *gin.Context, authMiddleware *jwt.GinJWTMiddleware) {
	localCtx := getLocalCtx(c)

	accountId := c.GetInt64(constants.AccountIdField)
	if accountId <= 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	accountIdStr := strconv.FormatInt(accountId, 10)
	refreshToken, _, err := repo.GetRefreshToken(accountId, localCtx)
	if err != nil {
		log.Error().Msgf("Failed to get refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get refresh token"})
		return
	}

	ok, err := repo.CheckRefreshToken(accountId, refreshToken, localCtx)
	if err != nil {
		log.Error().Msgf("Failed to check refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check refresh token"})
		return
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	claims := authMiddleware.PayloadFunc(accountId)
	tokenString, expire, err := authMiddleware.TokenGenerator(claims)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to generate token"})
		return
	}

	repo.InsertRefreshToken(accountIdStr, tokenString, authMiddleware.MaxRefresh, localCtx)

	authMiddleware.LoginResponse(c, http.StatusOK, tokenString, expire)
}
