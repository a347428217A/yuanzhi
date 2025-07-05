// controllers/customer/user.go
package customer

import (
	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

// 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的个人信息（包括统计信息）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} utils.Response "成功返回用户信息"
// @Failure 500 {object} utils.Response "获取用户信息失败"
// @Router /api/customer/profile [get]
func GetUserProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	user, err := models.GetUserByID(userID)
	if err != nil {
		utils.InternalError(c, "获取用户信息失败")
		return
	}

	// 获取用户预约统计
	stats, err := models.GetUserAppointmentStats(userID)
	if err != nil {
		stats = &models.AppointmentStats{}
	}

	response := gin.H{
		"id":       user.ID,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"phone":    user.Phone,
		"points":   user.Points,
		"stats": gin.H{
			"total":     stats.Total,
			"completed": stats.Completed,
			"upcoming":  stats.Upcoming,
		},
	}

	utils.Success(c, response)
}
