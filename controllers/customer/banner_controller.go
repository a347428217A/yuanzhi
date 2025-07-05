package customer

import (
	"admin-api/models"
	"admin-api/utils"
	"github.com/gin-gonic/gin"
	"time"
)

// @Summary 获取轮播图
// @Description 用户端：获取有效轮播图
// @Tags 轮播图
// @Accept json
// @Produce json
// @Param position query string true "展示位置(home,category)"
// @Param platform query string true "平台(weapp,h5,app)"
// @Success 200 {array} models.Banner
// @Router /api/customer/banners [get]
func GetBanners(c *gin.Context) {
	position := c.Query("position")
	platform := c.Query("platform")

	if position == "" || platform == "" {
		utils.BadRequest(c, "位置和平台参数不能为空")
		return
	}

	// 获取当前时间
	now := time.Now()

	// 获取有效的轮播图
	banners, err := models.GetActiveBanners(position, platform, now)
	if err != nil {
		utils.InternalError(c, "获取轮播图失败: "+err.Error())
		return
	}

	utils.Success(c, banners)
}
