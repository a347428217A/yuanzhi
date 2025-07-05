// controllers/customer/auth.go
package customer

import (
	"admin-api/models"
	"admin-api/pkg/auth"
	"admin-api/pkg/wechat"
	"admin-api/utils"
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	AppID     = "wx10ca8858028379ec"               // 替换为你的APPID
	AppSecret = "232fc9c655456253aed21efb6b230df3" // 替换为你的APPSECRET
)

// WechatLoginRequest 微信登录请求结构
type WechatLoginRequest struct {
	Code string `json:"code" binding:"required" example:"081klykl2NkUx64YqLml2NkUx6klykly"` // 微信授权码
}

// 微信登录
// @Summary 微信登录
// @Description 使用微信授权码登录或注册用户
// @Tags 客户-认证
// @Accept json
// @Produce json
// @Param request body WechatLoginRequest true "微信登录请求"
// @Success 200 {object} utils.Response "登录成功"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 401 {object} utils.Response "微信登录失败"
// @Failure 500 {object} utils.Response "用户处理失败或生成token失败"
// @Router /api/customer/login [post]
func WechatLogin(c *gin.Context) {
	type Request struct {
		Code string `json:"code" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// 获取微信openid
	openid, err := wechat.GetWechatSession(req.Code, AppID, AppSecret)
	if err != nil {
		utils.Error(c, 401, "微信登录失败")
		fmt.Println(err)
		return
	}

	// 查找或创建用户
	user, err := models.FindOrCreateUserByOpenID(openid.OpenID)
	if err != nil {
		utils.InternalError(c, "用户处理失败")
		fmt.Println(err)
		return
	}

	// 生成JWT
	token, err := auth.GenerateCustomerToken(user.ID)
	if err != nil {
		utils.InternalError(c, "生成token失败")
		return
	}

	utils.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"phone":    user.Phone,
		},
	})
}

type UpdatePhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// 更新用户手机号
// @Summary 更新用户手机号
// @Description 用户绑定新的手机号（需要短信验证码）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer Token"
// @Param body body UpdatePhoneRequest true "手机号更新请求"
// @Success 200 {object} string "手机号更新成功"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 500 {object} utils.Response "更新手机号失败"
// @Router /api/customer/phone [put]
func UpdatePhone(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdatePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	//// 验证短信验证码
	//if !utils.VerifySMSCode(req.Phone, req.Code) {
	//	utils.BadRequest(c, "验证码错误")
	//	return
	//}

	// 更新手机号
	if err := models.UpdateUserPhone(userID, req.Phone); err != nil {
		utils.InternalError(c, "更新手机号失败")
		return
	}

	utils.Success(c, "手机号更新成功")
}
