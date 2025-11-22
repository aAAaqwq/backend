package route

import (
	"backend/internal/handler"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// user相关接口
		userHandler := handler.NewUserHandler()
		api.POST("/users/register", userHandler.Register)
		api.POST("/users/login", userHandler.Login)
		api.GET("/users/all", middleware.JWTAuthMiddleware(), middleware.AdminOnlyMiddleware(), userHandler.GetUsers)
		api.GET("/users", middleware.JWTAuthMiddleware(), userHandler.GetUser)
		api.PUT("/users", middleware.JWTAuthMiddleware(), userHandler.UpdateUserInfo)
		api.PUT("/users/password", middleware.JWTAuthMiddleware(), userHandler.UpdateUserPassword)
		api.DELETE("/users", middleware.JWTAuthMiddleware(), middleware.AdminOnlyMiddleware(), userHandler.DeleteUser)
		api.GET("/users/:uid/devices", middleware.JWTAuthMiddleware(), userHandler.GetUserDevices)

		// device相关接口
		deviceHandler := handler.NewDeviceHandler()
		api.POST("/devices", middleware.JWTAuthMiddleware(), deviceHandler.CreateDevice)
		api.GET("/devices", middleware.JWTAuthMiddleware(), deviceHandler.GetDevices)
		api.GET("/devices/:dev_id", middleware.JWTAuthMiddleware(), deviceHandler.GetDevice)
		api.PUT("/devices/:dev_id", middleware.JWTAuthMiddleware(), deviceHandler.UpdateDevice)
		api.DELETE("/devices/:dev_id", middleware.JWTAuthMiddleware(), deviceHandler.DeleteDevice)
		api.GET("/devices/statistics", middleware.JWTAuthMiddleware(), deviceHandler.GetDeviceStatistics)

		// 设备用户绑定相关接口
		deviceUserHandler := handler.NewDeviceUserHandler()
		api.POST("/devices/:dev_id/users", middleware.JWTAuthMiddleware(), deviceUserHandler.BindDeviceUser)
		api.GET("/devices/:dev_id/users", middleware.JWTAuthMiddleware(), deviceUserHandler.GetDeviceUsers)
		api.PUT("/devices/:dev_id/users/:uid", middleware.JWTAuthMiddleware(), deviceUserHandler.UpdateDeviceUser)
		api.DELETE("/devices/:dev_id/users/:uid", middleware.JWTAuthMiddleware(), deviceUserHandler.UnbindDeviceUser)

		// sensor data相关接口
		sensorDataHandler := handler.NewSensorDataHandler()
		api.POST("/device/data/timeseries", sensorDataHandler.UploadSeriesData)
		api.POST("/device/data/file", sensorDataHandler.UploadFileData)
		api.GET("/device/data/timeseries", sensorDataHandler.GetSeriesData)
		api.GET("/device/data/file/list", sensorDataHandler.GetFileList)
		api.GET("/device/data/file/download", sensorDataHandler.DownloadFile)
		api.GET("/device/:dev_id/data/statistic", sensorDataHandler.GetSensorDataStatistic)
		api.DELETE("/device/:dev_id/data/:data_id", sensorDataHandler.DeleteSensorData)

		// warning info相关接口
		warningHandler := handler.NewWarningInfoHandler()
		api.POST("/warning_info", warningHandler.CreateWarningInfo)
		api.GET("/warning_info", warningHandler.GetWarningInfoList)
		api.GET("/warning_info/:alert_id", warningHandler.GetWarningInfo)
		api.PUT("/warning_info/:alert_id", warningHandler.UpdateWarningInfo)
		api.DELETE("/warning_info/:alert_id", warningHandler.DeleteWarningInfo)

		// 日志相关接口
		logHandler := handler.NewLogHandler()
		api.POST("/logs", logHandler.UploadLog)
		api.GET("/logs", logHandler.GetLogs)
	}
}
