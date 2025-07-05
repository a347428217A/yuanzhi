package merchant

import (
	"admin-api/database"
	"fmt"
	"strconv"
	"time"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

type TimeSlotRequest struct {
	StartTime string `json:"start_time" binding:"required"` // HH:MM
	EndTime   string `json:"end_time" binding:"required"`   // HH:MM
}

// @Summary 获取可用时间段
// @Description 获取指定员工在特定日期的可用时间段列表（商户端）
// @Tags 商户-时间管理
// @Security MerchantAuth
// @Accept json
// @Produce json
// @Param staff_id query int true "员工ID" example(5)
// @Param date query string true "日期 (格式: YYYY-MM-DD)" example("2023-06-15")
// @Param        Authorization header string true "Bearer Token"
// @Success 200 {array} utils.Response "成功返回时间段列表"
// @Failure 400 {object} utils.Response "日期格式错误"
// @Failure 500 {object} utils.Response "获取时间段失败"
// @Router /api/merchant/timeslots [get]
func GetTimeSlots(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	staffID, _ := strconv.Atoi(c.Query("staff_id"))
	date := c.Query("date")

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		utils.BadRequest(c, "无效的日期格式")
		return
	}

	slots, err := models.GetAvailableTimeSlots(merchantID, uint(staffID), parsedDate)
	if err != nil {
		utils.InternalError(c, "获取时间段失败")
		return
	}

	// 转换为响应格式
	type SlotResponse struct {
		ID          uint   `json:"id"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		IsAvailable bool   `json:"is_available"`
	}

	response := make([]SlotResponse, 0, len(slots))
	for _, slot := range slots {
		response = append(response, SlotResponse{
			ID:          slot.ID,
			StartTime:   slot.StartTime,
			EndTime:     slot.EndTime,
			IsAvailable: slot.IsAvailable,
		})
	}

	utils.Success(c, response)
}

// @Summary 批量创建时间段
// @Description 为指定员工在特定日期批量创建时间段（商户端）
// @Tags 商户-时间管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param staffId path int true "员工ID" example(5)
// @Param date query string true "日期 (格式: YYYY-MM-DD)" example("2023-06-15")
// @Param        Authorization header string true "Bearer Token"
// @Param request body []TimeSlotRequest true "时间段列表"
// @Success 200 {object} utils.Response "创建成功"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 500 {object} utils.Response "创建失败"
// @Router /api/merchant/timeslots/{staffId}/batch [post]
func BatchCreateTimeSlots(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	staffID, err := strconv.Atoi(c.Param("staffId"))
	if err != nil {
		utils.BadRequest(c, "无效的员工ID")
		fmt.Println(err)
		return
	}

	date := c.Query("date")
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		utils.BadRequest(c, "无效的日期格式")
		return
	}

	var req []TimeSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// 转换为TimeSlot
	var slots []models.TimeSlot
	for _, r := range req {
		//startTime, err := time.Parse("15:04", r.StartTime)
		//if err != nil {
		//	utils.BadRequest(c, "无效的开始时间")
		//	return
		//}
		//
		//endTime, err := time.Parse("15:04", r.EndTime)
		//if err != nil {
		//	utils.BadRequest(c, "无效的结束时间")
		//	return
		//}
		if r.StartTime > r.EndTime {
			utils.InternalError(c, "创建时间段不能开始时间晚于结束时间")
		}

		slots = append(slots, models.TimeSlot{
			StartTime:   r.StartTime,
			EndTime:     r.EndTime,
			IsAvailable: true,
		})
	}

	if err := models.BatchCreateTimeSlots(merchantID, uint(staffID), parsedDate, slots); err != nil {
		utils.InternalError(c, "创建时间段失败")
		fmt.Println(err)
		return
	}

	utils.Success(c, "时间段创建成功")
}

// @Summary 删除时间段
// @Description 删除指定的时间段（商户端）
// @Tags 商户-时间管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param timeslotId path int true "时间段ID" example(123)
// @Success 200 {object} utils.Response "删除成功"
// @Failure 400 {object} utils.Response "无效的时间段ID"
// @Failure 404 {object} utils.Response "时间段不存在"
// @Failure 500 {object} utils.Response "删除失败"
// @Router /api/merchant/timeslots/{timeslotId} [delete]
func DeleteTimeSlot(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	slotID, err := strconv.Atoi(c.Param("timeslotId"))
	if err != nil {
		utils.BadRequest(c, "无效的时间段ID")
		return
	}

	// 验证时间段属于该商家
	var slot models.TimeSlot
	if err := database.DB.First(&slot, slotID).Error; err != nil || slot.MerchantID != merchantID {
		utils.NotFound(c, "时间段不存在")
		fmt.Println(err)
		return
	}

	if err := models.DeleteTimeSlot(uint(slotID)); err != nil {
		utils.InternalError(c, "删除时间段失败")
		return
	}

	utils.Success(c, "时间段删除成功")
}
