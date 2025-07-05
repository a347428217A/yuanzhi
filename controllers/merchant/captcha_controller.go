package merchant

import (
	"admin-api/utils"
	"github.com/gin-gonic/gin"
)

// @Summary 获取图形验证码
// @Description 获取登录用的图形验证码
// @Tags 认证
// @Produce json
// @Success 200 {object} map[string]string "验证码响应"
// @Failure 500 {object} utils.Response "生成验证码失败"
// @Router /api/merchant/captcha [get]
func GetCaptcha(c *gin.Context) {
	captchaService := utils.CaptchaService{}
	id, b64s, err := captchaService.GenerateCaptcha()
	if err != nil {
		utils.InternalError(c, "生成验证码失败")
		return
	}

	utils.Success(c, gin.H{
		"captcha_id":  id,
		"captcha_img": b64s,
	})
}
