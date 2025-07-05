package merchant

import (
	"admin-api/models"
	"admin-api/pkg/auth"
	"admin-api/utils"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

//type LoginRequest struct {
//	Username    string `json:"username" binding:"required"`
//	Password    string `json:"password" binding:"required"`
//	CaptchaID   string `json:"captcha_id" binding:"required"`
//	CaptchaCode string `json:"captcha_code" binding:"required"`
//}

// @Summary 商户管理员登录
// @Description 使用用户名和密码登录，获取JWT令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param input body LoginRequest true "登录凭证"
// @Success 200 {object} map[string]interface{} "登录成功" Example({"token": "jwt_token", "admin": {"id":"1", "username":"admin", "role":"super_admin", "merchant_id":"123"}})
// @Failure 400 {object} map[string]string "参数错误" Example({"error": "参数错误"})
// @Failure 401 {object} map[string]string "用户名或密码错误" Example({"error": "用户名或密码错误"})
// @Failure 500 {object} map[string]string "生成token失败" Example({"error": "生成token失败"})
// @Router /api/merchant/login [post] // 修改这一行
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	//1. 验证验证码
	//captchaService := utils.CaptchaService{}
	//if !captchaService.Verify(req.CaptchaID, req.CaptchaCode) {
	//	utils.Error(c, 403, "验证码错误或已过期")
	//	return
	//}

	admin, err := models.GetMerchantAdminByUsername(req.Username)
	if err != nil {
		utils.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 验证密码
	match, err := utils.VerifyPassword(req.Password, admin.Password)
	if err != nil {
		utils.InternalError(c, "密码验证失败: "+err.Error())
		return
	}
	if !match {
		utils.Unauthorized(c, "用户名或密码错误")
		return
	}

	//if !auth.CheckPassword(admin.Password, req.Password) {
	//	utils.Unauthorized(c, "用户名或密码错误")
	//	return
	//}

	// 生成JWT
	token, err := auth.GenerateMerchantToken(admin.ID, admin.MerchantID, admin.Role, "default_secret")
	if err != nil {
		utils.InternalError(c, "生成token失败")
		return
	}

	utils.Success(c, gin.H{
		"token": token,
		"admin": gin.H{
			"id":          admin.ID,
			"username":    admin.Username,
			"role":        admin.Role,
			"merchant_id": admin.MerchantID,
		},
	})
}

//func GetAppointmentStats(c *gin.Context) {
//	merchantID := c.GetUint("merchant_id")
//
//	// 获取今日预约数
//	var todayCount int64
//	today := time.Now().Format("2006-01-02")
//	if err := models.DB.Model(&models.Appointment{}).
//		Where("merchant_id = ? AND appointment_date = ?", merchantID, today).
//		Count(&todayCount).Error; err != nil {
//		utils.InternalError(c, "获取数据失败")
//		return
//	}
//
//	// 获取本周预约数
//	var weekCount int64
//	startOfWeek := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+1).Format("2006-01-02")
//	endOfWeek := time.Now().AddDate(0, 0, 7-int(time.Now().Weekday())).Format("2006-01-02")
//	if err := models.DB.Model(&models.Appointment{}).
//		Where("merchant_id = ? AND appointment_date BETWEEN ? AND ?",
//			merchantID, startOfWeek, endOfWeek).
//		Count(&weekCount).Error; err != nil {
//		utils.InternalError(c, "获取数据失败")
//		return
//	}
//
//	// 获取本月预约数
//	var monthCount int64
//	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")
//	endOfMonth := time.Now().AddDate(0, 1, -time.Now().Day()).Format("2006-01-02")
//	if err := models.DB.Model(&models.Appointment{}).
//		Where("merchant_id = ? AND appointment_date BETWEEN ? AND ?",
//			merchantID, startOfMonth, endOfMonth).
//		Count(&monthCount).Error; err != nil {
//		utils.InternalError(c, "获取数据失败")
//		return
//	}
//
//	// 获取不同状态预约数
//	statusStats := make(map[string]int64)
//	statuses := []string{"pending", "confirmed", "completed", "canceled", "rejected"}
//
//	for _, status := range statuses {
//		var count int64
//		if err := models.DB.Model(&models.Appointment{}).
//			Where("merchant_id = ? AND status = ?", merchantID, status).
//			Count(&count).Error; err == nil {
//			statusStats[status] = count
//		}
//	}
//
//	utils.Success(c, gin.H{
//		"today":  todayCount,
//		"week":   weekCount,
//		"month":  monthCount,
//		"status": statusStats,
//	})
//}
