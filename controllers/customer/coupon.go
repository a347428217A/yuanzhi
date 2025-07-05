package customer

import (
	"fmt"
	"strconv"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

// 转换为响应格式
type CouponResponse struct {
	ID           uint   `json:"id"`
	CouponCode   string `json:"coupon_code"`
	Name         string `json:"name"`
	Discount     int    `json:"discount"`
	DiscountType string `json:"discount_type"`
	MinAmount    int    `json:"min_amount"`
	ValidFrom    string `json:"valid_from"`
	ValidTo      string `json:"valid_to"`
	Status       string `json:"status"`
}

// @Summary 获取用户优惠券列表
// @Description 获取当前登录用户的优惠券，支持按状态筛选
// @Tags 优惠券管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "优惠券状态筛选" Enums(unused, used, expired)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {array} CouponResponse "成功返回优惠券列表"
// @Failure 500 {object} utils.Response "获取优惠券失败"
// @Router /api/customer/coupons [get]
func GetUserCoupons(c *gin.Context) {
	userID := c.GetUint("user_id")
	status := c.Query("status") // unused, used, expired

	coupons, err := models.GetUserCoupons(userID, status)
	if err != nil {
		utils.InternalError(c, "获取优惠券失败")
		return
	}

	response := make([]CouponResponse, 0, len(coupons))
	for _, coupon := range coupons {
		response = append(response, CouponResponse{
			ID:           coupon.ID,
			CouponCode:   coupon.CouponCode,
			Name:         coupon.Template.Name,
			Discount:     coupon.Template.DiscountValue,
			DiscountType: coupon.Template.DiscountType,
			MinAmount:    coupon.Template.MinAmount,
			ValidFrom:    coupon.ValidFrom.Format("2006-01-02"),
			ValidTo:      coupon.ValidTo.Format("2006-01-02"),
			Status:       coupon.Status,
		})
	}

	utils.Success(c, response)
}

// @Summary 领取优惠券
// @Description 用户领取指定的优惠券模板
// @Tags 优惠券管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param couponTemplateId path int true "优惠券模板ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} models.User_coupon "成功返回领取的优惠券信息"
// @Failure 400 {object} utils.Response "无效的优惠券模板ID"
// @Failure 500 {object} utils.Response "领取优惠券失败"
// @Router /api/customer/coupons/{couponTemplateId}/claim [post]
func ClaimCoupon(c *gin.Context) {
	userID := c.GetUint("user_id")
	templateID, err := strconv.Atoi(c.Param("couponTemplateId"))
	if err != nil {
		utils.BadRequest(c, "无效的优惠券ID")
		return
	}

	coupon, err := models.ClaimUserCoupon(userID, uint(templateID))
	if err != nil {
		utils.InternalError(c, "领取优惠券失败: "+err.Error())
		return
	}

	utils.Success(c, coupon)
}

// 转换为响应格式
type CouponTemplateResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DiscountType string `json:"discount_type"`
	Discount     int    `json:"discount"`
	MinAmount    int    `json:"min_amount"`
	ValidityDays int    `json:"validity_days"`
}

// @Summary 获取可用优惠券
// @Description 获取指定商家的可用优惠券列表（客户端）
// @Tags 客户-优惠券
// @Produce json
// @Param merchantId query int true "商家ID" example(123)
// @Success 200 {array} CouponTemplateResponse "成功返回优惠券列表"
// @Failure 400 {object} utils.Response "无效的商家ID"
// @Failure 500 {object} utils.Response "获取优惠券失败"
// @Router /api/customer/coupons/available [get]
func GetAvailableCoupons(c *gin.Context) {
	merchantID, err := strconv.Atoi(c.Query("merchantId"))
	if err != nil {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	coupons, err := models.GetAvailableCoupons(uint(merchantID))
	if err != nil {
		fmt.Println(err)
		utils.InternalError(c, "获取优惠券失败")
		return
	}

	response := make([]CouponTemplateResponse, 0, len(coupons))
	for _, coupon := range coupons {
		response = append(response, CouponTemplateResponse{
			ID:           coupon.ID,
			Name:         coupon.Name,
			Description:  coupon.Description,
			DiscountType: coupon.DiscountType,
			Discount:     coupon.DiscountValue,
			MinAmount:    coupon.MinAmount,
			ValidityDays: coupon.ValidityDays,
		})
	}

	utils.Success(c, response)
}
