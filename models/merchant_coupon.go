package models

import (
	"admin-api/database"
	"time"
)

func GetCouponTemplates(merchantID uint, page, pageSize int, status string) ([]CouponTemplate, int64, error) {
	var templates []CouponTemplate
	db := database.DB.Where("merchant_id = ?", merchantID)

	// 添加状态过滤
	now := time.Now()
	switch status {
	case "active":
		// 活跃状态：当前时间在有效期内 (创建时间 + 有效期天数 > 当前时间)
		db = db.Where("created_at + INTERVAL validity_days DAY > ?", now)
	case "expired":
		// 过期状态：当前时间已超过有效期 (创建时间 + 有效期天数 <= 当前时间)
		db = db.Where("created_at + INTERVAL validity_days DAY <= ?", now)
	}

	// 计算总数
	var total int64
	if err := db.Model(&CouponTemplate{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 添加分页
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

//func GetCouponTemplates(merchantID uint) ([]CouponTemplate, error) {
//	var templates []CouponTemplate
//	err := database.DB.Where("merchant_id = ?", merchantID).Find(&templates).Error
//	return templates, err
//}

func CreateCouponTemplate(template *CouponTemplate) error {
	return database.DB.Create(template).Error
}

func UpdateCouponTemplate(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&CouponTemplate{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteCouponTemplate(id uint) error {
	return database.DB.Delete(&CouponTemplate{}, id).Error
}
