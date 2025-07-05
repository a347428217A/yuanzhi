package merchant

import (
	"admin-api/database"
	"fmt"
	"strconv"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

// GetMerchantStaff 获取商家员工列表
// @Summary      获取员工列表
// @Description  获取当前商家的所有员工列表
// @Tags         商家员工管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {object}  utils.Response{data=[]models.Staff} "员工列表"
// @Failure      401  {object}  utils.Response "认证失败"
// @Failure      500  {object}  utils.Response "获取员工列表失败"
// @Router       /api/merchant/staff [get]
func GetMerchantStaff(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")

	staff, err := models.GetMerchantStaff(merchantID)
	if err != nil {
		utils.InternalError(c, "获取员工列表失败")
		fmt.Println(err)
		return
	}

	utils.Success(c, staff)
}

type CreateStaffRequest struct {
	Name        string `json:"name" binding:"required"`
	Avatar      string `json:"avatar"`
	Position    string `json:"position"`
	Description string `json:"description"`
	Specialties string `json:"specialties"`
}

// CreateStaff 创建新员工
// @Summary      创建员工
// @Description  为当前商家创建新员工
// @Tags         商家员工管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body CreateStaffRequest true "员工信息"
// @Param        Authorization header string true "Bearer Token"
// @Success      201  {object}  utils.Response{data=models.Staff} "员工创建成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "认证失败"
// @Failure      500  {object}  utils.Response "创建员工失败"
// @Router       /api/merchant/staff [post]
func CreateStaff(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	var req CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	staff := models.Staff{
		MerchantID:  merchantID,
		Name:        req.Name,
		Avatar:      req.Avatar,
		Position:    req.Position,
		Description: req.Description,
		Specialties: req.Specialties,
		IsActive:    true,
	}

	if err := models.CreateStaff(&staff); err != nil {
		utils.InternalError(c, "创建员工失败")
		return
	}

	utils.Success(c, staff)
}

type UpdateStaffRequest struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Position    string `json:"position"`
	Description string `json:"description"`
	Specialties string `json:"specialties"`
	IsActive    bool   `json:"is_active"`
}

// UpdateStaff 更新员工信息
// @Summary      更新员工
// @Description  更新员工信息
// @Tags         商家员工管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        staffId path int true "员工ID" example(123)
// @Param        body body UpdateStaffRequest true "员工更新信息"
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {object}  utils.Response{data=models.Staff} "更新成功"
// @Failure      400  {object}  utils.Response "无效的员工ID | 参数错误"
// @Failure      401  {object}  utils.Response "认证失败"
// @Failure      404  {object}  utils.Response "员工不存在"
// @Failure      500  {object}  utils.Response "更新员工失败"
// @Router       /api/merchant/staff/{staffId} [put]
func UpdateStaff(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	staffID, err := strconv.Atoi(c.Param("staffId"))
	if err != nil {
		utils.BadRequest(c, "无效的员工ID")
		return
	}

	var req UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// 验证员工属于该商家
	var staff models.Staff
	if err := database.DB.First(&staff, staffID).Error; err != nil || staff.MerchantID != merchantID {
		utils.NotFound(c, "员工不存在")
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"avatar":      req.Avatar,
		"position":    req.Position,
		"description": req.Description,
		"specialties": req.Specialties,
		"is_active":   req.IsActive,
	}

	if err := models.UpdateStaff(uint(staffID), updates); err != nil {
		utils.InternalError(c, "更新员工失败")
		return
	}

	utils.Success(c, "更新成功")
}

// DeleteStaff 删除员工
// @Summary      删除员工
// @Description  删除员工（软删除或硬删除）
// @Tags         商家员工管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        staffId path int true "员工ID" example(123)
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {object}  utils.Response "员工删除成功"
// @Failure      400  {object}  utils.Response "无效的员工ID"
// @Failure      401  {object}  utils.Response "认证失败"
// @Failure      404  {object}  utils.Response "员工不存在"
// @Failure      500  {object}  utils.Response "删除员工失败"
// @Router       /api/merchant/staff/{staffId} [delete]
func DeleteStaff(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	staffID, err := strconv.Atoi(c.Param("staffId"))
	if err != nil {
		utils.BadRequest(c, "无效的员工ID")
		return
	}

	// 验证员工属于该商家
	var staff models.Staff
	if err := database.DB.First(&staff, staffID).Error; err != nil || staff.MerchantID != merchantID {
		utils.NotFound(c, "员工不存在")
		return
	}

	if err := models.DeleteStaff(uint(staffID)); err != nil {
		utils.InternalError(c, "删除员工失败")
		fmt.Println(err)
		return
	}

	utils.Success(c, "员工删除成功")
}
