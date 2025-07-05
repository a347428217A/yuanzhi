package models

import (
	"admin-api/database"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func CreateMerchant(name, address, phone, description, logo, businessHour string) (*Merchant, error) {
	// 基本字段验证
	if name == "" {
		return nil, errors.New("商家名称不能为空")
	}
	if address == "" {
		return nil, errors.New("商家地址不能为空")
	}
	if phone == "" {
		return nil, errors.New("联系电话不能为空")
	}

	// 创建商家对象
	merchant := &Merchant{
		Name:          name,
		Address:       address,
		Phone:         phone,
		Description:   description,
		Logo:          logo,
		BusinessHours: businessHour,
	}

	// 保存到数据库
	if err := database.DB.Create(merchant).Error; err != nil {
		return nil, err
	}

	return merchant, nil
}

func CreateMerchantAdmin(merchantID uint, username, password, role string) (*MerchantAdmin, error) {
	// 验证商家是否存在
	var merchant Merchant
	if err := database.DB.First(&merchant, merchantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商家不存在")
		}
		return nil, fmt.Errorf("查询商家失败: %w", err)
	}

	// 检查用户名是否已存在
	var count int64
	if err := database.DB.Model(&MerchantAdmin{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if count > 0 {
		return nil, errors.New("用户名已被使用")
	}

	// 验证角色
	if role != "admin" && role != "staff" {
		return nil, errors.New("无效的角色类型")
	}

	// 加密密码
	//hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	//if err != nil {
	//	return nil, fmt.Errorf("密码加密失败: %w", err)
	//}

	// 创建管理员对象
	admin := &MerchantAdmin{
		MerchantID: merchantID,
		Username:   username,
		Password:   password,
		Role:       role,
		IsActive:   true,
	}

	// 保存到数据库
	if err := database.DB.Create(admin).Error; err != nil {
		return nil, fmt.Errorf("保存管理员失败: %w", err)
	}

	return admin, nil
}
