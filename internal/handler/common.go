package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"


)

const (
	CodeSuccess = 200
	CodeError = 400
	CodeNotFound = 404
	CodeInternalServerError = 500
	CodeBadRequest = 400
	CodeUnauthorized = 401
	CodeForbidden = 403
	CodeConflict = 409
	CodeServiceUnavailable = 503
	CodeTooManyRequests = 429
)

type CommonResponse struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}


func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, CommonResponse{
		Code: 200,
		Message: message,
		Data: data,
	})
}

func SuccessWithCode(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, CommonResponse{
		Code: code,
		Message: message,
		Data: data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, CommonResponse{
		Code: code,
		Message: message,
		Data: nil,
	})
	c.Abort()
}