package handler

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{userService: service.NewUserService()}
}

type getUserDevicesQuery struct {
	UID             *int64 `form:"uid" binding:"omitempty,gt=0"`
	Page            int    `form:"page" binding:"required,min=1"`
	PageSize        int    `form:"page_size" binding:"required,min=1,max=200"`
	DevType         string `form:"dev_type" binding:"omitempty"`
	DevStatus       *int   `form:"dev_status" binding:"omitempty"`
	PermissionLevel string `form:"permission_level" binding:"omitempty"`
	IsActive        *bool  `form:"is_active" binding:"omitempty"`
}

func (h *UserHandler) Register(c *gin.Context) {
	// 从请求中获取注册信息
	user := &model.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 调用服务层注册用户
	if err := h.userService.Register(user); err != nil {
		logger.L().Error("注册用户失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "注册用户成功", nil)
}

func (h *UserHandler) Login(c *gin.Context) {
	// 从请求中获取登录信息
	loginReq := &model.User{}
	if err := c.ShouldBindJSON(loginReq); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 调用服务层登录用户
	token, err := h.userService.Login(loginReq)
	if err != nil {
		logger.L().Error("登录用户失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "登录用户成功", token)
}

func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	// 从请求中获取更新用户信息
	updateUserReq := &model.UpdateUserInfoRequest{}
	if err := c.ShouldBindJSON(updateUserReq); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}
	// 从上下文获取uid
	if utils.IsEmpty(updateUserReq.UID) {
		updateUserReq.UID, _ = middleware.GetCurrentUserID(c)
	}

	user := &model.User{
		UID:      updateUserReq.UID,
		Username: updateUserReq.Username,
		Email:    updateUserReq.Email,
	}

	// 调用服务层更新用户信息
	if _, err := h.userService.UpdateUserInfo(user); err != nil {
		logger.L().Error("更新用户信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "更新用户信息成功", nil)
}

func (h *UserHandler) UpdateUserPassword(c *gin.Context) {
	// 从请求中获取更新密码信息
	updatePasswordReq := &model.ChangePasswordReq{}
	if err := c.ShouldBindJSON(updatePasswordReq); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}
	if utils.IsEmpty(updatePasswordReq.UID) {
		var ok bool
		updatePasswordReq.UID, ok = middleware.GetCurrentUserID(c)
		if !ok {
			Error(c, CodeBadRequest, "用户ID不能为空")
			return
		}
	}
	if utils.IsEmpty(updatePasswordReq.NewPassword) || utils.IsEmpty(updatePasswordReq.OldPassword) {
		Error(c, CodeBadRequest, "新密码或旧密码不能为空")
		return
	}

	// 调用服务层更新用户密码
	if err := h.userService.ChangePassword(updatePasswordReq); err != nil {
		logger.L().Error("更新用户密码失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "更新用户密码成功", nil)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	// 获取查询参数
	page, _ := c.GetQuery("page")
	pageSize, _ := c.GetQuery("page_size")
	role, _ := c.GetQuery("role")
	keyword, _ := c.GetQuery("keyword")
	sortBy, _ := c.GetQuery("sort_by")
	sortOrder, _ := c.GetQuery("sort_order")

	// 转换分页参数
	pageInt, _ := utils.ConvertToInt64(page)
	if pageInt <= 0 {
		pageInt = 1
	}
	pageSizeInt, _ := utils.ConvertToInt64(pageSize)
	if pageSizeInt <= 0 {
		pageSizeInt = 10
	}

	// 调用服务层获取用户列表
	users, total, err := h.userService.GetUsers(int(pageInt), int(pageSizeInt), role, keyword, sortBy, sortOrder)
	if err != nil {
		logger.L().Error("获取用户列表失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	// 计算总页数
	totalPages := int64((total + int64(pageSizeInt) - 1) / int64(pageSizeInt))
	if totalPages == 0 {
		totalPages = 1
	}

	// 返回分页结果
	Success(c, "获取用户列表成功", gin.H{
		"items": users,
		"pagination": gin.H{
			"page":        pageInt,
			"page_size":   pageSizeInt,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *UserHandler) GetUserDevices(c *gin.Context) {
	var query getUserDevicesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	currentUID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, CodeUnauthorized, "未获取到用户身份信息")
		return
	}
	currentRole, _ := middleware.GetCurrentUserRole(c)

	targetUID := currentUID
	if query.UID != nil {
		targetUID = *query.UID
		if targetUID != currentUID && currentRole != model.RoleAdmin {
			Error(c, CodeForbidden, "仅管理员可查看其他用户的设备")
			return
		}
	}

	devices, total, err := h.userService.GetUserDevices(
		targetUID,
		query.Page,
		query.PageSize,
		query.DevType,
		query.DevStatus,
		query.PermissionLevel,
		query.IsActive,
	)
	if err != nil {
		logger.L().Error("获取用户设备列表失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	totalPages := int64((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	Success(c, "获取用户设备列表成功", gin.H{
		"items": []gin.H{
			{
				"dev_list": devices,
			},
		},
		"pagination": gin.H{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 从请求中获取用户ID
	uid, _ := c.GetQuery("uid")
	uidInt, _ := utils.ConvertToInt64(uid)
	if utils.IsEmpty(uidInt) {
		Error(c, CodeBadRequest, "用户ID不能为空")
		return
	}

	// 调用服务层删除用户
	if err := h.userService.DeleteUser(uidInt); err != nil {
		logger.L().Error("删除用户失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除用户成功", nil)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	user := &model.User{}
	uid, _ := c.GetQuery("uid")
	user.UID, _ = utils.ConvertToInt64(uid)
	if uid == "" {
		// 从上下文获取uid
		user.UID, _ = middleware.GetCurrentUserID(c)
		logger.L().Info("从上下文获取uid", logger.WithAny("uid", user.UID))
	}

	user.Email, _ = c.GetQuery("email")
	user.Username, _ = c.GetQuery("username")

	if utils.IsEmpty(user.UID) && utils.IsEmpty(user.Email) && utils.IsEmpty(user.Username) {
		Error(c, CodeBadRequest, "uid or email or username is required")
		return
	}
	// logger.L().Info("UID", logger.WithAny("uid", user.UID))
	user, err := h.userService.GetUser(user)
	if err != nil {
		logger.L().Error("获取用户失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取用户成功", user)
}
