package merchant

import (
	"admin-api/database"
	"fmt"
	"strconv"

	"admin-api/models"
	"admin-api/utils"

	"github.com/gin-gonic/gin"
)

// GetMerchantServices 获取商户服务列表
// @Summary      获取商户服务列表
// @Description  根据分类ID获取服务列表
// @Tags         服务管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        category_id query    integer  true  "分类ID"
// @Param        Authorization header string true "Bearer Token"
// @Success      200         {object} utils.Response{data=[]models.Service}  "成功返回服务列表"
// @Failure      500         {object} utils.Response                         "服务器内部错误"
// @Router       /api/merchant/services [get]
func GetMerchantServices(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	categoryID := c.GetUint("category_id")

	services, err := models.GetMerchantServices(merchantID, categoryID)
	if err != nil {
		utils.InternalError(c, "获取服务列表失败")
		return
	}

	utils.Success(c, services)
}

type CreateServiceRequest struct {
	CategoryID  uint   `json:"category_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
	Price       int    `json:"price" binding:"required"`
	Duration    int    `json:"duration" binding:"required"`
}

// CreateService 创建新的商家服务
// @Summary 创建新的商家服务
// @Description 商家创建新的服务项目（需要商家认证）
// @Tags Merchant Services
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "BearerToken" default(Bearer <token>)
// @Param request body CreateServiceRequest true "服务信息"
// @Success 200 {object} utils.Response{data=models.Service} "创建成功"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 401 {object} utils.Response "认证失败"
// @Failure 500 {object} utils.Response "内部错误"
// @Router /api/merchant/services [post]
func CreateService(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")

	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		utils.BadRequest(c, "参数错误")
		return
	}

	service := models.Service{
		MerchantID:  merchantID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		Price:       req.Price,
		Duration:    req.Duration,
		IsActive:    true,
	}

	if err := models.CreateService(&service); err != nil {
		fmt.Println(err)
		utils.InternalError(c, "创建服务失败")
		return
	}

	utils.Success(c, service)
}

type UpdateServiceRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CoverImage  string  `json:"cover_image"`
	Price       float64 `json:"price"`
	Duration    int     `json:"duration"`
	IsActive    bool    `json:"is_active"`
}

// UpdateService 更新商家服务
// @Summary      更新服务信息
// @Description  更新商家服务的详细信息
// @Tags         商家服务管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        serviceId path int true "服务ID"
// @Param        body body UpdateServiceRequest true "服务更新请求"
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {object}  utils.Response "更新成功"
// @Failure      400  {object}  utils.Response "无效的服务ID | 参数错误"
// @Failure      401  {object}  utils.Response "认证失败"
// @Failure      404  {object}  utils.Response "服务不存在"
// @Failure      500  {object}  utils.Response "更新服务失败"
// @Router       /api/merchant/services/{serviceId} [put]
func UpdateService(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	serviceID, err := strconv.Atoi(c.Param("serviceId"))
	if err != nil {
		utils.BadRequest(c, "无效的服务ID")
		fmt.Println(err)
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// 验证服务属于该商家
	var service models.Service
	if err := database.DB.First(&service, serviceID).Error; err != nil || service.MerchantID != merchantID {
		utils.NotFound(c, "服务不存在")
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"cover_image": req.CoverImage,
		"price":       req.Price,
		"duration":    req.Duration,
		"is_active":   req.IsActive,
	}

	if err := models.UpdateService(uint(serviceID), updates); err != nil {
		utils.InternalError(c, "更新服务失败")
		return
	}

	utils.Success(c, "更新成功")
}

// DeleteService 删除商家服务
// @Summary      删除服务
// @Description  删除商家服务（软删除或硬删除）
// @Tags         商家服务管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        serviceId path int true "服务ID" example(123)
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {object}  utils.Response "删除成功"
// @Failure      400  {object}  utils.Response "无效的服务ID"
// @Failure      401  {object}  utils.Response "认证失败"
// @Failure      404  {object}  utils.Response "服务不存在"
// @Failure      500  {object}  utils.Response "删除服务失败"
// @Router       /api/merchant/services/{serviceId} [delete]
func DeleteService(c *gin.Context) {
	merchantID := c.GetUint("merchant_id")
	serviceID, err := strconv.Atoi(c.Param("serviceId"))
	if err != nil {
		utils.BadRequest(c, "无效的服务ID")
		return
	}

	// 验证服务属于该商家
	var service models.Service
	if err := database.DB.First(&service, serviceID).Error; err != nil || service.MerchantID != merchantID {
		utils.NotFound(c, "服务不存在")
		return
	}

	if err := models.DeleteService(uint(serviceID)); err != nil {
		utils.InternalError(c, "删除服务失败")
		return
	}

	utils.Success(c, "删除成功")
}
