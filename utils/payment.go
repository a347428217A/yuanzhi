package utils

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"time"
)

// GenerateTradeNo 生成交易流水号
// prefix: P-支付, R-退款
func GenerateTradeNo(prefix string) string {
	// 时间戳（纳秒）
	timestamp := time.Now().UnixNano()

	// 随机数
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomStr := fmt.Sprintf("%x", randomBytes)

	// 组合生成
	return fmt.Sprintf("%s%d%s", prefix, timestamp, randomStr)
}

// ConvertToYuan 分转元
func ConvertToYuan(fen int) float64 {
	return float64(fen) / 100.0
}

// ConvertToFen 元转分
func ConvertToFen(yuan float64) int {
	return int(yuan * 100)
}

// FormatAmount 格式化金额显示
func FormatAmount(amount int) string {
	yuan := float64(amount) / 100.0
	return strconv.FormatFloat(yuan, 'f', 2, 64)
}
