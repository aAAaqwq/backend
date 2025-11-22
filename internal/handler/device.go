package handler

import (
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

	device, err := h.deviceService.CreateDevice(device)
	if err != nil {
		logger.L().Error("创建设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	SuccessWithCode(c, 201, "创建设备成功", device)
}

// GetDevices 获取设备列表:用户显示自己所有设备，管理员显示所有设备
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	page, _ := c.GetQuery("page")
	pageSize, _ := c.GetQuery("page_size")
	devType, _ := c.GetQuery("dev_type")
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

	devices, total, err := h.deviceService.GetDevices(int(pageInt), int(pageSizeInt), devType, devStatus, keyword, sortBy, sortOrder)
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

// GetDevice 获取指定设备
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	device, err := h.deviceService.GetDevice(devID)
	if err != nil {
		logger.L().Error("获取设备失败", logger.WithError(err))
		Error(c, CodeNotFound, "设备不存在")
		return
	}

	// 获取设备绑定的用户列表
	deviceUserService := service.NewDeviceUserService()
	// 从上下文获取当前用户信息（由JWT中间件设置）
	currentUID, _ := c.Get("uid")
	currentRole, _ := c.Get("role")
	var uidInt64 int64
	var roleStr string
	if currentUID != nil {
		uidInt64, _ = currentUID.(int64)
	}
	if currentRole != nil {
		roleStr, _ = currentRole.(string)
	}
	boundUsers, err := deviceUserService.GetDeviceUsers(devID, uidInt64, roleStr)
	if err != nil {
		// 如果获取绑定用户失败，返回空列表（可能是权限问题）
		boundUsers = []*model.DeviceUserWithInfo{}
		logger.L().Warn("获取设备绑定用户失败", logger.WithError(err))
	}

	Success(c, "获取设备成功", gin.H{
		"data":        device,
		"bound_users": boundUsers,
	})
}

// UpdateDevice 更新设备
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	device := &model.Device{}
	if err := c.ShouldBindJSON(device); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	device.DevID = devID
	device, err = h.deviceService.UpdateDevice(device)
	if err != nil {
		logger.L().Error("更新设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "更新设备成功", device)
}

// DeleteDevice 删除设备
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	err = h.deviceService.DeleteDevice(devID)
	if err != nil {
		logger.L().Error("删除设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除设备成功", nil)
}

// GetDeviceStatistics 获取设备统计信息
func (h *DeviceHandler) GetDeviceStatistics(c *gin.Context) {
	stats, err := h.deviceService.GetDeviceStatistics()
	if err != nil {
		logger.L().Error("获取设备统计信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取设备统计信息成功", stats)
}
