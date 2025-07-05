package merchant

import (
	"admin-api/database"
	"time"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

type DateRange struct {
	Start string `form:"start" binding:"required"`
	End   string `form:"end" binding:"required"`
}

// AppointmentStatsResponse 预约统计响应结构
type AppointmentStatsResponse struct {
	Today  int64            `json:"today" example:"15"`  // 今日预约数
	Week   int64            `json:"week" example:"120"`  // 本周预约数
	Month  int64            `json:"month" example:"450"` // 本月预约数
	Status map[string]int64 `json:"status"`              // 各状态预约数
}

// @Summary 获取预约统计数据
// @Description 获取当前商户的预约统计数据，包括今日、本周、本月预约数以及各状态预约数
// @Tags 商户-数据统计
// @Security ApiKeyAuth
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param start_date query string false "开始日期 (格式: YYYY-MM-DD)" example("2023-06-01")
// @Param end_date query string false "结束日期 (格式: YYYY-MM-DD)" example("2023-06-30")
// @Success 200 {object} AppointmentStatsResponse "成功返回预约统计数据"
// @Failure 500 {object} utils.Response "获取数据失败"
// @Router /api/merchant/stats/appointments [get]
func GetAppointmentStats(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")

	// 添加日期参数支持
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")

	// 获取统计数据
	stats, err := models.GetAppointmentStats(merchantID, startDate, endDate)
	if err != nil {
		utils.InternalError(c, "获取数据失败: "+err.Error())
		return
	}

	utils.Success(c, stats)

	//merchantID := c.GetUint("merchant_id")
	//
	//// 获取今日预约数
	//var todayCount int64
	//today := time.Now().Format("2006-01-02")
	//if err := database.DB.Model(&models.Appointment{}).
	//	Where("merchant_id = ? AND appointment_date = ?", merchantID, today).
	//	Count(&todayCount).Error; err != nil {
	//	utils.InternalError(c, "获取数据失败")
	//	return
	//}
	//
	//// 获取本周预约数
	//var weekCount int64
	//startOfWeek := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+1).Format("2006-01-02")
	//endOfWeek := time.Now().AddDate(0, 0, 7-int(time.Now().Weekday())).Format("2006-01-02")
	//if err := database.DB.Model(&models.Appointment{}).
	//	Where("merchant_id = ? AND appointment_date BETWEEN ? AND ?",
	//		merchantID, startOfWeek, endOfWeek).
	//	Count(&weekCount).Error; err != nil {
	//	utils.InternalError(c, "获取数据失败")
	//	return
	//}
	//
	//// 获取本月预约数
	//var monthCount int64
	//startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")
	//endOfMonth := time.Now().AddDate(0, 1, -time.Now().Day()).Format("2006-01-02")
	//if err := database.DB.Model(&models.Appointment{}).
	//	Where("merchant_id = ? AND appointment_date BETWEEN ? AND ?",
	//		merchantID, startOfMonth, endOfMonth).
	//	Count(&monthCount).Error; err != nil {
	//	utils.InternalError(c, "获取数据失败")
	//	return
	//}
	//
	//// 获取不同状态预约数
	//statusStats := make(map[string]int64)
	//statuses := []string{"pending", "confirmed", "completed", "canceled", "rejected"}
	//
	//for _, status := range statuses {
	//	var count int64
	//	if err := database.DB.Model(&models.Appointment{}).
	//		Where("merchant_id = ? AND status = ?", merchantID, status).
	//		Count(&count).Error; err == nil {
	//		statusStats[status] = count
	//	}
	//}
	//
	//utils.Success(c, gin.H{
	//	"today":  todayCount,
	//	"week":   weekCount,
	//	"month":  monthCount,
	//	"status": statusStats,
	//})
}

// RevenueStatsResponse 营收统计响应结构
type RevenueStatsResponse struct {
	TotalRevenue   float64          `json:"total_revenue" example:"15000.50"` // 总收入
	DailyRevenue   []DailyRevenue   `json:"daily_revenue"`                    // 每日收入数据
	ServiceRevenue []ServiceRevenue `json:"service_revenue"`                  // 服务收入分布
}

// DailyRevenue 每日收入数据
type DailyRevenue struct {
	Date   string  `json:"date" example:"2023-06-15"` // 日期
	Amount float64 `json:"amount" example:"1200.50"`  // 当日收入
}

// ServiceRevenue 服务收入数据
type ServiceRevenue struct {
	ServiceID   uint    `json:"service_id" example:"1"`      // 服务ID
	ServiceName string  `json:"service_name" example:"基础护理"` // 服务名称
	Amount      float64 `json:"amount" example:"5000.00"`    // 该服务总收入
}

// @Summary 获取营收统计数据
// @Description 获取当前商户的营收统计数据，包括总收入、每日收入趋势和各服务收入分布
// @Tags 商户-数据统计
// @Security ApiKeyAuth
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param start query string true "开始日期 (格式: YYYY-MM-DD)" example("2023-06-01")
// @Param end query string true "结束日期 (格式: YYYY-MM-DD)" example("2023-06-30")
// @Success 200 {object} RevenueStatsResponse "成功返回营收统计数据"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 500 {object} utils.Response "获取数据失败"
// @Router /api/merchant/stats/revenue [get]
func GetRevenueStats(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")

	var dateRange DateRange
	if err := c.ShouldBindQuery(&dateRange); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// 解析日期
	startDate, err := time.Parse("2006-01-02", dateRange.Start)
	if err != nil {
		utils.BadRequest(c, "无效的开始日期")
		return
	}

	endDate, err := time.Parse("2006-01-02", dateRange.End)
	if err != nil {
		utils.BadRequest(c, "无效的结束日期")
		return
	}

	// 获取总收入
	var totalRevenue float64
	if err := database.DB.Model(&models.Appointment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("merchant_id = ? AND status = 'completed' AND appointment_date BETWEEN ? AND ?",
			merchantID, startDate, endDate).
		Scan(&totalRevenue).Error; err != nil {
		utils.InternalError(c, "获取收入数据失败")
		return
	}

	// 获取每日收入
	type DailyRevenue struct {
		Date   string  `json:"date"`
		Amount float64 `json:"amount"`
	}

	var dailyRevenues []DailyRevenue
	rows, err := database.DB.Model(&models.Appointment{}).
		Select("appointment_date as date, SUM(amount) as amount").
		Where("merchant_id = ? AND status = 'completed' AND appointment_date BETWEEN ? AND ?",
			merchantID, startDate, endDate).
		Group("appointment_date").
		Order("appointment_date ASC").
		Rows()

	if err != nil {
		utils.InternalError(c, "获取每日收入失败")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var dr DailyRevenue
		var date time.Time
		if err := rows.Scan(&date, &dr.Amount); err == nil {
			dr.Date = date.Format("2006-01-02")
			dailyRevenues = append(dailyRevenues, dr)
		}
	}

	// 获取服务收入分布
	type ServiceRevenue struct {
		ServiceID   uint    `json:"service_id"`
		ServiceName string  `json:"service_name"`
		Amount      float64 `json:"amount"`
	}

	var serviceRevenues []ServiceRevenue
	rows, err = database.DB.Model(&models.Appointment{}).
		Select("services.id as service_id, services.name as service_name, SUM(appointments.amount) as amount").
		Joins("JOIN services ON services.id = appointments.service_id").
		Where("appointments.merchant_id = ? AND appointments.status = 'completed' AND appointments.appointment_date BETWEEN ? AND ?",
			merchantID, startDate, endDate).
		Group("services.id, services.name").
		Rows()

	if err != nil {
		utils.InternalError(c, "获取服务收入分布失败")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sr ServiceRevenue
		if err := rows.Scan(&sr.ServiceID, &sr.ServiceName, &sr.Amount); err == nil {
			serviceRevenues = append(serviceRevenues, sr)
		}
	}

	utils.Success(c, gin.H{
		"total_revenue":   totalRevenue,
		"daily_revenue":   dailyRevenues,
		"service_revenue": serviceRevenues,
	})
}
