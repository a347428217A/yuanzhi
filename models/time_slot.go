package models

import (
	"admin-api/database"
	"time"
	//"gorm.io/gorm"
)

type TimeSlot struct {
	ID          uint      `gorm:"primaryKey"`
	MerchantID  uint      `gorm:"index;not null"`
	StaffID     uint      `gorm:"index;not null"`
	Date        time.Time `gorm:"type:date;not null"`
	StartTime   string    `gorm:"type:time;not null"`
	EndTime     string    `gorm:"type:time;not null"`
	IsAvailable bool      `gorm:"default:true;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func GetAvailableDates(merchantID, staffID, serviceID uint, days int) ([]time.Time, error) {
	// 获取服务时长
	var duration int
	if serviceID > 0 {
		var service Service
		if err := database.DB.Select("duration").First(&service, serviceID).Error; err != nil {
			return nil, err
		}
		duration = service.Duration
	}

	// 计算开始和结束日期
	now := time.Now()
	startDate := now.AddDate(0, 0, 1) // 从明天开始
	endDate := startDate.AddDate(0, 0, days-1)

	// 构建查询
	query := database.DB.Model(&TimeSlot{}).
		Select("DISTINCT date").
		Where("merchant_id = ? AND date BETWEEN ? AND ? AND is_available = true",
			merchantID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	if staffID > 0 {
		query = query.Where("staff_id = ?", staffID)
	}

	// 如果有服务时长要求，过滤掉时间不足的日期
	if duration > 0 {
		query = query.Where("TIME_TO_SEC(TIMEDIFF(end_time, start_time)) >= ?", duration*60)
	}

	var dates []struct {
		Date time.Time `gorm:"column:date"`
	}

	if err := query.Scan(&dates).Error; err != nil {
		return nil, err
	}

	// 提取日期
	var result []time.Time
	for _, d := range dates {
		result = append(result, d.Date)
	}

	return result, nil
}

func GetAvailableTimeSlots(merchantID, staffID uint, date time.Time) ([]TimeSlot, error) {
	var slots []TimeSlot
	err := database.DB.
		Where("merchant_id = ? AND staff_id = ? AND date = ? AND is_available = true",
			merchantID, staffID, date.Format("2006-01-02")).
		Order("start_time ASC").
		Find(&slots).Error
	return slots, err
}

func CheckTimeSlotAvailable(slotID uint) (bool, error) {
	var slot TimeSlot
	if err := database.DB.Select("is_available").First(&slot, slotID).Error; err != nil {
		return false, err
	}
	return slot.IsAvailable, nil
}

func BatchCreateTimeSlots(merchantID, staffID uint, date time.Time, slots []TimeSlot) error {
	tx := database.DB.Begin()

	// 1. 查找所有相关的时间段ID
	var timeSlotIDs []uint
	if err := tx.Model(&TimeSlot{}).
		Where("merchant_id = ? AND staff_id = ? AND date = ?",
			merchantID, staffID, date.Format("2006-01-02")).
		Pluck("id", &timeSlotIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 删除关联的预约
	if len(timeSlotIDs) > 0 {
		if err := tx.Where("time_slot_id IN (?)", timeSlotIDs).
			Delete(&Appointment{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 3. 删除时间段
	if err := tx.Where("merchant_id = ? AND staff_id = ? AND date = ?",
		merchantID, staffID, date.Format("2006-01-02")).
		Delete(&TimeSlot{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4. 创建新的时间段
	for _, slot := range slots {
		slot.MerchantID = merchantID
		slot.StaffID = staffID
		slot.Date = date
		if err := tx.Create(&slot).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func DeleteTimeSlot(id uint) error {
	return database.DB.Delete(&TimeSlot{}, id).Error
}
