package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type DeviceUserHandler struct {
	deviceUserService *service.DeviceUserService
}

func NewDeviceUserHandler() *DeviceUserHandler {
	return &DeviceUserHandler{deviceUserService: service.NewDeviceUserService()}
}

// BindDeviceUser 绑定用户到设备
func (h *DeviceUserHandler) BindDeviceUser(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	req := &model.DeviceUserBindingReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 获取当前用户信息（由JWT中间件设置）
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

	deviceUser, err := h.deviceUserService.BindDeviceUser(devID, req, uidInt64, roleStr)
	if err != nil {
		logger.L().Error("绑定用户到设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	SuccessWithCode(c, 201, "绑定用户到设备成功", deviceUser)
}

// GetDeviceUsers 获取设备的绑定用户列表
func (h *DeviceUserHandler) GetDeviceUsers(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	// 获取当前用户信息（由JWT中间件设置）
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

	users, err := h.deviceUserService.GetDeviceUsers(devID, uidInt64, roleStr)
	if err != nil {
		logger.L().Error("获取设备绑定用户列表失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取设备绑定用户列表成功", gin.H{
		"items": users,
	})
}

// UpdateDeviceUser 更新设备用户绑定关系
func (h *DeviceUserHandler) UpdateDeviceUser(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	uidStr := c.Param("uid")

	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	uid, err := utils.ConvertToInt64(uidStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的用户ID")
		return
	}

	req := &model.DeviceUserUpdateReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 获取当前用户信息（由JWT中间件设置）
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

	deviceUser, err := h.deviceUserService.UpdateDeviceUser(devID, uid, req, uidInt64, roleStr)
	if err != nil {
		logger.L().Error("更新设备用户绑定关系失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "更新设备用户绑定关系成功", deviceUser)
}

// UnbindDeviceUser 解绑用户设备
func (h *DeviceUserHandler) UnbindDeviceUser(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	uidStr := c.Param("uid")

	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	uid, err := utils.ConvertToInt64(uidStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的用户ID")
		return
	}

	// 获取当前用户信息（由JWT中间件设置）
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

	err = h.deviceUserService.UnbindDeviceUser(devID, uid, uidInt64, roleStr)
	if err != nil {
		logger.L().Error("解绑用户设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "解绑用户设备成功", nil)
}
