package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"chating_service/internal/constants"
	"chating_service/internal/db"
	"chating_service/internal/model"
)

func SuccessResponse(ctx *gin.Context) {
	ctx.JSON(http.StatusOK,
		gin.H{
			"code": constants.Success,
		})
}

func UpdateSuccessResponse(ctx *gin.Context) {
	ctx.JSON(http.StatusNoContent,
		gin.H{
			"code": constants.Success,
		})
}

func FailureResponse(ctx *gin.Context, code int) {
	ctx.JSON(http.StatusOK,
		gin.H{
			"code": code,
		})
}

func TokenExpireLoginResponse(ctx *gin.Context, code int, message string) {

	ctx.JSON(http.StatusUnauthorized,
		gin.H{
			"code":    code,
			"message": message,
		})
}

func SignedUrlResponse(ginCtx *gin.Context, s3ObjectKey string, signedUrl string) {
	ginCtx.JSON(http.StatusOK, gin.H{
		"code":      constants.Success,
		"timestamp": time.Now(),
		"objectKey": s3ObjectKey,
		"url":       signedUrl,
	})
}

func ResponseWithData(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, data)
}

func FailedResponseWithData(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusInternalServerError, data)
}

func LoginResponse(ctx *gin.Context, code int, access string, accessExp time.Time) {
	ctx.JSON(code,
		gin.H{
			"accessToken": access,
			"accessExp":   accessExp.Unix(),
		})
}

func LoginFailureResponse(ctx *gin.Context, responseCode int, value int) {
	ctx.JSON(http.StatusOK,
		gin.H{
			"code":            responseCode,
			"domainCodeValue": value,
		})
}

func getLocalCtx(ginCtx *gin.Context) *model.LocalCtx {
	localCtx, isExist := ginCtx.Get("localCtx")
	if !isExist {
		DbCtx := db.GetDbConnection(ginCtx)
		RdsCtx := db.GetRedisConnection(ginCtx)
		return &model.LocalCtx{
			AccountId: 0,
			RdbCtx:    &DbCtx,
			RedisCtx:  &RdsCtx,
		}
	}

	return localCtx.(*model.LocalCtx)
}
