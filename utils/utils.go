package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, msg string) {
	Error(c, 400, msg)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, msg string) {
	Error(c, 401, msg)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, msg string) {
	Error(c, 403, msg)
}

// NotFound 404错误
func NotFound(c *gin.Context, msg string) {
	Error(c, 404, msg)
}

// InternalError 500错误
func InternalError(c *gin.Context, msg string) {
	Error(c, 500, msg)
}

// GinLogger 自定义Gin日志格式
func GinLogger(param gin.LogFormatterParams) string {
	return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %s\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		param.StatusCode,
		param.Latency,
		param.ClientIP,
		param.Method,
		param.Path,
		param.ErrorMessage,
	)
}

// GenerateFilename 生成唯一文件名
func GenerateFilename(original string) string {
	ext := filepath.Ext(original)
	timestamp := time.Now().UnixNano()
	random := rand.Intn(10000)
	return fmt.Sprintf("%d_%d%s", timestamp, random, ext)
}

// 分页响应结构体
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

func PaginatedSuccess(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	c.JSON(http.StatusOK, PaginatedResponse{
		Success:    true,
		Message:    "操作成功",
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func ToJSONString(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "{\"error\":\"json marshal failed\"}"
	}
	return string(bytes)
}
