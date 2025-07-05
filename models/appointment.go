package models

import (
	"admin-api/database"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"time"
	//"gorm.io/gorm"
)

type Appointment struct {
	ID              uint      `gorm:"primaryKey"`
	OrderNo         string    `gorm:"size:32;uniqueIndex;not null"`
	UserID          uint      `gorm:"index;not null"`
	MerchantID      uint      `gorm:"index;not null"`
	ServiceID       uint      `gorm:"index;not null"`
	StaffID         uint      `gorm:"index;not null"`
	TimeSlotID      uint      `gorm:"index;not null"`
	AppointmentDate time.Time `gorm:"type:date;not null"`
	StartTime       string    `gorm:"type:time;not null"`
	EndTime         string    `gorm:"type:time;not null"`
	Status          string    `gorm:"size:20;default:'pending';not null"`
	Amount          int       `gorm:"type:int;default:0;not null"` // 改为分单位的整数
	PaymentID       uint      `gorm:"index"`                       // 新增支付ID关联
	Remark          string    `gorm:"size:255"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	User     User        `gorm:"foreignKey:UserID"`
	Merchant Merchant    `gorm:"foreignKey:MerchantID"`
	Service  Service     `gorm:"foreignKey:ServiceID"`
	Staff    Staff       `gorm:"foreignKey:StaffID"`
	Coupon   *UserCoupon `gorm:"foreignKey:AppointmentID"`

	TimeSlot TimeSlot `gorm:"foreignKey:TimeSlotID;references:ID"`
}

const (
	AppointmentStatusPending   = "pending"
	AppointmentStatusConfirmed = "confirmed"
	AppointmentStatusPaid      = "paid" // 新增已支付状态
	AppointmentStatusCompleted = "completed"
	AppointmentStatusCancelled = "cancelled"
)

func CreateCustomerAppointment(userID, merchantID, serviceID, staffID, timeSlotID uint,
	date time.Time, couponID uint, remark string) (*Appointment, error) {

	// 开始事务
	tx := database.DB.Begin()

	// 1. 获取时间段信息并锁定
	var timeSlot TimeSlot
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&timeSlot, timeSlotID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("时间段不存在")
	}

	// 检查时间段是否可用
	if !timeSlot.IsAvailable {
		tx.Rollback()
		return nil, fmt.Errorf("该时间段已被预约")
	}

	// 2. 获取服务信息
	var service Service
	if err := tx.First(&service, serviceID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("服务不存在")
	}

	// 3. 计算最终价格（考虑优惠券）
	finalAmount := service.Price
	var coupon *UserCoupon
	//var coupon *CouponApplication
	if couponID > 0 {
		//app, err := ApplyCoupon(tx, userID, couponID, service.Price)
		//if err != nil {
		//	tx.Rollback()
		//	utils.BadRequest(c, err.Error())
		//	return nil, err
		//}
		//finalAmount = app.FinalPrice
		//couponApplication := app
		//usedCouponID := &app.UserCoupon.ID

		c, err := ApplyCoupon(tx, userID, couponID, service.Price)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("优惠券不可用: %v", err)
		}
		finalAmount = c.FinalPrice
		coupon = c.UserCoupon
	}

	// 4. 创建预约记录
	appointment := &Appointment{
		OrderNo:         generateOrderNo(),
		UserID:          userID,
		MerchantID:      merchantID,
		ServiceID:       serviceID,
		StaffID:         staffID,
		TimeSlotID:      timeSlotID,
		AppointmentDate: date,
		StartTime:       timeSlot.StartTime,
		EndTime:         timeSlot.EndTime,
		Status:          "pending", // 待确认状态
		Amount:          int(finalAmount),
		Remark:          remark,
	}

	if err := tx.Create(appointment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建预约失败")
	}

	// 5. 标记时间段为不可用
	if err := tx.Model(&TimeSlot{}).Where("id = ?", timeSlotID).Update("is_available", false).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新时间段状态失败")
	}

	// 6. 如果使用了优惠券，标记为已使用
	if coupon != nil {
		coupon.AppointmentID = &appointment.ID
		coupon.Status = "used"
		coupon.UsedAt = time.Now()
		if err := tx.Save(coupon).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新优惠券状态失败")
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败")
	}

	// TODO: 发送通知给商家

	return appointment, nil
}

// 生成订单号
func generateOrderNo() string {
	return fmt.Sprintf("ORD%d%06d", time.Now().Unix(), rand.Intn(1000000))
}

func GetUserAppointments(userID uint, status string) ([]Appointment, error) {
	var appointments []Appointment

	query := database.DB.Preload("Merchant").Preload("Service").Preload("Staff").
		Where("user_id = ?", userID).
		Order("appointment_date DESC, start_time DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&appointments).Error
	return appointments, err
}

func GetUserAppointmentDetail(userID, appointmentID uint) (*Appointment, error) {
	var appointment Appointment
	err := database.DB.Preload("Merchant").Preload("Service").Preload("Staff").
		Where("id = ? AND user_id = ?", appointmentID, userID).
		First(&appointment).Error
	return &appointment, err
}

func CancelUserAppointment(userID, appointmentID uint) error {
	tx := database.DB.Begin()

	// 直接查询预约信息（无需预加载Appointments）
	var appointment Appointment
	if err := tx.First(&appointment, appointmentID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("预约不存在")
		}
		return errors.New("查询预约失败")
	}

	// 验证用户权限
	if appointment.UserID != userID {
		tx.Rollback()
		return errors.New("无权操作此预约")
	}

	// 检查是否允许取消
	if appointment.Status != "pending" {
		tx.Rollback()
		return errors.New("当前状态不允许取消")
	}

	// 更新预约状态
	if err := tx.Model(&Appointment{}).Where("id = ?", appointmentID).
		Update("status", "canceled").Error; err != nil {
		tx.Rollback()
		return err
	}

	// 释放时间段
	if err := tx.Model(&TimeSlot{}).Where("id = ?", appointment.TimeSlotID).
		Update("is_available", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果使用了优惠券，恢复优惠券
	// 注意：这里需要确保已经预加载了Coupon关系
	if appointment.Coupon != nil {
		if err := tx.Model(&UserCoupon{}).Where("id = ?", appointment.Coupon.ID).
			Updates(map[string]interface{}{
				"status":  "unused",
				"used_at": nil,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func UpdateAppointment(appointment *Appointment) error {
	result := database.DB.Save(appointment)
	return result.Error
}

//func CancelUserAppointment(userID, appointmentID uint) error {
//	tx := database.DB.Begin()
//
//	// 获取预约信息
//	var appointment Appointment
//	if err := tx.Preload("Appointments").First(&appointment, appointmentID).Error; err != nil {
//		tx.Rollback()
//		return err
//	}
//
//	// 验证用户权限
//	if appointment.UserID != userID {
//		tx.Rollback()
//		return errors.New("无权操作此预约")
//	}
//
//	// 检查是否允许取消
//	if appointment.Status != "pending" && appointment.Status != "confirmed" {
//		tx.Rollback()
//		return errors.New("当前状态不允许取消")
//	}
//
//	// 更新预约状态
//	if err := tx.Model(&Appointment{}).Where("id = ?", appointmentID).
//		Update("status", "canceled").Error; err != nil {
//		tx.Rollback()
//		return err
//	}
//
//	// 释放时间段
//	if err := tx.Model(&TimeSlot{}).Where("id = ?", appointment.TimeSlotID).
//		Update("is_available", true).Error; err != nil {
//		tx.Rollback()
//		return err
//	}
//
//	// 如果使用了优惠券，恢复优惠券
//	if appointment.Coupon != nil {
//		if err := tx.Model(&UserCoupon{}).Where("id = ?", appointment.Coupon.ID).
//			Updates(map[string]interface{}{
//				"status":  "unused",
//				"used_at": nil,
//			}).Error; err != nil {
//			tx.Rollback()
//			return err
//		}
//	}
//
//	return tx.Commit().Error
//}

func GetAppointmentByID(id uint) (*Appointment, error) {
	var appointment Appointment
	result := database.DB.Preload("TimeSlot").
		Where("id = ?", id).
		First(&appointment)

	if result.Error != nil {
		return nil, result.Error
	}
	return &appointment, nil
}
