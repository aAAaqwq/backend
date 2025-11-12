package route

import (
	"backend/internal/handler"

	"github.com/gin-gonic/gin"
)


func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// user相关接口
		userHandler := handler.NewUserHandler()
		api.POST("/users/register", userHandler.Register)
		api.POST("/users/login", userHandler.Login)
		api.GET("/users", userHandler.GetUser)
		api.PUT("/users", userHandler.UpdateUserInfo)
		api.PUT("/users/password", userHandler.UpdateUserPassword)
		api.DELETE("/users", userHandler.DeleteUser)

		// // device相关接口
		// api.POST("/devices", createDevice)
		// api.GET("/devices/:dev_id", getDevice)
		// api.PUT("/devices/:dev_id", updateDevice)
		// api.DELETE("/devices/:dev_id", deleteDevice)

		// // sensor data相关接口
		// api.POST("/sensor-data", createSensorData)
		// api.GET("/sensor-data/:data_id", getSensorData)


		// // warning info相关接口
		// api.POST("/warning-info", createWarningInfo)
		// api.GET("/warning-info/:alert_id", getWarningInfo)
		// api.PUT("/warning-info/:alert_id", updateWarningInfo)
		// api.DELETE("/warning-info/:alert_id", deleteWarningInfo)

		// // metadata相关接口
		// api.POST("/metadata", createMetadata)
		// api.GET("/metadata/:data_id", getMetadata)

		// // 日志相关接口
		// api.GET("/logs", getLogs)
		// api.GET("/logs/:log_id", getLog)
		// api.DELETE("/logs/:log_id", deleteLog)

		// // 系统相关接口
		// api.GET("/system/statistics", getSystemStatistics)
		// api.GET("/system/logs", getSystemLogs)
		// api.GET("/system/logs/:log_id", getSystemLog)
		// api.DELETE("/system/logs/:log_id", deleteSystemLog)
	}
}