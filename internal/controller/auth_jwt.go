package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"

	"chating_service/internal/config"
	"chating_service/internal/constants"
	"chating_service/internal/model"
	"chating_service/internal/repo"
	"chating_service/internal/utils"
)

var AccessSecret []byte
var RefreshSecret []byte

type CustomClaims struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func InitJwt(appConfig *config.AppConfig) gin.HandlerFunc {

	AccessSecret = []byte(appConfig.Jwt.SignKey)
	RefreshSecret = []byte(appConfig.Jwt.RefreshKey)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenString := authHeader[len("Bearer "):]
		claims := &CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return AccessSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		c.Set(constants.AccountIdField, claims.ID)
		c.Set("claims", claims)
		c.Next()
	}
}

func authenticateAccount(localCtx *model.LocalCtx, userId, password string) (model.Account, error) {

	account, err := repo.GetUserAccount(localCtx.RdbCtx, userId)
	if err != nil {
		return model.Account{}, err
	}

	if utils.IsInvalidPassword(password, account.Password) {
		return model.Account{}, errors.New("incorrect password")
	}

	return account, nil
}

func generateToken(account model.Account, secret []byte, duration time.Duration) (string, time.Time, error) {
	expire := time.Now().Add(duration)
	claims := CustomClaims{
		ID: account.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expire, nil
}

func LoginHandler(c *gin.Context, authMiddleware gin.HandlerFunc) {
	appconfig := config.GetAppConfig()

	localCtx := getLocalCtx(c)
	log.Info().Msg("LoginHandler")

	var loginForm model.UserLogin
	if err := c.ShouldBindJSON(&loginForm); err != nil {
		log.Error().Msgf("Failed to bind login form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing login values"})
		return
	}

	account, err := authenticateAccount(localCtx, loginForm.UserId, loginForm.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect username or password"})
		return
	}

	accessToken, accessTokenExpire, err := generateToken(account, AccessSecret, time.Minute*time.Duration(appconfig.Jwt.ExpireMinutes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	refreshToken, refreshTokenExpire, err := generateToken(account, RefreshSecret, time.Hour*24*time.Duration(appconfig.Jwt.RefreshDays))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	err = repo.InsertRefreshToken(account.Id, refreshToken, time.Hour*24*time.Duration(appconfig.Jwt.RefreshDays), localCtx)
	if err != nil {
		log.Error().Msgf("Failed to insert refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":   accessToken,
		"refresh_token":  refreshToken,
		"access_expire":  accessTokenExpire,
		"refresh_expire": refreshTokenExpire,
	})
}

func RefreshTokenHandler(c *gin.Context, authMiddleware gin.HandlerFunc) {
	localCtx := getLocalCtx(c)

	var refreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&refreshTokenRequest); err != nil {
		log.Error().Msgf("Failed to bind refresh token: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh token"})
		return
	}

	token, err := jwt.ParseWithClaims(refreshTokenRequest.RefreshToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return RefreshSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token claims"})
		return
	}

	storedRefreshToken, _, err := repo.GetRefreshToken(claims.ID, localCtx)
	if err != nil || storedRefreshToken != refreshTokenRequest.RefreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	account := model.Account{Id: claims.ID}
	accessToken, accessTokenExpire, err := generateToken(account, AccessSecret, time.Minute*15)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	newRefreshToken, refreshTokenExpire, err := generateToken(account, RefreshSecret, time.Hour*24*7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":   accessToken,
		"refresh_token":  newRefreshToken,
		"access_expire":  accessTokenExpire,
		"refresh_expire": refreshTokenExpire,
	})
}
