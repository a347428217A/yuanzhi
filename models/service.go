package models

import (
	"admin-api/database"
	"time"
)

type Service struct {
	ID          uint   `gorm:"primaryKey"`
	CategoryID  uint   `gorm:"index;not null"`
	MerchantID  uint   `gorm:"index;not null"`
	Name        string `gorm:"size:100;not null"`
	Description string `gorm:"type:text"`
	CoverImage  string `gorm:"size:255"`
	Price       int    `gorm:"type:int;default:0;not null"`
	Duration    int    `gorm:"default:30;not null"` // 分钟
	IsActive    bool   `gorm:"default:true;not null"`
	Sort        int    `gorm:"default:0;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func GetMerchantServiceCategories(merchantID uint) ([]ServiceCategory, error) {
	var categories []ServiceCategory
	err := database.DB.Where("merchant_id = ?", merchantID).Order("sort ASC").Find(&categories).Error
	return categories, err
}

func GetMerchantServices(merchantID, categoryID uint) ([]Service, error) {
	var services []Service

	query := database.DB.Where("merchant_id = ?", merchantID)
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	err := query.Order("sort ASC").Find(&services).Error
	return services, err
}

func GetServiceByID(id uint) (*Service, error) {
	var service Service
	err := database.DB.First(&service, id).Error
	return &service, err
}

func GetServiceAvailableStaff(serviceID uint) ([]Staff, error) {
	// 实际项目中这里应该根据服务关联的员工来查询
	// 这里简化处理，返回商家所有在职员工
	var staff []Staff
	err := database.DB.Where("is_active = true").Find(&staff).Error
	return staff, err
}

func CreateService(service *Service) error {
	return database.DB.Create(service).Error
}

func UpdateService(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&Service{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteService(id uint) error {
	return database.DB.Delete(&Service{}, id).Error
}
