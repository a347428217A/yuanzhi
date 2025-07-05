package models

import (
	"admin-api/database"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// ServiceCategory 服务类别模型
type ServiceCategory struct {
	ID         uint      `gorm:"primaryKey"`
	MerchantID uint      `gorm:"index;not null"`     // 关联商家ID
	Name       string    `gorm:"size:100;not null"`  // 类别名称
	Sort       int       `gorm:"default:0;not null"` // 排序权重
	CreatedAt  time.Time `gorm:"autoCreateTime"`     // 自动设置创建时间
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`     // 自动设置更新时间
}

// CreateServiceCategory 创建服务类别
func CreateServiceCategory(merchantID uint, name string, sort int) (*ServiceCategory, error) {
	// 基本字段验证
	if name == "" {
		return nil, errors.New("类别名称不能为空")
	}

	if len(name) > 100 {
		return nil, errors.New("类别名称过长")
	}

	// 检查商家是否存在
	if err := database.DB.First(&Merchant{}, merchantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商家不存在")
		}
		return nil, fmt.Errorf("查询商家失败: %w", err)
	}

	// 检查是否已存在同名类别
	var count int64
	if err := database.DB.Model(&ServiceCategory{}).
		Where("merchant_id = ? AND name = ?", merchantID, name).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查类别名称失败: %w", err)
	}

	if count > 0 {
		return nil, errors.New("该类别名称已存在")
	}

	// 创建类别对象
	category := &ServiceCategory{
		MerchantID: merchantID,
		Name:       name,
		Sort:       sort,
	}

	// 保存到数据库
	if err := database.DB.Create(category).Error; err != nil {
		return nil, fmt.Errorf("保存服务类别失败: %w", err)
	}

	return category, nil
}

// GetCategoriesByMerchant 获取商家的所有服务类别
func GetCategoriesByMerchant(merchantID uint) ([]ServiceCategory, error) {
	var categories []ServiceCategory
	if err := database.DB.
		Where("merchant_id = ?", merchantID).
		Order("sort DESC").
		Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// UpdateCategory 更新服务类别
func UpdateCategory(categoryID, merchantID uint, name string, sort int) (*ServiceCategory, error) {
	// 验证类别存在且属于该商家
	var category ServiceCategory
	if err := database.DB.
		Where("id = ? AND merchant_id = ?", categoryID, merchantID).
		First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("服务类别不存在")
		}
		return nil, err
	}

	// 更新字段
	updates := map[string]interface{}{}

	if name != "" {
		// 检查名称是否已存在
		var count int64
		if err := database.DB.Model(&ServiceCategory{}).
			Where("merchant_id = ? AND name = ? AND id != ?", merchantID, name, categoryID).
			Count(&count).Error; err != nil {
			return nil, fmt.Errorf("检查类别名称失败: %w", err)
		}

		if count > 0 {
			return nil, errors.New("该类别名称已被使用")
		}

		updates["name"] = name
	}

	if sort >= 0 {
		updates["sort"] = sort
	}

	// 执行更新
	if len(updates) > 0 {
		if err := database.DB.Model(&category).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新服务类别失败: %w", err)
		}
	}

	return &category, nil
}

// DeleteCategory 删除服务类别
func DeleteCategory(categoryID, merchantID uint) error {
	// 验证类别存在且属于该商家
	var count int64
	if err := database.DB.Model(&ServiceCategory{}).
		Where("id = ? AND merchant_id = ?", categoryID, merchantID).
		Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		return errors.New("服务类别不存在")
	}

	// 检查是否有服务使用该类别
	var serviceCount int64
	if err := database.DB.Model(&Service{}).
		Where("category_id = ?", categoryID).
		Count(&serviceCount).Error; err != nil {
		return err
	}

	if serviceCount > 0 {
		return errors.New("该类别下存在服务，无法删除")
	}

	// 删除类别
	if err := database.DB.Delete(&ServiceCategory{}, categoryID).Error; err != nil {
		return fmt.Errorf("删除服务类别失败: %w", err)
	}

	return nil
}
