package handler

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	deviceService *service.DeviceService
}

func NewDeviceHandler() *DeviceHandler {
	return &DeviceHandler{deviceService: service.NewDeviceService()}
}

// CreateDevice 创建设备
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	device := &model.Device{}
	if err := c.ShouldBindJSON(device); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 验证必填字段
	if device.DevType == "" {
		Error(c, CodeBadRequest, "dev_type为必填字段")
		return
	}

	// 获取当前用户信息
	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	device, err := h.deviceService.CreateDevice(device, currentUID, role)
	if err != nil {
		logger.L().Error("创建设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	SuccessWithCode(c, 201, "创建设备成功", device)
}

// GetDevices 获取设备列表
// 普通用户：只能查看自己绑定的设备
// 管理员：可以查看所有设备
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	page, _ := c.GetQuery("page")
	pageSize, _ := c.GetQuery("page_size")
	devStatusStr, _ := c.GetQuery("dev_status")
	keyword, _ := c.GetQuery("keyword")
	sortBy, _ := c.GetQuery("sort_by")
	sortOrder, _ := c.GetQuery("sort_order")

	pageInt, _ := utils.ConvertToInt64(page)
	if pageInt <= 0 {
		pageInt = 1
	}
	pageSizeInt, _ := utils.ConvertToInt64(pageSize)
	if pageSizeInt <= 0 {
		pageSizeInt = 10
	}

	var devStatus *int
	if !utils.IsEmpty(devStatusStr) {
		status, _ := utils.ConvertToInt64(devStatusStr)
		statusInt := int(status)
		devStatus = &statusInt
	}

	// 获取当前用户信息
	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	devices, total, err := h.deviceService.GetDevices(int(pageInt), int(pageSizeInt), devStatus, keyword, sortBy, sortOrder, currentUID, role)
	if err != nil {
		logger.L().Error("获取设备列表失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	totalPages := int64((total + int64(pageSizeInt) - 1) / int64(pageSizeInt))
	if totalPages == 0 {
		totalPages = 1
	}

	Success(c, "获取设备列表成功", gin.H{
		"items": devices,
		"pagination": gin.H{
			"page":        pageInt,
			"page_size":   pageSizeInt,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// UpdateDevice 更新设备信息
// 普通用户：只能更新有（w或rw）权限的设备
// 管理员：可以更新所有设备
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	device := &model.Device{}
	if err := c.ShouldBindJSON(device); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 验证dev_id
	if device.DevID == 0 {
		Error(c, CodeBadRequest, "dev_id不能为空")
		return
	}

	// 获取当前用户信息
	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	device, err := h.deviceService.UpdateDevice(device, currentUID, role)
	if err != nil {
		logger.L().Error("更新设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "更新设备成功", device)
}

// DeleteDevice 删除设备
// 普通用户：只能删除有（w或rw）权限的设备
// 管理员：可以删除所有设备
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	devIDStr := c.Query("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil || devID == 0 {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	// 获取当前用户信息
	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	err = h.deviceService.DeleteDevice(devID, currentUID, role)
	if err != nil {
		logger.L().Error("删除设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除设备成功", nil)
}

// GetDeviceStatistics 获取设备统计信息
// 普通用户：只统计有权限的设备
// 管理员：统计所有设备
func (h *DeviceHandler) GetDeviceStatistics(c *gin.Context) {
	// 获取当前用户信息
	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	stats, err := h.deviceService.GetDeviceStatistics(currentUID, role)
	if err != nil {
		logger.L().Error("获取设备统计信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取设备统计信息成功", stats)
}
