package models

import (
	"admin-api/database"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type CouponTemplate struct {
	ID            uint   `gorm:"primaryKey"`
	MerchantID    uint   `gorm:"not null"`
	Name          string `gorm:"size:100;not null"`
	Description   string `gorm:"type:text"`
	DiscountType  string `gorm:"size:10;default:'fixed';not null"` // fixed, percent
	DiscountValue int    `gorm:"type:int;not null"`
	MinAmount     int    `gorm:"type:int;not null"`
	ValidityDays  int    `gorm:"default:7;not null"`
	TotalCount    int    `gorm:"default:0;not null"`
	//IsActive      bool    `gorm:"default:true;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserCoupon struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"index;not null"`
	TemplateID    uint      `gorm:"index;not null"`
	CouponCode    string    `gorm:"size:20;uniqueIndex;not null"`
	Status        string    `gorm:"size:10;default:'unused';not null"` // unused, used, expired
	ValidFrom     time.Time `gorm:"type:date;not null"`
	ValidTo       time.Time `gorm:"type:date;not null"`
	UsedAt        time.Time
	AppointmentID *uint `gorm:"index"`
	CreatedAt     time.Time

	Template CouponTemplate `gorm:"foreignKey:TemplateID"`
}

type CouponApplication struct {
	OriginalPrice int
	FinalPrice    int
	Discount      int
	UserCoupon    *UserCoupon
}

func ApplyCoupon(tx *gorm.DB, userID, couponID uint, originalPrice int) (*CouponApplication, error) {
	// 1. 获取用户优惠券
	var userCoupon UserCoupon
	if err := tx.
		Where("user_id = ? AND id = ?", userID, couponID).
		Preload("Template").
		First(&userCoupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("优惠券不存在")
		}
		return nil, fmt.Errorf("查询优惠券失败: %w", err)
	}

	// 2. 验证优惠券状态
	if userCoupon.Status != "unused" {
		return nil, fmt.Errorf("优惠券状态无效: %s", userCoupon.Status)
	}

	// 3. 验证有效期
	now := time.Now()
	if now.Before(userCoupon.ValidFrom) {
		return nil, fmt.Errorf("优惠券尚未生效（生效时间: %s）", userCoupon.ValidFrom.Format("2006-01-02"))
	}
	if now.After(userCoupon.ValidTo) {
		return nil, fmt.Errorf("优惠券已过期（过期时间: %s）", userCoupon.ValidTo.Format("2006-01-02"))
	}

	// 4. 验证使用条件
	template := userCoupon.Template
	if originalPrice < template.MinAmount {
		return nil, fmt.Errorf("未达到最低消费金额 %.2f", template.MinAmount)
	}
	//if template.ServiceType != "" && template.ServiceType != "any" {
	//	// 这里需要根据实际服务类型验证
	//	// if serviceType != template.ServiceType { ... }
	//}

	// 5. 计算折扣
	result := &CouponApplication{
		OriginalPrice: originalPrice,
		UserCoupon:    &userCoupon,
	}

	switch template.DiscountType {
	case "fixed":
		result.Discount = template.DiscountValue
	case "percent":
		result.Discount = originalPrice * template.DiscountValue / 100
	default:
		return nil, fmt.Errorf("未知的折扣类型: %s", template.DiscountType)
	}

	//// 6. 应用折扣上限
	//if template.MaxDiscount > 0 && result.Discount > template.MaxDiscount {
	//	result.Discount = template.MaxDiscount
	//}

	// 7. 计算最终价格
	result.FinalPrice = originalPrice - result.Discount
	if result.FinalPrice < 0 {
		result.FinalPrice = 0
	}

	// 8. 标记优惠券为使用中（非最终使用）
	if err := tx.Model(&userCoupon).
		Update("status", "using").Error; err != nil {
		return nil, fmt.Errorf("锁定优惠券失败: %w", err)
	}

	return result, nil

	//// 获取用户优惠券
	//var userCoupon UserCoupon
	//if err := tx.Where("user_id = ? AND id = ?", userID, couponID).
	//	Preload("Template").First(&userCoupon).Error; err != nil {
	//	return nil, errors.New("优惠券不存在")
	//}
	//
	//// 检查优惠券状态
	//if userCoupon.Status != "unused" {
	//	return nil, errors.New("优惠券已使用或过期")
	//}
	//
	//// 检查有效期
	//now := time.Now()
	//if now.Before(userCoupon.ValidFrom) || now.After(userCoupon.ValidTo) {
	//	return nil, errors.New("优惠券不在有效期内")
	//}
	//
	//// 检查最低消费金额
	//template := userCoupon.Template
	//if originalPrice < template.MinAmount {
	//	return nil, fmt.Errorf("未满足最低消费金额%.2f", template.MinAmount)
	//}
	//
	//// 计算折扣
	//result := &CouponApplication{
	//	OriginalPrice: originalPrice,
	//	UserCoupon:    &userCoupon,
	//}
	//
	//switch template.DiscountType {
	//case "fixed":
	//	result.Discount = template.Discount
	//case "percent":
	//	result.Discount = originalPrice * template.Discount / 100
	//default:
	//	return nil, errors.New("无效的折扣类型")
	//}
	//
	//// 计算最终价格
	//result.FinalPrice = originalPrice - result.Discount
	//if result.FinalPrice < 0 {
	//	result.FinalPrice = 0
	//}
	//
	//return result, nil
}

func GetUserCoupons(userID uint, status string) ([]UserCoupon, error) {
	var coupons []UserCoupon

	query := database.DB.Preload("Template").Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("valid_to ASC").Find(&coupons).Error
	return coupons, err
}

type User_coupon struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	TemplateID uint      `gorm:"index;not null"`
	CouponCode string    `gorm:"size:20;uniqueIndex;not null"`
	Status     string    `gorm:"size:10;default:'unused';not null"` // unused, used, expired
	ValidFrom  time.Time `gorm:"type:date;not null"`
	ValidTo    time.Time `gorm:"type:date;not null"`
}

func ClaimUserCoupon(userID, templateID uint) (*User_coupon, error) {
	tx := database.DB.Begin()

	// 1. 获取优惠券模板并锁定
	var template CouponTemplate
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&template, templateID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("优惠券模板不存在")
	}

	// 检查是否还有剩余
	if template.TotalCount <= 0 {
		tx.Rollback()
		return nil, errors.New("优惠券已领完")
	}

	// 2. 生成优惠券码
	couponCode := generateCouponCode()

	// 3. 创建用户优惠券
	now := time.Now()

	coupon := &User_coupon{
		UserID:     userID,
		TemplateID: templateID,
		CouponCode: couponCode,
		Status:     "unused",
		ValidFrom:  now,
		ValidTo:    now.AddDate(0, 0, template.ValidityDays),
	}

	if err := tx.Create(coupon).Error; err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, errors.New("创建优惠券失败")
	}

	// 4. 减少优惠券模板的总量
	if err := tx.Model(&CouponTemplate{}).Where("id = ?", templateID).
		Update("total_count", gorm.Expr("total_count - 1")).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("更新优惠券模板失败")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交事务失败")
	}

	return coupon, nil
}

func generateCouponCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
func GetAvailableCoupons(merchantID uint) ([]CouponTemplate, error) {
	var templates []CouponTemplate
	err := database.DB.Where("merchant_id = ?", merchantID).Find(&templates).Error
	return templates, err
}
