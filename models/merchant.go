package models

import (
	"admin-api/database"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Merchant struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"size:100;not null"`
	Address       string `gorm:"size:255;not null"`
	Phone         string `gorm:"size:20;not null"`
	Description   string `gorm:"type:text"`
	Logo          string `gorm:"size:255"`
	BusinessHours string `gorm:"size:100"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type MerchantAdmin struct {
	ID         uint   `gorm:"primaryKey"`
	MerchantID uint   `gorm:"index;not null"`
	Username   string `gorm:"size:50;uniqueIndex;not null"`
	Password   string `gorm:"size:100;not null"`
	Role       string `gorm:"size:20;default:'staff';not null"`
	IsActive   bool   `gorm:"default:true;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func GetMerchantByID(id uint) (*Merchant, error) {
	var merchant Merchant
	err := database.DB.First(&merchant, id).Error
	return &merchant, err
}

func GetMerchantAdminByUsername(username string) (*MerchantAdmin, error) {
	var admin MerchantAdmin
	err := database.DB.Where("username = ?", username).First(&admin).Error
	return &admin, err
}

//func GetMerchantServices(merchantID uint) ([]Service, error) {
//	var services []Service
//	err := database.DB.Where("merchant_id = ?", merchantID).Find(&services).Error
//	return services, err
//}
//
//func GetMerchantStaff(merchantID uint) ([]Staff, error) {
//	var staff []Staff
//	err := database.DB.Where("merchant_id = ?", merchantID).Find(&staff).Error
//	return staff, err
//}

func GetMerchantAppointments(merchantID uint, status string, date *time.Time) ([]Appointment, error) {
	var appointments []Appointment

	query := database.DB.Preload("User").Preload("Service").Preload("Staff").
		Where("merchant_id = ?", merchantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if date != nil {
		query = query.Where("appointment_date = ?", date.Format("2006-01-02"))
	}

	err := query.Order("appointment_date ASC, start_time ASC").Find(&appointments).Error
	return appointments, err
}

func UpdateAppointmentStatus(appointmentID uint, status, reason string) error {
	return database.DB.Model(&Appointment{}).
		Where("id = ?", appointmentID).
		Updates(map[string]interface{}{
			"status": status,
			"remark": gorm.Expr("CONCAT(remark, ?)", " | 商家备注: "+reason),
		}).Error
}

func ReleaseTimeSlot(timeSlotID uint) error {
	return database.DB.Model(&TimeSlot{}).
		Where("id = ?", timeSlotID).
		Update("is_available", true).Error
}

func GetRecommendedMerchants() ([]Merchant, error) {
	var merchants []Merchant
	err := database.DB.Limit(10).Order("id DESC").Find(&merchants).Error
	return merchants, err
}

// 在 models 包中定义 TimeOnly 类型
type TimeOnly time.Time

func (t *TimeOnly) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		parsedTime, err := time.Parse("15:04:05", string(v))
		if err != nil {
			return err
		}
		*t = TimeOnly(parsedTime)
		return nil
	case time.Time:
		*t = TimeOnly(v)
		return nil
	default:
		return fmt.Errorf("无法扫描类型 %T 到 TimeOnly", value)
	}
}

func (t TimeOnly) Value() (driver.Value, error) {
	return time.Time(t).Format("15:04:05"), nil
}

func (t TimeOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format("15:04:05"))
}

func (t *TimeOnly) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsedTime, err := time.Parse("15:04:05", s)
	if err != nil {
		return err
	}
	*t = TimeOnly(parsedTime)
	return nil
}

// 获取商家列表（分页+搜索）
func GetMerchants(name, phone string, page, limit int) ([]Merchant, int64, error) {
	var merchants []Merchant
	var total int64

	query := database.DB.Model(&Merchant{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if phone != "" {
		query = query.Where("phone = ?", phone)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&merchants).Error

	return merchants, total, err
}

// 获取商家管理员列表（分页+过滤）
func GetMerchantAdmins(merchantID uint, username, role string, page, limit int) ([]MerchantAdmin, int64, error) {
	var admins []MerchantAdmin
	var total int64

	query := database.DB.Where("merchant_id = ?", merchantID)
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// 获取总数
	if err := query.Model(&MerchantAdmin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&admins).Error

	return admins, total, err
}

// 获取单个管理员（验证商家ID匹配）
func GetMerchantAdminByID(adminID, merchantID uint) (*MerchantAdmin, error) {
	var admin MerchantAdmin
	err := database.DB.Where("id = ? AND merchant_id = ?", adminID, merchantID).First(&admin).Error
	return &admin, err
}
