package models

import (
	"admin-api/database"
	"time"
)

// Banner 轮播图模型
type Banner struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:100;not null" json:"title"`     // 轮播图标题
	Description string    `gorm:"size:255" json:"description"`        // 描述
	ImageURL    string    `gorm:"size:255;not null" json:"image_url"` // 图片URL
	LinkURL     string    `gorm:"size:255" json:"link_url"`           // 跳转链接
	Position    string    `gorm:"size:50;not null" json:"position"`   // 展示位置(home-首页, category-分类页)
	Platform    string    `gorm:"size:50;not null" json:"platform"`   // 平台(all, weapp, h5, app)
	Sort        int       `gorm:"default:0" json:"sort"`              // 排序(数字越小越靠前)
	Status      int       `gorm:"default:1" json:"status"`            // 状态(0-禁用,1-启用)
	StartTime   time.Time `json:"start_time"`                         // 开始展示时间
	EndTime     time.Time `json:"end_time"`                           // 结束展示时间
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateBanner 创建轮播图
func CreateBanner(banner *Banner) error {
	result := database.DB.Create(banner)
	return result.Error
}

// UpdateBanner 更新轮播图
func UpdateBanner(banner *Banner) error {
	result := database.DB.Save(banner)
	return result.Error
}

// DeleteBanner 删除轮播图
func DeleteBanner(id uint) error {
	result := database.DB.Delete(&Banner{}, id)
	return result.Error
}

// GetBannerByID 根据ID获取轮播图
func GetBannerByID(id uint) (*Banner, error) {
	var banner Banner
	result := database.DB.First(&banner, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &banner, nil
}

// GetBanners 获取轮播图列表（管理端用）
func GetBanners(position, platform string, status int) ([]Banner, error) {
	var banners []Banner
	query := database.DB.Order("sort asc")

	if position != "" {
		query = query.Where("position = ?", position)
	}
	if platform != "" {
		query = query.Where("platform = ? OR platform = 'all'", platform)
	}
	if status != 0 {
		query = query.Where("status = ?", status)
	}

	result := query.Find(&banners)
	return banners, result.Error
}

// GetActiveBanners 获取有效的轮播图（用户端用）
func GetActiveBanners(position, platform string, now time.Time) ([]Banner, error) {
	var banners []Banner
	result := database.DB.
		Where("position = ?", position).
		Where("(platform = ? OR platform = 'all')", platform).
		Where("status = 1").
		Where("start_time <= ?", now).
		Where("end_time >= ?", now).
		Order("sort asc").
		Find(&banners)

	return banners, result.Error
}
