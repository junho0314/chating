package router

import (
	"net/http"

	"chating_service/internal/constants"
	"chating_service/internal/controller"
	"chating_service/internal/db"
	"chating_service/internal/model"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRoute(router *gin.Engine, autoMiddleware gin.HandlerFunc) {
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{})
	})

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	hub := controller.NewHub()
	go hub.Run()

	routerGrout := router.Group("/api")
	routerGrout.Use(autoMiddleware, localCtxMiddleware())
	{
		routerGrout.POST("/", controller.RdsTest)

		routerGrout.GET("/chating_room", controller.GetChatingRoom)

	}

	router.GET("/chating/:roomId", func(c *gin.Context) {
		controller.WebsocketHandler(hub, c)
	})

	router.POST("/login", func(ctx *gin.Context) {
		controller.LoginHandler(ctx, autoMiddleware)
	})
	router.POST("/refresh_token", func(ctx *gin.Context) {
		controller.RefreshTokenHandler(ctx, autoMiddleware)
	})
	router.POST("/logout", func(ctx *gin.Context) {

	})
}

func localCtxMiddleware() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		accountId := ginCtx.GetInt64(constants.AccountIdField)
		if accountId <= 0 {
			ginCtx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		DbCtx := db.GetDbConnection(ginCtx)
		RdsCtx := db.GetRedisConnection(ginCtx)
		localCtx := model.LocalCtx{
			AccountId: accountId,
			RdbCtx:    &DbCtx,
			RedisCtx:  &RdsCtx,
		}
		ginCtx.Set("localCtx", &localCtx)
	}
}
