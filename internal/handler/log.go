package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/utils"
	"fmt"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *service.LogService
}

func NewLogHandler() *LogHandler {
	return &LogHandler{logService: service.NewLogService()}
}

// CreateLog 日志上传
func (h *LogHandler) CreateLog(c *gin.Context) {
	log := &model.Log{}
	if err := c.ShouldBindJSON(log); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 验证必填字段
	if log.Type == "" || log.Message == "" || log.UserAgent == "" {
		Error(c, CodeBadRequest, "缺少必填字段")
		return
	}

	log, err := h.logService.CreateLog(log)
	if err != nil {
		logger.L().Error("创建日志失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "日志上传成功", gin.H{
		"log_id": fmt.Sprintf("%d", log.LogID),
	})
}

// GetLogs 日志查询
func (h *LogHandler) GetLogs(c *gin.Context) {
	logIDStr, _ := c.GetQuery("log_id")
	startTime, _ := c.GetQuery("start_time")
	endTime, _ := c.GetQuery("end_time")
	logType, _ := c.GetQuery("type")
	levelStr, _ := c.GetQuery("level")

	// 如果提供了 log_id，直接查询单个日志
	if logIDStr != "" {
		logID, err := utils.ConvertToInt64(logIDStr)
		if err != nil {
			Error(c, CodeBadRequest, "无效的日志ID")
			return
		}

		log, err := h.logService.GetLog(logID)
		if err != nil {
			logger.L().Error("查询日志失败", logger.WithError(err))
			Error(c, CodeNotFound, "日志不存在")
			return
		}

		Success(c, "日志查询成功", gin.H{
			"log_id":  fmt.Sprintf("%d", log.LogID),
			"message": log.Message,
		})
		return
	}

	// 查询日志列表
	var level *int
	if levelStr != "" {
		levelInt, err := utils.ConvertToInt64(levelStr)
		if err == nil {
			levelValue := int(levelInt)
			level = &levelValue
		}
	}

	logs, err := h.logService.GetLogs(logType, level, startTime, endTime)
	if err != nil {
		logger.L().Error("查询日志列表失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "日志查询成功", gin.H{
		"logs": logs,
	})
}

// DeleteLog 日志删除
func (h *LogHandler) DeleteLog(c *gin.Context) {
	logIDStr := c.Query("log_id")
	if logIDStr == "" {
		Error(c, CodeBadRequest, "缺少日志ID")
		return
	}

	logID, err := utils.ConvertToInt64(logIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的日志ID")
		return
	}

	err = h.logService.DeleteLog(logID)
	if err != nil {
		logger.L().Error("删除日志失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "日志删除成功", nil)
}
