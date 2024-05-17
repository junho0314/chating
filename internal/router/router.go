package router

import (
	"net/http"

	"chating_service/internal/constants"
	"chating_service/internal/controller"
	"chating_service/internal/db"
	"chating_service/internal/model"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/rs/zerolog/log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRoute(router *gin.Engine, autoMiddleware *jwt.GinJWTMiddleware) {
	router.NoRoute(autoMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Error().Msgf("No route claims: %#v", claims)
		c.JSON(http.StatusNotFound, gin.H{})
	})

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}))

	hub := controller.NewHub()
	go hub.Run()

	routerGrout := router.Group("/api")
	routerGrout.Use(autoMiddleware.MiddlewareFunc(), localCtxMiddleware())
	{
		routerGrout.POST("/", controller.RdsTest)
		routerGrout.GET("/chating/:roomId", func(c *gin.Context) {
			roomId := c.Param("roomId")
			controller.WebsocketHandler(hub, c.Writer, c.Request, roomId)
		})

	}

	router.POST("/login", autoMiddleware.LoginHandler)
	router.POST("/refresh_token", autoMiddleware.RefreshHandler)
	router.POST("/logout", autoMiddleware.LogoutHandler)
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
