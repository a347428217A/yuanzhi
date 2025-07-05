package models

import (
	"admin-api/database"
	"time"
)

type Staff struct {
	ID          uint   `gorm:"primaryKey"`
	MerchantID  uint   `gorm:"index;not null"`
	Name        string `gorm:"size:50;not null"`
	Avatar      string `gorm:"size:255"`
	Position    string `gorm:"size:50"`
	Description string `gorm:"type:text"`
	Specialties string `gorm:"type:text"`
	IsActive    bool   `gorm:"default:true;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Staff) TableName() string {
	return "staff"
}

func GetMerchantStaff(merchantID uint) ([]Staff, error) {
	var staff []Staff
	err := database.DB.Where("merchant_id = ?", merchantID).Find(&staff).Error
	return staff, err
}

func CreateStaff(staff *Staff) error {
	return database.DB.Create(staff).Error
}

func UpdateStaff(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&Staff{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteStaff(id uint) error {
	result := database.DB.Model(&Staff{}).Where("id = ?", id).Update("is_active", false)
	return result.Error
	//return database.DB.Delete(&Staff{}, id).Error
}
