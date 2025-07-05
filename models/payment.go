package models

import (
	"admin-api/database"
	"time"
)

// 支付状态
const (
	PaymentStatusPending   = "pending"   // 待支付
	PaymentStatusSucceeded = "success"   // 支付成功
	PaymentStatusRefunding = "refunding" // 退款中
	PaymentStatusRefunded  = "refunded"  // 已退款
	PaymentStatusFailed    = "failed"    // 支付失败
	PaymentStatusClosed    = "closed"    // 已关闭
)

// 退款状态
const (
	RefundStatusProcessing = "processing" // 处理中
	RefundStatusSuccess    = "success"    // 退款成功
	RefundStatusFailed     = "failed"     // 退款失败
)

// Payment 支付记录模型
type Payment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	CustomerID    uint `gorm:"index" json:"customerId"`    // 用户ID
	MerchantID    uint `gorm:"index" json:"merchantId"`    // 商家ID
	AppointmentID uint `gorm:"index" json:"appointmentId"` // 关联预约ID

	OutTradeNo    string `gorm:"size:64;uniqueIndex" json:"outTradeNo"` // 商户订单号
	TransactionID string `gorm:"size:64" json:"transactionId"`          // 微信交易号

	Amount      int    `gorm:"index" json:"amount"`         // 支付金额(分)
	Description string `gorm:"size:255" json:"description"` // 支付描述

	Status     string     `gorm:"size:20" json:"status"`      // 支付状态
	PaidAt     *time.Time `json:"paidAt"`                     // 支付时间
	FailReason string     `gorm:"size:255" json:"failReason"` // 失败原因

	RawNotify string `gorm:"type:text" json:"-"` // 原始回调数据
}

// Refund 退款记录模型
type Refund struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	PaymentID     uint `gorm:"index" json:"paymentId"`     // 支付ID
	AppointmentID uint `gorm:"index" json:"appointmentId"` // 关联预约ID

	OutRefundNo string `gorm:"size:64;uniqueIndex" json:"outRefundNo"` // 商户退款单号
	RefundID    string `gorm:"size:64" json:"refundId"`                // 微信退款单号

	Amount int    `json:"amount"`                 // 退款金额(分)
	Reason string `gorm:"size:255" json:"reason"` // 退款原因

	Status     string     `gorm:"size:20" json:"status"`      // 退款状态
	RefundedAt *time.Time `json:"refundedAt"`                 // 退款时间
	FailReason string     `gorm:"size:255" json:"failReason"` // 失败原因
}

// CreatePayment 创建支付记录
func CreatePayment(payment *Payment) error {
	return database.DB.Create(payment).Error
}

// UpdatePayment 更新支付记录
func UpdatePayment(payment *Payment) error {
	return database.DB.Save(payment).Error
}

// GetPaymentByID 通过ID获取支付记录
func GetPaymentByID(id uint) (*Payment, error) {
	var payment Payment
	err := database.DB.First(&payment, id).Error
	return &payment, err
}

// GetPaymentByOutTradeNo 通过商户订单号获取支付记录
func GetPaymentByOutTradeNo(outTradeNo string) (*Payment, error) {
	var payment Payment
	err := database.DB.Where("out_trade_no = ?", outTradeNo).First(&payment).Error
	return &payment, err
}

// GetPaymentByAppointment 通过预约ID获取支付记录
func GetPaymentByAppointment(appointmentID uint) (*Payment, error) {
	var payment Payment
	err := database.DB.Where("appointment_id = ?", appointmentID).First(&payment).Error
	return &payment, err
}

// GetPaymentsByMerchant 获取商家支付记录
func GetPaymentsByMerchant(merchantID uint, status string, page, limit int) ([]Payment, int64, error) {
	var payments []Payment
	var total int64

	query := database.DB.Where("merchant_id = ?", merchantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Model(&Payment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&payments).Error

	return payments, total, err
}

// CreateRefund 创建退款记录
func CreateRefund(refund *Refund) error {
	return database.DB.Create(refund).Error
}

// UpdateRefund 更新退款记录
func UpdateRefund(refund *Refund) error {
	return database.DB.Save(refund).Error
}
