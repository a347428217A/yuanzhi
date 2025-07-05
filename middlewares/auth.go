package middlewares

import (
	"admin-api/common/constant"
	"admin-api/pkg/auth"
	"admin-api/pkg/redis"
	"admin-api/utils"
	"fmt"
	"strings"

	"context"
	"github.com/gin-gonic/gin"
)

// CustomerAuthMiddleware 用户认证中间件
func CustomerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "认证信息为空")
			c.Abort()
			return
		}

		// 格式: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := auth.ParseCustomerToken(token)
		if err != nil {
			utils.Unauthorized(c, "认证信息无效")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// MerchantAuthMiddleware 商家认证中间件
func MerchantAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "认证信息为空")
			c.Abort()
			return
		}

		// 格式: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := auth.ParseMerchantToken(token, "default_secret")
		if err != nil {
			utils.Unauthorized(c, "认证信息无效")
			c.Abort()
			return
		}

		// 将商家信息存入上下文
		c.Set("admin_id", claims.AdminID)
		c.Set("merchant_id", claims.MerchantID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func LoginGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只保护登录接口
		if c.Request.URL.Path != "/api/merchant/login" {
			c.Next()
			return
		}

		// 获取客户端IP
		clientIP := c.ClientIP()
		key := fmt.Sprintf("login:fail:%s", clientIP)

		// 检查失败次数
		failCount, err := redis.RedisDb.Get(context.Background(), key).Int()
		if err == nil && failCount >= constant.LOGIN_FAIL_LIMIT {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "登录尝试过于频繁，请5分钟后再试",
			})
			return
		}

		c.Next()

		// 登录失败后增加计数
		if c.Writer.Status() == 401 || c.Writer.Status() == 403 {
			redis.RedisDb.Incr(context.Background(), key)
			redis.RedisDb.Expire(context.Background(), key, constant.LOGIN_FAIL_LOCK_TIME)
		}
	}
}
