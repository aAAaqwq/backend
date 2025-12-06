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
		api.GET("/users/bind_devices", middleware.JWTAuthMiddleware(), userHandler.GetUserDevices)

		// device相关接口
		deviceHandler := handler.NewDeviceHandler()
		api.POST("/devices", middleware.JWTAuthMiddleware(), deviceHandler.CreateDevice)
		api.GET("/devices", middleware.JWTAuthMiddleware(), deviceHandler.GetDevices)
		api.PUT("/devices", middleware.JWTAuthMiddleware(), deviceHandler.UpdateDevice)
		api.DELETE("/devices", middleware.JWTAuthMiddleware(), deviceHandler.DeleteDevice)
		api.GET("/devices/statistics", middleware.JWTAuthMiddleware(), deviceHandler.GetDeviceStatistics)

		// 设备用户绑定相关接口
		deviceUserHandler := handler.NewDeviceUserHandler()
		api.POST("/devices/bind_user", middleware.JWTAuthMiddleware(), deviceUserHandler.BindDeviceUser)
		api.DELETE("/devices/unbind_user", middleware.JWTAuthMiddleware(), deviceUserHandler.UnbindDeviceUser)
		api.GET("/devices/bind_users", middleware.JWTAuthMiddleware(), deviceUserHandler.GetDeviceUsers)

		// sensor data相关接口
		sensorDataHandler := handler.NewSensorDataHandler()
		api.POST("/device/data", middleware.JWTAuthMiddleware(), sensorDataHandler.UploadSensorData) // 统一上传接口，根据data_type判断是时序数据还是文件数据
		api.POST("/device/data/timeseries", middleware.JWTAuthMiddleware(), sensorDataHandler.GetSeriesData)
		api.DELETE("/device/data/timeseries", middleware.JWTAuthMiddleware(), sensorDataHandler.DeleteSeriesData)
		api.GET("/device/data/statistic", middleware.JWTAuthMiddleware(), sensorDataHandler.GetSensorDataStatistic)
		api.GET("/device/data/file/list", middleware.JWTAuthMiddleware(), sensorDataHandler.GetFileList)
		api.GET("/device/data/file/download", middleware.JWTAuthMiddleware(), sensorDataHandler.DownloadFile)
		api.DELETE("/device/data/file", middleware.JWTAuthMiddleware(), sensorDataHandler.DeleteFileData)

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
