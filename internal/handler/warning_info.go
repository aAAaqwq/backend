package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type WarningInfoHandler struct {
	warningService *service.WarningInfoService
}

func NewWarningInfoHandler() *WarningInfoHandler {
	return &WarningInfoHandler{warningService: service.NewWarningInfoService()}
}

// CreateWarningInfo 上传告警信息
func (h *WarningInfoHandler) CreateWarningInfo(c *gin.Context) {
	warning := &model.WarningInfo{}
	if err := c.ShouldBindJSON(warning); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	warning, err := h.warningService.CreateWarningInfo(warning)
	if err != nil {
		logger.L().Error("上传告警信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	SuccessWithCode(c, 201, "上传告警信息成功", warning)
}

// GetWarningInfoList 查询告警信息列表
func (h *WarningInfoHandler) GetWarningInfoList(c *gin.Context) {
	page, _ := c.GetQuery("page")
	pageSize, _ := c.GetQuery("page_size")
	alertType, _ := c.GetQuery("alert_type")
	alertStatus, _ := c.GetQuery("alert_status")
	devIDStr, _ := c.GetQuery("dev_id")
	dataIDStr, _ := c.GetQuery("data_id")

	pageInt, _ := utils.ConvertToInt64(page)
	if pageInt <= 0 {
		pageInt = 1
	}
	pageSizeInt, _ := utils.ConvertToInt64(pageSize)
	if pageSizeInt <= 0 {
		pageSizeInt = 10
	}

	// 转换 dev_id 和 data_id
	var devID, dataID *int64
	if devIDStr != "" {
		devIDInt, err := utils.ConvertToInt64(devIDStr)
		if err == nil {
			devID = &devIDInt
		}
	}
	if dataIDStr != "" {
		dataIDInt, err := utils.ConvertToInt64(dataIDStr)
		if err == nil {
			dataID = &dataIDInt
		}
	}

	warnings, total, err := h.warningService.GetWarningInfoList(int(pageInt), int(pageSizeInt), alertType, alertStatus, devID, dataID)
	if err != nil {
		logger.L().Error("查询告警信息列表失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	totalPages := int64((total + int64(pageSizeInt) - 1) / int64(pageSizeInt))
	if totalPages == 0 {
		totalPages = 1
	}

	Success(c, "查询告警信息列表成功", gin.H{
		"warning_lists": warnings,
		"pagination": gin.H{
			"page":        pageInt,
			"page_size":   pageSizeInt,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetWarningInfo 查询单个告警信息
func (h *WarningInfoHandler) GetWarningInfo(c *gin.Context) {
	alertIDStr := c.Param("alert_id")
	alertID, err := utils.ConvertToInt64(alertIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的告警ID")
		return
	}

	warning, err := h.warningService.GetWarningInfo(alertID)
	if err != nil {
		logger.L().Error("查询告警信息失败", logger.WithError(err))
		Error(c, CodeNotFound, "告警信息不存在")
		return
	}

	Success(c, "查询告警信息成功", warning)
}

// UpdateWarningInfo 更新告警信息
func (h *WarningInfoHandler) UpdateWarningInfo(c *gin.Context) {
	// 从查询参数获取 alert_id 和 alert_status
	alertIDStr := c.Query("alert_id")
	alertStatus := c.Query("alert_status")

	// 验证参数
	if alertIDStr == "" {
		Error(c, CodeBadRequest, "缺少告警ID")
		return
	}
	if alertStatus == "" {
		Error(c, CodeBadRequest, "缺少告警状态")
		return
	}

	alertID, err := utils.ConvertToInt64(alertIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的告警ID")
		return
	}

	// 验证 alert_status 的有效值
	if alertStatus != "active" && alertStatus != "resolved" && alertStatus != "ignored" {
		Error(c, CodeBadRequest, "无效的告警状态，必须是 active/resolved/ignored")
		return
	}

	// 调用服务层更新告警状态
	warning, err := h.warningService.UpdateWarningStatus(alertID, alertStatus)
	if err != nil {
		logger.L().Error("更新告警信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "更新告警信息成功", warning)
}

// DeleteWarningInfo 删除告警信息
func (h *WarningInfoHandler) DeleteWarningInfo(c *gin.Context) {
	// 从查询参数获取 alert_id
	alertIDStr := c.Query("alert_id")
	if alertIDStr == "" {
		Error(c, CodeBadRequest, "缺少告警ID")
		return
	}

	alertID, err := utils.ConvertToInt64(alertIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的告警ID")
		return
	}

	err = h.warningService.DeleteWarningInfo(alertID)
	if err != nil {
		logger.L().Error("删除告警信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除告警信息成功", nil)
}
