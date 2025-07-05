package customer

import (
	"strconv"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

type AppointmentResponse struct {
	ID              uint   `json:"id"`
	OrderNo         string `json:"order_no"`
	MerchantName    string `json:"merchant_name"`
	ServiceName     string `json:"service_name"`
	StaffName       string `json:"staff_name"`
	AppointmentDate string `json:"appointment_date"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Status          string `json:"status"`
	Amount          int    `json:"amount"`
	CreatedAt       string `json:"created_at"`
}

// 获取用户预约列表
// @Summary 获取用户预约列表
// @Description 获取当前登录用户的预约记录，支持按状态筛选
// @Tags 预约管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "预约状态筛选" Enums(pending, confirmed, completed, cancelled)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {array} AppointmentResponse "成功返回预约列表"
// @Failure 500 {object} utils.Response "获取预约列表失败"
// @Router /api/customer/appointments [get]
func GetUserAppointments(c *gin.Context) {
	userID := c.GetUint("user_id")
	status := c.Query("status")

	appointments, err := models.GetUserAppointments(userID, status)
	if err != nil {
		utils.InternalError(c, "获取预约列表失败")
		return
	}

	// 转换为响应格式

	response := make([]AppointmentResponse, 0, len(appointments))
	for _, appt := range appointments {
		response = append(response, AppointmentResponse{
			ID:              appt.ID,
			OrderNo:         appt.OrderNo,
			MerchantName:    appt.Merchant.Name,
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

// 构建响应
type DetailResponse struct {
	ID              uint   `json:"id"`
	OrderNo         string `json:"order_no"`
	MerchantName    string `json:"merchant_name"`
	MerchantAddress string `json:"merchant_address"`
	MerchantPhone   string `json:"merchant_phone"`
	ServiceName     string `json:"service_name"`
	ServiceImage    string `json:"service_image"`
	StaffName       string `json:"staff_name"`
	StaffAvatar     string `json:"staff_avatar"`
	AppointmentDate string `json:"appointment_date"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Status          string `json:"status"`
	Amount          int    `json:"amount"`
	Remark          string `json:"remark"`
	CreatedAt       string `json:"created_at"`
	CouponUsed      *struct {
		Code   string `json:"code"`
		Name   string `json:"name"`
		Amount int    `json:"amount"`
	} `json:"coupon_used,omitempty"`
}

// 获取预约详情
// @Summary 获取预约详情
// @Description 获取特定预约的详细信息
// @Tags 预约管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param appointmentId path int true "预约ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} DetailResponse "成功返回预约详情"
// @Failure 400 {object} utils.Response "无效的预约ID"
// @Failure 404 {object} utils.Response "预约不存在"
// @Router /api/customer/appointments/{appointmentId} [get]
func GetAppointmentDetail(c *gin.Context) {
	userID := c.GetUint("user_id")
	appointmentID, err := strconv.Atoi(c.Param("appointmentId"))
	if err != nil {
		utils.BadRequest(c, "无效的预约ID")
		return
	}

	appointment, err := models.GetUserAppointmentDetail(userID, uint(appointmentID))
	if err != nil {
		utils.NotFound(c, "预约不存在")
		return
	}

	response := DetailResponse{
		ID:              appointment.ID,
		OrderNo:         appointment.OrderNo,
		MerchantName:    appointment.Merchant.Name,
		MerchantAddress: appointment.Merchant.Address,
		MerchantPhone:   appointment.Merchant.Phone,
		ServiceName:     appointment.Service.Name,
		ServiceImage:    appointment.Service.CoverImage,
		StaffName:       appointment.Staff.Name,
		StaffAvatar:     appointment.Staff.Avatar,
		AppointmentDate: appointment.AppointmentDate.Format("2006-01-02"),
		StartTime:       appointment.StartTime,
		EndTime:         appointment.EndTime,
		Status:          appointment.Status,
		Amount:          appointment.Amount,
		Remark:          appointment.Remark,
		CreatedAt:       appointment.CreatedAt.Format("2006-01-02 15:04"),
	}

	// 如果有使用优惠券
	if appointment.Coupon != nil {
		response.CouponUsed = &struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			Amount int    `json:"amount"`
		}{
			Code:   appointment.Coupon.CouponCode,
			Name:   appointment.Coupon.Template.Name,
			Amount: appointment.Amount - appointment.Service.Price, // 计算优惠金额
		}
	}

	utils.Success(c, response)
}

// 取消预约
// @Summary 取消预约
// @Description 用户取消指定的预约
// @Tags 预约管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param appointmentId path int true "预约ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {string} string "预约已取消"
// @Failure 400 {object} utils.Response "无效的预约ID"
// @Failure 500 {object} utils.Response "取消预约失败"
// @Router /api/customer/appointments/{appointmentId}/cancel [put]
func CancelAppointment(c *gin.Context) {
	userID := c.GetUint("user_id")
	appointmentID, err := strconv.Atoi(c.Param("appointmentId"))
	if err != nil {
		utils.BadRequest(c, "无效的预约ID")
		return
	}

	if err := models.CancelUserAppointment(userID, uint(appointmentID)); err != nil {
		utils.InternalError(c, "取消预约失败: "+err.Error())
		return
	}

	utils.Success(c, "预约已取消")
}
