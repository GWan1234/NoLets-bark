package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sunvc/NoLets/controller"
)

func SetupRouter(router *gin.Engine) {

	router.GET("/", controller.Home)
	router.GET("/info", controller.Info)

	// App内部使用
	router.GET("/ping", controller.Ping)
	router.GET("/health", controller.Health)
	router.GET("/monitor", controller.GetServerInfo)

	// 注册
	router.GET("/register/:deviceKey", GCMDecryptMiddleware(), controller.Register)
	router.POST("/register", GCMDecryptMiddleware(), controller.Register)

	router.GET("/upload", controller.Upload)
	router.POST("/upload", controller.Upload)
	wellKnowGroup := router.Group("/.well-known")
	{
		wellKnowGroup.GET("/apple-app-site-association", controller.AppleSite)
	}
	// 推送请求
	router.POST("/push", controller.BasePush)
	// 获取设备Token
	router.GET("/:deviceKey/token", controller.GetDeviceToken)
	// title subtitle body
	router.GET("/:deviceKey/:params1/:params2/:params3", controller.BasePush)
	router.POST("/:deviceKey/:params1/:params2/:params3", controller.BasePush)
	// title body
	router.GET("/:deviceKey/:params1/:params2", controller.BasePush)
	router.POST("/:deviceKey/:params1/:params2", controller.BasePush)
	// body
	router.GET("/:deviceKey/:params1", controller.BasePush)
	router.POST("/:deviceKey/:params1", controller.BasePush)

	// 参数化的推送
	router.GET("/:deviceKey", CheckDotParamMiddleware(), controller.BasePush)
	router.POST("/:deviceKey", controller.BasePush)
}
