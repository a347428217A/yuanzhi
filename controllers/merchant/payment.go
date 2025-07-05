package merchant

import (
	"admin-api/config"
	"admin-api/database"
	"admin-api/models"
	"admin-api/payment"
	"admin-api/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

// GetMerchantPayments 获取商家支付记录
// @Summary 获取商家支付记录
// @Description 查询商家的支付记录
// @Tags 商家支付
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param status query string false "支付状态" Enums(pending, success, refunded, failed)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} utils.PaginatedResponse{data=[]models.Payment} "支付记录列表"
// @Failure 400 {object} utils.Response "参数错误"
// @Router /api/merchant/payments [get]
func GetMerchantPayments(c *gin.Context) {
	// 获取当前商家
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 参数验证
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	payments, total, err := models.GetPaymentsByMerchant(
		merchantID.(uint),
		status,
		page,
		limit,
	)

	if err != nil {
		utils.InternalError(c, "获取支付记录失败: "+err.Error())
		return
	}

	utils.PaginatedSuccess(c, payments, total, page, limit)
}

// GetPaymentDetail 获取支付详情
// @Summary 获取支付详情
// @Description 查询支付订单详细信息
// @Tags 商家支付
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param paymentId path int true "支付订单ID"
// @Success 200 {object} models.Payment "支付订单详情"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 404 {object} utils.Response "支付订单不存在"
// @Router /api/merchant/payments/{paymentId} [get]
func GetPaymentDetail(c *gin.Context) {
	paymentID, err := strconv.Atoi(c.Param("paymentId"))
	if err != nil || paymentID <= 0 {
		utils.BadRequest(c, "无效的支付ID")
		return
	}

	// 获取当前商家
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	payment_, err := models.GetPaymentByID(uint(paymentID))
	if err != nil {
		utils.NotFound(c, "支付订单不存在")
		return
	}

	// 验证支付关联的预约是否属于当前商家
	if payment_.AppointmentID != 0 {
		appointment, err := models.GetAppointmentByID(payment_.AppointmentID)
		if err != nil || appointment.MerchantID != merchantID.(uint) {
			utils.Forbidden(c, "无权查看此订单")
			return
		}
	} else {
		// 如果没有关联预约，则直接拒绝
		utils.Forbidden(c, "无权查看此订单")
		return
	}

	utils.Success(c, payment_)
}

// RefundRequest 退款请求
type RefundRequest struct {
	RefundAmount int    `json:"refundAmount" binding:"required,min=1"` // 退款金额(分)
	RefundReason string `json:"refundReason"`                          // 退款原因
}

// InitiateRefund 发起退款
// @Summary 发起退款
// @Description 为已支付的预约发起退款
// @Tags 商家支付
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param appointmentId path int true "预约ID"
// @Param body body RefundRequest true "退款请求"
// @Success 200 {object} models.Refund "退款记录"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 500 {object} utils.Response "退款失败"
// @Router /api/merchant/appointments/{appointmentId}/refund [post]
func InitiateRefund(c *gin.Context) {
	appointmentID, err := strconv.Atoi(c.Param("appointmentId"))
	if err != nil || appointmentID <= 0 {
		utils.BadRequest(c, "无效的预约ID")
		return
	}

	// 获取当前商家
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	//merchantID, _ := c.Get("merchantId")

	// 获取预约信息
	appointment, err := models.GetAppointmentByID(uint(appointmentID))
	if err != nil {
		utils.NotFound(c, "预约不存在")
		return
	}

	// 验证预约属于当前商家
	if appointment.MerchantID != merchantID.(uint) {
		utils.Forbidden(c, "无权操作此预约")
		return
	}

	// 检查预约状态是否可退款
	if appointment.Status != models.AppointmentStatusCompleted {

		utils.BadRequest(c, "当前状态不允许退款")
		return
	}

	// 获取关联的支付记录
	payment_, err := models.GetPaymentByAppointment(appointment.ID)
	if err != nil {
		utils.BadRequest(c, "未找到支付记录")
		return
	}

	// 检查支付状态
	if payment_.Status != models.PaymentStatusSucceeded {
		fmt.Println(payment_.Status)
		utils.BadRequest(c, "支付未完成，无法退款")
		return
	}

	// 解析退款请求
	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证退款金额
	if req.RefundAmount > payment_.Amount {
		utils.BadRequest(c, "退款金额不能超过支付金额")
		return
	}

	// 创建退款记录

	if config.Config.WechatPay.UseSimulate {
		// 创建模拟退款记录
		refundRecord := models.Refund{
			PaymentID:     payment_.ID,
			AppointmentID: appointment.ID,
			Amount:        req.RefundAmount,
			Reason:        req.RefundReason,
			Status:        models.RefundStatusSuccess,    // 直接标记为成功
			OutRefundNo:   utils.GenerateTradeNo("SIMR"), // 添加SIMR前缀
		}

		if err := models.CreateRefund(&refundRecord); err != nil {
			utils.InternalError(c, "创建模拟退款记录失败: "+err.Error())
			return
		}

		// 更新支付状态
		payment_.Status = models.PaymentStatusRefunded
		if err := models.UpdatePayment(payment_); err != nil {
			utils.InternalError(c, "更新支付状态失败: "+err.Error())
			return
		}

		// 更新预约状态
		appointment.Status = models.AppointmentStatusCancelled
		if err := database.DB.Save(&appointment).Error; err != nil {
			utils.InternalError(c, "更新预约状态失败: "+err.Error())
			return
		}

		utils.Success(c, refundRecord)
		return
	}

	refundRecord := models.Refund{
		PaymentID:     payment_.ID,
		AppointmentID: appointment.ID,
		Amount:        req.RefundAmount,
		Reason:        req.RefundReason,
		Status:        models.RefundStatusProcessing,
		OutRefundNo:   utils.GenerateTradeNo("R"),
	}

	if err := models.CreateRefund(&refundRecord); err != nil {
		utils.InternalError(c, "创建退款记录失败: "+err.Error())
		return
	}

	// 调用微信退款API
	err = payment.CreateWechatRefund(
		payment_.OutTradeNo,
		refundRecord.OutRefundNo,
		payment_.Amount,
		refundRecord.Amount,
		refundRecord.Reason,
	)

	if err != nil {
		// 更新退款状态为失败
		refundRecord.Status = models.RefundStatusFailed
		refundRecord.FailReason = err.Error()
		models.UpdateRefund(&refundRecord)

		utils.InternalError(c, "发起退款失败: "+err.Error())
		return
	}

	// 更新支付状态为退款中
	payment_.Status = models.PaymentStatusRefunding
	models.UpdatePayment(payment_)

	utils.Success(c, refundRecord)
}
