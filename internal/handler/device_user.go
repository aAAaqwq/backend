package handler

import (
	"backend/internal/middleware"
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
// 普通用户：只能将自己绑定到设备
// 管理员：可以将任意用户绑定到设备
func (h *DeviceUserHandler) BindDeviceUser(c *gin.Context) {
	req := &model.DeviceUserBindingReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 验证必填字段
	if req.DevID == 0 {
		Error(c, CodeBadRequest, "dev_id为必填字段")
		return
	}

	// 获取当前用户信息
	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	// 如果没有指定uid，默认绑定当前用户
	if req.UID == 0 {
		req.UID = currentUID
	}

	// 权限检查：普通用户只能绑定自己
	if role != "admin" && req.UID != currentUID {
		Error(c, CodeForbidden, "您只能绑定自己到设备")
		return
	}

	deviceUser, err := h.deviceUserService.BindDeviceUser(req.DevID, req, currentUID, role)
	if err != nil {
		logger.L().Error("绑定用户到设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	SuccessWithCode(c, 201, "绑定用户到设备成功", deviceUser)
}

// UnbindDeviceUser 解绑用户设备
// 普通用户：只能解绑自己（需要有w或rw权限）
// 管理员：可以解绑任意用户
func (h *DeviceUserHandler) UnbindDeviceUser(c *gin.Context) {
	var req struct {
		UID   int64 `json:"uid"`
		DevID int64 `json:"dev_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 获取当前用户信息
	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	// 如果没有指定uid，默认解绑当前用户
	if req.UID == 0 {
		req.UID = currentUID
	}

	// 权限检查：普通用户只能解绑自己
	if role != "admin" && req.UID != currentUID {
		Error(c, CodeForbidden, "您只能解绑自己")
		return
	}

	err := h.deviceUserService.UnbindDeviceUser(req.DevID, req.UID, currentUID, role)
	if err != nil {
		logger.L().Error("解绑用户设备失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "解绑用户设备成功", nil)
}

// GetDeviceUsers 获取指定设备的绑定用户列表
func (h *DeviceUserHandler) GetDeviceUsers(c *gin.Context) {
	devIDStr := c.Query("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil || devID == 0 {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	// 获取当前用户信息
	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	users, err := h.deviceUserService.GetDeviceUsers(devID, currentUID, role)
	if err != nil {
		logger.L().Error("获取设备绑定用户列表失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取设备绑定用户列表成功", gin.H{
		"bound_users": users,
	})
}
