package merchant

import (
	"admin-api/database"
	"admin-api/models"
	"admin-api/utils"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary 获取优惠券模板列表
// @Description 获取当前商户的所有优惠券模板（商户端）
// @Tags 商户-优惠券管理
// @Security ApiKeyAuth
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param page query int false "页码" example(1)
// @Param page_size query int false "每页数量" example(20)
// @Param status query string false "状态" Enums(active,expired) example(active)
// @Success 200 {array} models.CouponTemplate "成功返回优惠券模板列表"
// @Failure 500 {object} utils.Response "获取失败"
// @Router /api/merchant/coupons [get]
func GetCouponTemplates(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	// 添加分页和过滤
	templates, total, err := models.GetCouponTemplates(merchantID, page, pageSize, status)
	if err != nil {
		utils.InternalError(c, "获取优惠券列表失败")
		return
	}

	// 返回分页结果
	utils.Success(c, gin.H{
		"list":      templates,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

//// @Summary 创建优惠券模板
//// @Description 创建新的优惠券模板（商户端）
//// @Tags 商户-优惠券管理
//// @Security ApiKeyAuth
//// @Accept json
//// @Produce json
//// @Param Authorization header string true "Bearer Token"
//// @Success 200 {object} utils.Response "成功返回创建的优惠券模板"
//// @Failure 400 {object} utils.Response "参数错误"
//// @Failure 500 {object} utils.Response "创建失败"
//// @Router /api/merchant/coupons [post]
//func GetCouponTemplates(c *gin.Context) {
//	merchantID := c.GetUint("merchant_id")
//
//	templates, err := models.GetCouponTemplates(merchantID)
//	if err != nil {
//		utils.InternalError(c, "获取优惠券列表失败")
//		return
//	}
//
//	utils.Success(c, templates)
//}

// @Summary 创建优惠券模板
// @Description 创建新的优惠券模板（商户端）
// @Tags 商户-优惠券管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body models.CouponTemplate true "优惠券模板信息"
// @Success 200 {object} models.CouponTemplate "成功返回创建的优惠券模板"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 500 {object} utils.Response "创建失败"
// @Router /api/merchant/coupons [post]
func CreateCouponTemplate(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")

	var req models.CouponTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证参数
	if req.DiscountType == "percent" && (req.DiscountValue <= 0 || req.DiscountValue > 100) {
		utils.BadRequest(c, "折扣百分比必须在1-100之间")
		return
	}

	if req.MinAmount < 0 {
		utils.BadRequest(c, "最低消费金额不能为负数")
		return
	}

	if req.ValidityDays <= 0 {
		utils.BadRequest(c, "有效期天数必须大于0")
		return
	}

	if req.TotalCount <= 0 {
		utils.BadRequest(c, "发行数量必须大于0")
		return
	}

	// 计算有效期
	//now := time.Now()
	template := models.CouponTemplate{
		MerchantID:    merchantID,
		Name:          req.Name,
		Description:   req.Description,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MinAmount:     req.MinAmount,
		ValidityDays:  req.ValidityDays,
		TotalCount:    req.TotalCount,
		//IsActive:      true,
	}

	// 使用事务创建模板和库存
	//tx := database.DB.Begin()
	//if err := tx.Create(&template).Error; err != nil {
	//	tx.Rollback()
	//	utils.InternalError(c, "创建优惠券失败: "+err.Error())
	//	return
	//}

	// 创建初始库存
	if err := models.CreateCouponTemplate(&template); err != nil {
		//tx.Rollback()
		utils.InternalError(c, "创建优惠券库存失败")
		return
	}

	//if err := tx.Commit().Error; err != nil {
	//	utils.InternalError(c, "提交事务失败")
	//	return
	//}

	utils.Success(c, template)
}

//func CreateCouponTemplate(c *gin.Context) {
//	merchantID := c.GetUint("merchant_id")
//
//	type Request struct {
//		Name         string  `json:"name" binding:"required"`
//		Description  string  `json:"description"`
//		DiscountType string  `json:"discount_type" binding:"required,oneof=fixed percent"`
//		Discount     float64 `json:"discount" binding:"required"`
//		MinAmount    float64 `json:"min_amount" binding:"required"`
//		ValidityDays int     `json:"validity_days" binding:"required"`
//		TotalCount   int     `json:"total_count" binding:"required"`
//	}
//
//	var req Request
//	if err := c.ShouldBindJSON(&req); err != nil {
//		utils.BadRequest(c, "参数错误")
//		return
//	}
//
//	template := models.CouponTemplate{
//		MerchantID:   merchantID,
//		Name:         req.Name,
//		Description:  req.Description,
//		DiscountType: req.DiscountType,
//		Discount:     req.Discount,
//		MinAmount:    req.MinAmount,
//		ValidityDays: req.ValidityDays,
//		TotalCount:   req.TotalCount,
//		IsActive:     true,
//	}
//
//	if err := models.CreateCouponTemplate(&template); err != nil {
//		utils.InternalError(c, "创建优惠券失败")
//		return
//	}
//
//	utils.Success(c, template)
//}

// @Summary 更新优惠券模板
// @Description 更新指定的优惠券模板（商户端）
// @Tags 商户-优惠券管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param couponTemplateId path int true "优惠券模板ID" example(123)
// @Param request body models.CouponTemplate true "优惠券模板更新信息"
// @Success 200 {object} utils.Response "更新成功"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 404 {object} utils.Response "优惠券不存在"
// @Failure 500 {object} utils.Response "更新失败"
// @Router /api/merchant/coupons/{couponTemplateId} [put]
func UpdateCouponTemplate(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	templateID, err := strconv.Atoi(c.Param("couponTemplateId"))
	if err != nil {
		utils.BadRequest(c, "无效的优惠券ID")
		return
	}

	//type Request struct {
	//	Name         string  `json:"name"`
	//	Description  string  `json:"description"`
	//	DiscountType string  `json:"discount_type" oneof=fixed percent"`
	//	Discount     float64 `json:"discount"`
	//	MinAmount    float64 `json:"min_amount"`
	//	ValidityDays int     `json:"validity_days"`
	//	TotalCount   int     `json:"total_count"`
	//	IsActive     bool    `json:"is_active"`
	//}

	var req models.CouponTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// 验证优惠券属于该商家
	var template models.CouponTemplate
	if err := database.DB.First(&template, templateID).Error; err != nil || template.MerchantID != merchantID {
		utils.NotFound(c, "优惠券不存在")
		return
	}

	updates := map[string]interface{}{
		"name":           req.Name,
		"description":    req.Description,
		"discount_type":  req.DiscountType,
		"discount_value": req.DiscountValue,
		"min_amount":     req.MinAmount,
		"validity_days":  req.ValidityDays,
		"total_count":    req.TotalCount,
	}

	if err := models.UpdateCouponTemplate(uint(templateID), updates); err != nil {
		utils.InternalError(c, "更新优惠券失败")
		fmt.Println(err)
		return
	}

	utils.Success(c, "更新成功")
}

// @Summary 删除优惠券模板
// @Description 删除指定的优惠券模板（商户端）。注意：只能删除未被使用的优惠券模板。
// @Tags 商户-优惠券管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param couponTemplateId path int true "优惠券模板ID" example(123)
// @Success 200 {object} utils.Response "删除成功"
// @Failure 400 {object} utils.Response "无效的优惠券ID或优惠券已被使用"
// @Failure 404 {object} utils.Response "优惠券不存在"
// @Failure 500 {object} utils.Response "删除失败"
// @Router /api/merchant/coupons/{couponTemplateId} [delete]
func DeleteCouponTemplate(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	templateID, err := strconv.Atoi(c.Param("couponTemplateId"))
	if err != nil {
		utils.BadRequest(c, "无效的优惠券ID")
		return
	}

	// 检查是否有已发放的优惠券
	var couponCount int64
	if err := database.DB.Model(&models.CouponTemplate{}).
		Where("id = ?", templateID).
		Count(&couponCount).Error; err != nil {
		utils.InternalError(c, "检查优惠券使用情况失败")
		return
	}

	if couponCount > 0 {
		utils.BadRequest(c, "优惠券已被使用，无法删除")
		return
	}

	// 验证优惠券属于该商家
	var template models.CouponTemplate
	if err := database.DB.First(&template, templateID).Error; err != nil || template.MerchantID != merchantID {
		utils.NotFound(c, "优惠券不存在")
		return
	}

	if err := models.DeleteCouponTemplate(uint(templateID)); err != nil {
		utils.InternalError(c, "删除优惠券失败")
		return
	}

	utils.Success(c, "优惠券删除成功")
}
