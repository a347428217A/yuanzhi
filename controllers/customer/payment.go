package customer

import (
	"admin-api/config"
	"admin-api/database"
	"admin-api/models"
	"admin-api/payment"
	"admin-api/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// PaymentRequest 支付请求结构体
type PaymentRequest struct {
	Amount        int    `json:"amount" binding:"required,min=1"` // 支付金额(分)
	Description   string `json:"description" binding:"required"`  // 支付描述
	AppointmentID uint   `json:"appointmentId"`                   // 关联预约ID(可选)
}

// CreatePayment 创建支付订单
// @Summary 创建支付订单
// @Description 用户创建支付订单
// @Tags 客户支付
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer Token"
// @Param body body PaymentRequest true "支付请求"
// @Success 200 {object} payment.PrepayResponse "支付预订单信息"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 500 {object} utils.Response "创建支付失败"
// @Router /api/customer/payments [post]
func CreatePayment(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := c.GetUint("user_id")

	// 获取当前用户
	//userID, _ := c.Get("customerId")
	customer, err := models.GetUserByID(userID)
	//customer, err := models.GetCustomerByID(userID.(uint))
	if err != nil {
		utils.Unauthorized(c, "用户信息错误")
		return
	}

	// 创建本地支付记录
	paymentRecord := models.Payment{
		CustomerID:    customer.ID,
		AppointmentID: req.AppointmentID,
		Amount:        req.Amount,
		Description:   req.Description,
		Status:        models.PaymentStatusPending,
		OutTradeNo:    utils.GenerateTradeNo("P"),
	}

	if err := models.CreatePayment(&paymentRecord); err != nil {

		utils.InternalError(c, "创建支付记录失败: "+err.Error())
		return
	}

	// 调用微信支付
	prepayResp, err := payment.CreateWechatPayOrder(payment.OutTradeNo(paymentRecord.OutTradeNo),
		payment.Amount(req.Amount),
		payment.Description(req.Description),
		payment.OpenID(customer.Openid),
	)

	if err != nil {
		// 更新支付状态为失败
		fmt.Println(err)
		paymentRecord.Status = models.PaymentStatusFailed
		paymentRecord.FailReason = err.Error()
		models.UpdatePayment(&paymentRecord)

		utils.InternalError(c, "创建微信支付失败: "+err.Error())
		return
	}

	utils.Success(c, prepayResp)
}

// PayForAppointment 为预约支付
// @Summary 为预约支付
// @Description 用户为预约创建支付订单
// @Tags 客户支付
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param appointmentId path int true "预约ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} payment.PrepayResponse "支付预订单信息"
// @Failure 400 {object} utils.Response "预约信息错误"
// @Failure 500 {object} utils.Response "创建支付失败"
// @Router /api/customer/appointments/{appointmentId}/pay [post]
func PayForAppointment(c *gin.Context) {
	//if config.Config.

	appointmentID, err := strconv.Atoi(c.Param("appointmentId"))
	if err != nil || appointmentID <= 0 {
		utils.BadRequest(c, "无效的预约ID")
		return
	}

	// 获取当前用户
	//userID, _ := c.Get("customerId")
	userID := c.GetUint("user_id")

	customer, err := models.GetUserByID(userID)
	//customer, err := models.GetCustomerByID(userID.(uint))
	if err != nil {
		utils.Unauthorized(c, "用户信息错误")
		return
	}

	// 获取预约信息
	appointment, err := models.GetAppointmentByID(uint(appointmentID))
	if err != nil {
		utils.NotFound(c, "预约不存在")
		return
	}

	// 验证预约属于当前用户
	if appointment.UserID != customer.ID {
		utils.Forbidden(c, "无权操作此预约")
		return
	}

	// 检查预约状态是否可支付

	if appointment.Status != "confirmed" {
		utils.BadRequest(c, "预约状态不允许支付")
		return
	}

	//模拟支付
	if config.Config.WechatPay.UseSimulate {
		// 创建模拟支付记录
		paymentRecord := models.Payment{
			CustomerID:    customer.ID,
			MerchantID:    appointment.MerchantID,
			AppointmentID: appointment.ID,
			Amount:        appointment.Amount,
			Description:   fmt.Sprintf("模拟支付-%s", appointment.OrderNo),
			Status:        models.PaymentStatusPending,
			OutTradeNo:    utils.GenerateTradeNo("SIM"), // 添加SIM前缀标识模拟支付
		}

		if err := models.CreatePayment(&paymentRecord); err != nil {
			utils.InternalError(c, "创建模拟支付记录失败: "+err.Error())
			return
		}

		// 返回模拟支付信息
		utils.Success(c, gin.H{
			"simulate":   true,
			"paymentId":  paymentRecord.ID,
			"outTradeNo": paymentRecord.OutTradeNo,
			//"simulateUrl": fmt.Sprintf("%s/pay/simulate?paymentId=%d&outTradeNo=%s",
			//	config.Config.App.BaseURL, paymentRecord.ID, paymentRecord.OutTradeNo),
		})
		return
	}

	// 创建支付请求
	req := PaymentRequest{
		Amount:        appointment.Amount, // 单位:分
		Description:   fmt.Sprintf("预约支付-%s", appointment.OrderNo),
		AppointmentID: appointment.ID,
	}

	paymentRecord := models.Payment{
		CustomerID:    customer.ID,
		AppointmentID: req.AppointmentID,
		Amount:        req.Amount,
		Description:   req.Description,
		Status:        models.PaymentStatusPending,
		OutTradeNo:    utils.GenerateTradeNo("P"),
	}

	if err := models.CreatePayment(&paymentRecord); err != nil {
		utils.InternalError(c, "创建支付记录失败: "+err.Error())
		return
	}

	// 调用微信支付
	prepayResp, err := payment.CreateWechatPayOrder(payment.OutTradeNo(paymentRecord.OutTradeNo),
		payment.Amount(req.Amount),
		payment.Description(req.Description),
		payment.OpenID(customer.Openid),
	)

	if err != nil {
		// 更新支付状态为失败
		paymentRecord.Status = models.PaymentStatusFailed
		paymentRecord.FailReason = err.Error()
		models.UpdatePayment(&paymentRecord)

		utils.InternalError(c, "创建微信支付失败: "+err.Error())
		return
	}

	utils.Success(c, prepayResp)

	// 使用CreatePayment逻辑
	//c.Set("paymentRequest", req)
	//CreatePayment(c)
}

// GetPayment 获取支付状态
// @Summary 获取支付状态
// @Description 查询支付订单状态
// @Tags 客户支付
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paymentId path int true "支付订单ID"
// @Success 200 {object} models.Payment "支付订单详情"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 404 {object} utils.Response "支付订单不存在"
// @Router /api/customer/payments/{paymentId} [get]
func GetPayment(c *gin.Context) {
	paymentID, err := strconv.Atoi(c.Param("paymentId"))
	if err != nil || paymentID <= 0 {
		utils.BadRequest(c, "无效的支付ID")
		return
	}

	// 获取当前用户
	userID, _ := c.Get("customerId")

	payment_, err := models.GetPaymentByID(uint(paymentID))
	if err != nil {
		utils.NotFound(c, "支付订单不存在")
		return
	}

	// 验证订单属于当前用户
	if payment_.CustomerID != userID.(uint) {
		utils.Forbidden(c, "无权查看此订单")
		return
	}

	utils.Success(c, payment_)
}

// HandlePaymentNotify 支付回调通知
// @Summary 微信支付回调
// @Description 微信支付结果通知
// @Tags 支付回调
// @Accept xml
// @Produce xml
// @Param xml body payment.WechatNotifyRequest true "回调数据"
// @Success 200 {object} payment.WechatNotifyResponse "处理结果"
// @Router /api/customer/payments/notify [post]
func HandlePaymentNotify(c *gin.Context) {
	var notifyReq payment.WechatNotifyRequest

	// 解析XML请求
	if err := c.ShouldBindXML(&notifyReq); err != nil {
		c.XML(http.StatusBadRequest, payment.WechatNotifyResponse{
			ReturnCode: "FAIL",
			ReturnMsg:  "解析XML失败",
		})
		return
	}

	// 验证签名
	if valid := payment.VerifyWechatSign(notifyReq, config.Config.WechatPay.APIKey); !valid {
		c.XML(http.StatusBadRequest, payment.WechatNotifyResponse{
			ReturnCode: "FAIL",
			ReturnMsg:  "签名验证失败",
		})
		return
	}

	// 处理支付结果
	err := payment.HandlePaymentResult(notifyReq)
	if err != nil {
		// 记录错误日志
		//utils.LogError("支付回调处理失败", "out_trade_no", notifyReq.OutTradeNo, "error", err.Error())

		c.XML(http.StatusOK, payment.WechatNotifyResponse{
			ReturnCode: "FAIL",
			ReturnMsg:  "处理失败",
		})
		return
	}

	// 返回成功响应
	c.XML(http.StatusOK, payment.WechatNotifyResponse{
		ReturnCode: "SUCCESS",
		ReturnMsg:  "OK",
	})
}

// SimulatePaymentNotifyRequest 模拟支付回调请求
type SimulatePaymentNotifyRequest struct {
	PaymentID  uint   `json:"paymentId" binding:"required"`
	OutTradeNo string `json:"outTradeNo" binding:"required"`
	Status     string `json:"status" binding:"required,oneof=success failed"`
}

// HandleSimulatePaymentNotify 模拟支付回调
// @Summary 模拟支付回调
// @Description 用于开发环境的模拟支付回调
// @Tags 支付回调
// @Accept json
// @Produce json
// @Param body body SimulatePaymentNotifyRequest true "回调数据"
// @Success 200 {object} utils.Response
// @Router /api/customer/payments/simulate-notify [post]
func HandleSimulatePaymentNotify(c *gin.Context) {
	var req SimulatePaymentNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取支付记录
	payment, err := models.GetPaymentByID(req.PaymentID)
	if err != nil {
		utils.NotFound(c, "支付记录不存在")
		return
	}

	// 验证订单号
	if payment.OutTradeNo != req.OutTradeNo {
		utils.BadRequest(c, "订单号不匹配")
		return
	}

	// 验证是模拟支付
	if !strings.HasPrefix(payment.OutTradeNo, "SIM") {
		utils.Forbidden(c, "仅支持模拟支付订单")
		return
	}

	// 使用事务确保数据一致性
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新支付状态
	updateData := map[string]interface{}{
		"status": req.Status,
	}

	if req.Status == models.PaymentStatusSucceeded {
		now := time.Now()
		updateData["paid_at"] = &now
	}

	if err := tx.Model(&models.Payment{}).
		Where("id = ?", payment.ID).
		Updates(updateData).Error; err != nil {
		tx.Rollback()
		utils.InternalError(c, "更新支付状态失败: "+err.Error())
		return
	}

	// 只有在支付成功时才更新预约状态
	if req.Status == models.PaymentStatusSucceeded {
		// 获取关联的预约
		appointment, err := models.GetAppointmentByID(payment.AppointmentID)
		if err != nil {
			tx.Rollback()
			utils.InternalError(c, "获取预约信息失败: "+err.Error())
			return
		}

		// 更新预约状态为已确认
		appointment.Status = models.AppointmentStatusCompleted
		appointment.PaymentID = payment.ID
		//appointment.PaymentStatus = models.PaymentStatusSucceeded

		if err := tx.Save(&appointment).Error; err != nil {
			tx.Rollback()
			utils.InternalError(c, "更新预约状态失败: "+err.Error())
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.InternalError(c, "提交事务失败: "+err.Error())
		return
	}

	utils.Success(c, "支付状态更新成功")
}
