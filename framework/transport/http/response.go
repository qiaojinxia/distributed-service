package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// ResponseHandler 响应处理器
type ResponseHandler struct{}

// NewResponseHandler 创建响应处理器
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// Success 成功响应
func (r *ResponseHandler) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
		TraceID: getTraceID(c),
	})
}

// Error 错误响应
func (r *ResponseHandler) Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		TraceID: getTraceID(c),
	})
}

// BadRequest 400错误
func (r *ResponseHandler) BadRequest(c *gin.Context, message string) {
	r.Error(c, http.StatusBadRequest, message)
}

// Unauthorized 401错误
func (r *ResponseHandler) Unauthorized(c *gin.Context, message string) {
	r.Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403错误
func (r *ResponseHandler) Forbidden(c *gin.Context, message string) {
	r.Error(c, http.StatusForbidden, message)
}

// NotFound 404错误
func (r *ResponseHandler) NotFound(c *gin.Context, message string) {
	r.Error(c, http.StatusNotFound, message)
}

// InternalError 500错误
func (r *ResponseHandler) InternalError(c *gin.Context, message string) {
	r.Error(c, http.StatusInternalServerError, message)
}

// ServiceUnavailable 503错误
func (r *ResponseHandler) ServiceUnavailable(c *gin.Context, message string) {
	r.Error(c, http.StatusServiceUnavailable, message)
}

// JSON 自定义JSON响应
func (r *ResponseHandler) JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}

// getTraceID 获取链路追踪ID
func getTraceID(c *gin.Context) string {
	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		return traceID
	}
	if traceID := c.GetString("trace_id"); traceID != "" {
		return traceID
	}
	return ""
}

// DefaultResponseHandler 全局响应处理器实例
var DefaultResponseHandler = NewResponseHandler()

// Success 便捷函数
func Success(c *gin.Context, data interface{}) {
	DefaultResponseHandler.Success(c, data)
}

func Error(c *gin.Context, code int, message string) {
	DefaultResponseHandler.Error(c, code, message)
}

func BadRequest(c *gin.Context, message string) {
	DefaultResponseHandler.BadRequest(c, message)
}

func Unauthorized(c *gin.Context, message string) {
	DefaultResponseHandler.Unauthorized(c, message)
}

func Forbidden(c *gin.Context, message string) {
	DefaultResponseHandler.Forbidden(c, message)
}

func NotFound(c *gin.Context, message string) {
	DefaultResponseHandler.NotFound(c, message)
}

func InternalError(c *gin.Context, message string) {
	DefaultResponseHandler.InternalError(c, message)
}

func ServiceUnavailable(c *gin.Context, message string) {
	DefaultResponseHandler.ServiceUnavailable(c, message)
}
