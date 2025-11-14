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

	pageInt, _ := utils.ConvertToInt64(page)
	if pageInt <= 0 {
		pageInt = 1
	}
	pageSizeInt, _ := utils.ConvertToInt64(pageSize)
	if pageSizeInt <= 0 {
		pageSizeInt = 10
	}

	warnings, total, err := h.warningService.GetWarningInfoList(int(pageInt), int(pageSizeInt), alertType, alertStatus)
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
		"items": warnings,
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
	alertIDStr := c.Param("alert_id")
	alertID, err := utils.ConvertToInt64(alertIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的告警ID")
		return
	}

	warning := &model.WarningInfo{}
	if err := c.ShouldBindJSON(warning); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	warning.AlertID = alertID
	warning, err = h.warningService.UpdateWarningInfo(warning)
	if err != nil {
		logger.L().Error("更新告警信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "更新告警信息成功", warning)
}

// DeleteWarningInfo 删除告警信息
func (h *WarningInfoHandler) DeleteWarningInfo(c *gin.Context) {
	alertIDStr := c.Param("alert_id")
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
