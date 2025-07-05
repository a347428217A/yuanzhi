package merchant

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

// GetMerchantAppointments 获取商家预约列表
// @Summary 获取商家预约列表
// @Description 获取当前商户的预约列表，可按状态和日期筛选
// @Tags 预约管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param status query string false "预约状态（pending, confirmed, completed, canceled）"
// @Param date query string false "预约日期（格式: YYYY-MM-DD）"
// @Success 200 {array} map[string]string "预约列表"
// @Failure 401 {object} map[string]string "未授权" Example({"error": "身份认证失败"})
// @Failure 500 {object} map[string]string "服务器错误" Example({"error": "获取预约列表失败"})
// @Router /api/merchant/appointments [get]
func GetMerchantAppointments(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	status := c.Query("status")
	date := c.Query("date")

	// 解析日期参数
	var parsedDate *time.Time
	if date != "" {
		d, err := time.Parse("2006-01-02", date)
		if err == nil {
			parsedDate = &d
		}
	}

	appointments, err := models.GetMerchantAppointments(merchantID, status, parsedDate)
	if err != nil {
		utils.InternalError(c, "获取预约列表失败")
		return
	}

	// 转换为响应格式
	type AppointmentResponse struct {
		ID              uint   `json:"id"`
		OrderNo         string `json:"order_no"`
		UserName        string `json:"user_name"`
		UserPhone       string `json:"user_phone"`
		ServiceName     string `json:"service_name"`
		StaffName       string `json:"staff_name"`
		AppointmentDate string `json:"appointment_date"`
		StartTime       string `json:"start_time"`
		EndTime         string `json:"end_time"`
		Status          string `json:"status"`
		Amount          int    `json:"amount"`
		CreatedAt       string `json:"created_at"`
	}

	response := make([]AppointmentResponse, 0, len(appointments))
	for _, appt := range appointments {
		response = append(response, AppointmentResponse{
			ID:              appt.ID,
			OrderNo:         appt.OrderNo,
			UserName:        appt.User.Nickname,
			UserPhone:       appt.User.Phone,
			ServiceName:     appt.Service.Name,
			StaffName:       appt.Staff.Name,
			AppointmentDate: appt.AppointmentDate.Format("2006-01-02"),
			StartTime:       appt.StartTime,
			EndTime:         appt.EndTime,
			Status:          appt.Status,
			Amount:          appt.Amount,
			CreatedAt:       appt.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	utils.Success(c, response)
}

type UpdateAppointRequest struct {
	Status string `json:"status" binding:"required,oneof=confirmed completed canceled rejected"`
	Reason string `json:"reason"`
}

// 更新预约状态
// UpdateAppointmentStatus 更新预约状态
// @Summary      更新预约状态
// @Description  商家更新预约的状态（confirmed/completed/canceled/rejected）
// @Tags         商家预约管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        appointmentId path int true "预约ID"
// @Param        body body UpdateAppointRequest true "状态更新请求"
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {object}  utils.Response "状态更新成功"
// @Failure      400  {object}  utils.Response "无效的预约ID | 参数错误 | 无效的状态转换"
// @Failure      401  {object}  utils.Response "认证失败"
// @Failure      404  {object}  utils.Response "预约不存在"
// @Failure      500  {object}  utils.Response "更新状态失败"
// @Router       /api/merchant/appointments/{appointmentId}/status [put]
func UpdateAppointmentStatus(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	appointmentID, err := strconv.Atoi(c.Param("appointmentId"))
	if err != nil {
		utils.BadRequest(c, "无效的预约ID")
		return
	}

	var req UpdateAppointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		fmt.Println(err)
		return
	}

	// 验证预约属于该商家
	appointment, err := models.GetAppointmentByID(uint(appointmentID))

	fmt.Println("1111111111111", merchantID)

	if err != nil || appointment.MerchantID != merchantID {
		fmt.Println(err)
		utils.NotFound(c, "预约不存在")
		return
	}

	// 检查状态转换是否有效
	if !isValidStatusTransition(appointment.Status, req.Status) {
		utils.BadRequest(c, "无效的状态转换")
		return
	}

	// 更新状态
	if err := models.UpdateAppointmentStatus(uint(appointmentID), req.Status, req.Reason); err != nil {
		utils.InternalError(c, "更新状态失败")
		return
	}

	// 如果是取消或拒绝，释放时间段
	if req.Status == "canceled" || req.Status == "rejected" {
		if err := models.ReleaseTimeSlot(appointment.TimeSlotID); err != nil {
			// 记录错误但继续
			log.Printf("释放时间段失败: %v", err)
		}
	}

	// TODO: 发送状态变更通知给用户

	utils.Success(c, "状态更新成功")
}

func isValidStatusTransition(oldStatus, newStatus string) bool {
	validTransitions := map[string][]string{
		"pending":   {"confirmed", "rejected"},
		"confirmed": {"completed", "canceled"},
		"paid":      {"completed", "refunding"},
		"completed": {},
		"canceled":  {},
		"rejected":  {},
	}

	for _, s := range validTransitions[oldStatus] {
		if s == newStatus {
			return true
		}
	}
	return false
}
