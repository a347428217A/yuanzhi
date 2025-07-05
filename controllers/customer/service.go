// controllers/customer/service.go
package customer

import (
	"fmt"
	"strconv"
	"time"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

// 获取商家服务列表
// @Summary 获取商家服务列表
// @Description 根据商家ID获取服务列表，支持按分类ID筛选
// @Tags 商家服务
// @Accept json
// @Produce json
// @Param merchantId path int true "商家ID" Example(123)
// @Param categoryId query int false "分类ID" Example(456)
// @Success 200 {object} models.Service "成功返回服务列表"
// @Failure 400 {object} utils.Response "无效的商家ID"
// @Failure 500 {object} utils.Response "获取服务列表失败"
// @Router /api/customer/merchants/{merchantId}/services [get]
func GetMerchantServices(c *gin.Context) {
	merchantID, err := strconv.Atoi(c.Param("merchantId"))
	if err != nil {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	categoryID, _ := strconv.Atoi(c.Query("categoryId"))

	services, err := models.GetMerchantServices(uint(merchantID), uint(categoryID))
	if err != nil {
		utils.InternalError(c, "获取服务列表失败")
		return
	}

	utils.Success(c, services)
}

// 获取服务可选员工
// @Summary 获取服务的可用员工
// @Description 根据服务ID获取可提供该服务的员工列表
// @Tags 服务管理
// @Accept json
// @Produce json
// @Param serviceId path int true "服务ID" Example(789)
// @Success 200 {array} models.Staff "成功返回可用员工列表"
// @Failure 400 {object} utils.Response "无效的服务ID"
// @Failure 500 {object} utils.Response "获取可选员工失败"
// @Router /api/customer/services/{serviceId}/staff [get]
func GetServiceAvailableStaff(c *gin.Context) {
	serviceID, err := strconv.Atoi(c.Param("serviceId"))
	if err != nil {
		utils.BadRequest(c, "无效的服务ID")
		return
	}

	staff, err := models.GetServiceAvailableStaff(uint(serviceID))
	if err != nil {
		utils.InternalError(c, "获取可选员工失败")
		return
	}

	utils.Success(c, staff)
}

// 获取可预约日期
// @Summary 获取可预约日期列表
// @Description 查询指定商家、技师和服务的可预约日期范围（默认14天内）
// @Tags 时间槽管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param merchantId query int true "商家ID"
// @Param staffId query int false "技师ID"
// @Param serviceId query int false "服务ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {array} string "成功返回可预约日期列表（YYYY-MM-DD格式）"
// @Failure 400 {object} utils.Response "无效的商家ID"
// @Failure 500 {object} utils.Response "获取可预约日期失败"
// @Router /api/customer/timeslots/dates [get]
func GetAvailableDates(c *gin.Context) {
	merchantID, err := strconv.Atoi(c.Query("merchantId"))
	if err != nil {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	staffID, _ := strconv.Atoi(c.Query("staffId"))
	serviceID, _ := strconv.Atoi(c.Query("serviceId"))

	dates, err := models.GetAvailableDates(uint(merchantID), uint(staffID), uint(serviceID), 14)
	if err != nil {
		utils.InternalError(c, "获取可预约日期失败")
		return
	}

	// 转换为字符串数组
	var dateStrings []string
	for _, date := range dates {
		dateStrings = append(dateStrings, date.Format("2006-01-02"))
	}

	utils.Success(c, dateStrings)
}

// 转换格式
type SlotResponse struct {
	ID        uint   `json:"id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// 获取某天的可预约时间段
// @Summary 获取可预约时间段
// @Description 查询指定商家、技师在指定日期的可预约时间段
// @Tags 时间槽管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param merchantId query int true "商家ID"
// @Param staffId query int true "技师ID"
// @Param date query string true "日期 (格式: YYYY-MM-DD)" Example(2023-06-15)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {array} SlotResponse "成功返回可预约时间段列表"
// @Failure 400 {object} utils.Response "参数错误（无效的商家ID、员工ID或日期格式）"
// @Failure 500 {object} utils.Response "获取时间段失败"
// @Router /api/customer/timeslots [get]
func GetTimeSlots(c *gin.Context) {
	merchantID, err := strconv.Atoi(c.Query("merchantId"))
	if err != nil {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	staffID, err := strconv.Atoi(c.Query("staffId"))
	if err != nil {
		utils.BadRequest(c, "无效的员工ID")
		return
	}

	dateStr := c.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		utils.BadRequest(c, "无效的日期格式")
		return
	}

	slots, err := models.GetAvailableTimeSlots(uint(merchantID), uint(staffID), date)
	if err != nil {
		utils.InternalError(c, "获取时间段失败")
		return
	}

	var response []SlotResponse
	for _, slot := range slots {
		response = append(response, SlotResponse{
			ID:        slot.ID,
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
		})
	}

	utils.Success(c, response)
}

type CreateAppointmentRequest struct {
	MerchantID uint   `json:"merchant_id" binding:"required"`
	ServiceID  uint   `json:"service_id" binding:"required"`
	StaffID    uint   `json:"staff_id" binding:"required"`
	TimeSlotID uint   `json:"time_slot_id" binding:"required"`
	Date       string `json:"date" binding:"required"`
	CouponID   uint   `json:"coupon_id"` // 可选
	Remark     string `json:"remark"`
}

// 创建预约
// @Summary 创建预约
// @Description 用户创建新的服务预约
// @Tags 预约管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer Token"
// @Param body body CreateAppointmentRequest true "预约创建请求"
// @Success 200 {object} utils.Response "成功返回预约信息"
// @Failure 400 {object} utils.Response "参数错误或日期格式错误"
// @Failure 500 {object} utils.Response "创建预约失败"
// @Router /api/customer/appointments [post]
func CreateAppointment(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		utils.BadRequest(c, "参数错误")
		return
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		utils.BadRequest(c, "日期格式错误")
		fmt.Println(err)
		return
	}

	// 创建预约
	appointment, err := models.CreateCustomerAppointment(userID, req.MerchantID, req.ServiceID,
		req.StaffID, req.TimeSlotID, date, req.CouponID, req.Remark)
	if err != nil {
		utils.InternalError(c, "创建预约失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"appointment_id":   appointment.ID,
		"order_no":         appointment.OrderNo,
		"appointment_date": appointment.AppointmentDate.Format("2006-01-02"),
		"start_time":       appointment.StartTime,
		"end_time":         appointment.EndTime,
		"status":           appointment.Status,
	})
}
