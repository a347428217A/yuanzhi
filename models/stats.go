package models

import (
	"admin-api/database"
	"github.com/gin-gonic/gin"
	"time"
)

func GetAppointmentStats(merchantID uint, startDate, endDate string) (gin.H, error) {
	stats := gin.H{}
	now := time.Now()

	// 1. 获取今日预约数
	today := now.Format("2006-01-02")
	var todayCount int64
	if err := database.DB.Model(&Appointment{}).
		Where("merchant_id = ? AND appointment_date = ?", merchantID, today).
		Count(&todayCount).Error; err != nil {
		return nil, err
	}
	stats["today"] = todayCount

	// 2. 获取本周预约数
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 周日转为7
	}
	startOfWeek := now.AddDate(0, 0, -weekday+1).Format("2006-01-02")
	endOfWeek := now.AddDate(0, 0, 7-weekday).Format("2006-01-02")

	var weekCount int64
	if err := database.DB.Model(&Appointment{}).
		Where("merchant_id = ? AND appointment_date BETWEEN ? AND ?",
			merchantID, startOfWeek, endOfWeek).
		Count(&weekCount).Error; err != nil {
		return nil, err
	}
	stats["week"] = weekCount

	// 3. 获取本月预约数
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	var monthCount int64
	if err := database.DB.Model(&Appointment{}).
		Where("merchant_id = ? AND appointment_date BETWEEN ? AND ?",
			merchantID, startOfMonth, endOfMonth).
		Count(&monthCount).Error; err != nil {
		return nil, err
	}
	stats["month"] = monthCount

	// 4. 获取不同状态预约数
	statusStats := make(map[string]int64)
	statuses := []string{"pending", "confirmed", "completed", "canceled", "rejected"}

	for _, status := range statuses {
		var count int64
		query := database.DB.Model(&Appointment{}).Where("merchant_id = ?", merchantID)

		// 添加日期范围过滤
		if startDate != "" && endDate != "" {
			query = query.Where("appointment_date BETWEEN ? AND ?", startDate, endDate)
		}

		if err := query.Where("status = ?", status).Count(&count).Error; err == nil {
			statusStats[status] = count
		}
	}
	stats["status"] = statusStats

	// 5. 添加自定义日期范围统计（如果提供了日期范围）
	if startDate != "" && endDate != "" {
		var customCount int64
		if err := database.DB.Model(&Appointment{}).
			Where("merchant_id = ? AND appointment_date BETWEEN ? AND ?",
				merchantID, startDate, endDate).
			Count(&customCount).Error; err == nil {
			stats["custom"] = customCount
		}
	}

	return stats, nil
}
