// controllers/customer/merchant.go
package customer

import (
	"strconv"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

// 获取推荐商家列表
// @Summary 获取推荐商家列表
// @Description 获取系统推荐的商家列表（客户端简单版）
// @Tags 客户-商家
// @Produce json
// @Success 200 {array} models.Merchant "成功返回商家列表"
// @Failure 500 {object} utils.Response "获取商家列表失败"
// @Router /api/customer/merchants [get]
func GetRecommendedMerchants(c *gin.Context) {
	merchants, err := models.GetRecommendedMerchants()
	if err != nil {
		utils.InternalError(c, "获取商家列表失败")
		return
	}

	utils.Success(c, merchants)
}

// 获取商家详情
// @Summary 获取商家详情
// @Description 获取指定商家的详细信息（客户端）
// @Tags 客户-商家
// @Produce json
// @Param merchantId path int true "商家ID" example(123)
// @Success 200 {object} utils.Response "成功返回商家详情"
// @Failure 400 {object} utils.Response "无效的商家ID"
// @Failure 404 {object} utils.Response "商家不存在"
// @Router /api/customer/merchants/{merchantId} [get]
func GetMerchantDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("merchantId"))
	if err != nil {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	merchant, err := models.GetMerchantByID(uint(id))
	if err != nil {
		utils.NotFound(c, "商家不存在")
		return
	}

	utils.Success(c, merchant)
}

// 获取商家服务分类
// @Summary 获取商家服务分类
// @Description 获取指定商家的服务分类列表（客户端）
// @Tags 客户-商家
// @Produce json
// @Param merchantId path int true "商家ID" example(123)
// @Success 200 {array} utils.Response "成功返回分类列表"
// @Failure 400 {object} utils.Response "无效的商家ID"
// @Failure 404 {object} utils.Response "商家不存在"
// @Failure 500 {object} utils.Response "获取服务分类失败"
// @Router /api/customer/merchants/{merchantId}/categories [get]
func GetMerchantServiceCategories(c *gin.Context) {
	merchantID, err := strconv.Atoi(c.Param("merchantId"))
	if err != nil {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	categories, err := models.GetMerchantServiceCategories(uint(merchantID))
	if err != nil {
		utils.InternalError(c, "获取服务分类失败")
		return
	}

	utils.Success(c, categories)
}
