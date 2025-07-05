package models

import (
	"admin-api/database"
	"time"
	//"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Openid    string `gorm:"size:64;uniqueIndex;not null"`
	Nickname  string `gorm:"size:64"`
	Avatar    string `gorm:"size:255"`
	Phone     string `gorm:"size:20;index"`
	Points    int    `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func FindOrCreateUserByOpenID(openID string) (*User, error) {
	var user User
	result := database.DB.Where(User{Openid: openID}).FirstOrCreate(&user, User{Openid: openID})
	return &user, result.Error
}

func UpdateUserPhone(userID uint, phone string) error {
	return database.DB.Model(&User{}).Where("id = ?", userID).Update("phone", phone).Error
}

func GetUserByID(userID uint) (*User, error) {
	var user User
	err := database.DB.First(&user, userID).Error
	return &user, err
}

type AppointmentStats struct {
	Total     int64
	Completed int64
	Upcoming  int64
}

func GetUserAppointmentStats(userID uint) (*AppointmentStats, error) {
	var stats AppointmentStats

	// 获取总预约数
	if err := database.DB.Model(&Appointment{}).
		Where("user_id = ?", userID).
		Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 获取已完成预约数
	if err := database.DB.Model(&Appointment{}).
		Where("user_id = ? AND status = ?", userID, "completed").
		Count(&stats.Completed).Error; err != nil {
		return nil, err
	}

	// 获取即将到来的预约数
	now := time.Now()
	if err := database.DB.Model(&Appointment{}).
		Where("user_id = ? AND status IN (?) AND appointment_date >= ?",
			userID, []string{"pending", "confirmed"}, now.Format("2006-01-02")).
		Count(&stats.Upcoming).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}
