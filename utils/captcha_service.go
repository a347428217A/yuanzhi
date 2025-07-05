package utils

import (
	"admin-api/common/constant"
	"admin-api/pkg/redis"
	"context"
	"image/color"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
)

type CaptchaService struct {
	store RedisStore // 添加RedisStore作为成员
}

// 创建新的验证码服务实例
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{
		store: RedisStore{},
	}
}

// GenerateCaptcha 生成图形验证码
func (cs *CaptchaService) GenerateCaptcha() (id, b64s string, err error) {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	// 配置验证码

	driver := base64Captcha.DriverString{
		Height:          80,  // 增加高度
		Width:           240, // 增加宽度
		NoiseCount:      10,  // 增加噪点数量
		ShowLineOptions: base64Captcha.OptionShowSineLine | base64Captcha.OptionShowSlimeLine | base64Captcha.OptionShowHollowLine,
		Length:          4,
		Source:          "23456789ABCDEFGHJKLMNPQRSTUVWXYZ", // 去掉了容易混淆的字符
		BgColor: &color.RGBA{
			R: uint8(rand.Intn(100) + 155), // 随机背景色(R)
			G: uint8(rand.Intn(100) + 155), // 随机背景色(G)
			B: uint8(rand.Intn(100) + 155), // 随机背景色(B)
			A: 255,
		},
		Fonts: []string{"RitaSmith.ttf", "actionj.ttf", "chromohv.ttf"}, // 多种字体
		//FontSize:    50,                                                       // 增大字体大小
		//CurveNumber: 2,                                                        // 增加曲线干扰
	}

	captcha := base64Captcha.NewCaptcha(driver.ConvertFonts(), cs.store) // 使用成员store
	return captcha.Generate()
}

// Verify 验证验证码 (新增方法)
func (cs *CaptchaService) Verify(id, answer string) bool {
	return cs.store.Verify(id, answer, true) // true表示验证后删除
}

// RedisStore 验证码存储实现
type RedisStore struct{}

var ctx = context.Background()

// Set 存储验证码
func (r RedisStore) Set(id string, value string) {
	key := constant.LOGIN_CODE_KEY_PREFIX + id
	err := redis.RedisDb.Set(ctx, key, value, 5*time.Minute).Err()
	if err != nil {
		log.Printf("验证码存储失败: %v", err)
	}
}

// Get 获取验证码
func (r RedisStore) Get(id string, clear bool) string {
	key := constant.LOGIN_CODE_KEY_PREFIX + id
	val, err := redis.RedisDb.Get(ctx, key).Result()
	if err != nil {
		return ""
	}

	// 验证后删除
	if clear {
		redis.RedisDb.Del(ctx, key)
	}
	return val
}

// Verify 验证验证码
func (r RedisStore) Verify(id, answer string, clear bool) bool {
	storedCode := r.Get(id, clear)
	return strings.ToLower(storedCode) == strings.ToLower(answer)
	//return strings.EqualFold(storedCode, answer)
}
