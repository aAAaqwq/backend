package handler

import (
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
	updateUserReq := &model.User{}
	if err := c.ShouldBindJSON(updateUserReq); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 调用服务层更新用户信息
	if _, err := h.userService.UpdateUserInfo(updateUserReq); err != nil {
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

	// 调用服务层更新用户密码
	if err := h.userService.ChangePassword(updatePasswordReq); err != nil {
		logger.L().Error("更新用户密码失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "更新用户密码成功", nil)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 从请求中获取用户ID
	uid, _ := c.GetQuery("uid")
	uidInt, _ := utils.ConvertToInt64(uid)

	// 调用服务层删除用户
	if err := h.userService.DeleteUser(uidInt); err != nil {
		logger.L().Error("删除用户失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除用户成功", nil)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	user:=&model.User{}
	uid, _ := c.GetQuery("uid")
	user.UID, _ = utils.ConvertToInt64(uid)
	user.Email,_ = c.GetQuery("email")
	user.Username,_ = c.GetQuery("username")

	if utils.IsEmpty(user.UID) && utils.IsEmpty(user.Email) && utils.IsEmpty(user.Username) {
		Error(c, CodeBadRequest, "uid or email or username is required")
		return
	}

	user, err := h.userService.GetUser(user)
	if err != nil {
		logger.L().Error("获取用户失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取用户成功",user)
}

func (h *UserHandler) GetUserDevices(c *gin.Context) {
	
}

