package constant

import "time"

const (
	// 验证码相关
	LOGIN_CODE_KEY_PREFIX = "login:captcha:" // Redis键前缀

	// 登录失败限制
	LOGIN_FAIL_LIMIT     = 10              // 最大失败次数
	LOGIN_FAIL_LOCK_TIME = 5 * time.Minute // 锁定时间
)
