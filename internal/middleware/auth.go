package middleware

import (
	"backend/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CommonResponse 通用响应结构
type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// errorResponse 错误响应
func errorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, CommonResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
	c.Abort()
}

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "缺少Authorization头")
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			errorResponse(c, http.StatusUnauthorized, "Authorization格式错误，应为: Bearer <token>")
			return
		}

		token := parts[1]
		if token == "" {
			errorResponse(c, http.StatusUnauthorized, "token不能为空")
			return
		}

		// 解析token
		claims, err := utils.ParseToken(token)
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "无效的token: "+err.Error())
			return
		}

		// 将用户信息存储到上下文
		c.Set("uid", claims.UID)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// AdminOnlyMiddleware 仅管理员中间件
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			errorResponse(c, http.StatusForbidden, "未认证")
			return
		}

		if role != "admin" {
			errorResponse(c, http.StatusForbidden, "需要管理员权限")
			return
		}

		c.Next()
	}
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) (int64, bool) {
	uid, exists := c.Get("uid")
	if !exists {
		return 0, false
	}
	uidInt64, ok := uid.(int64)
	return uidInt64, ok
}

// GetCurrentUserRole 从上下文获取当前用户角色
func GetCurrentUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	return roleStr, ok
}
