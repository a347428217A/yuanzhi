package merchant

import (
	"strconv"

	"admin-api/models"
	"admin-api/utils"
	"github.com/gin-gonic/gin"
)

type AddServiceCategoryRequest struct {
	Name string `json:"name" binding:"required"`
	Sort int    `json:"sort"` // 可选，默认为0
}

// @Summary 添加服务类别
// @Description 商家管理员添加新的服务类别
// @Tags 服务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "BearerToken" default(Bearer <token>)
// @Param body body AddServiceCategoryRequest true "类别信息"
// @Success 201 {object} models.ServiceCategory "成功返回创建的类别"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 401 {object} utils.Response "未授权"
// @Failure 403 {object} utils.Response "权限不足"
// @Failure 409 {object} utils.Response "类别名称已存在"
// @Failure 500 {object} utils.Response "添加类别失败"
// @Router /api/merchant/categories [post]
func AddServiceCategory(c *gin.Context) {
	// 从认证信息中获取商家ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	merchantIDUint, ok := merchantID.(uint)
	if !ok {
		utils.InternalError(c, "商家ID格式错误")
		return
	}

	var req AddServiceCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 调用模型层创建服务类别
	category, err := models.CreateServiceCategory(merchantIDUint, req.Name, req.Sort)
	if err != nil {
		switch {
		case err.Error() == "商家不存在":
			utils.NotFound(c, err.Error())
		case err.Error() == "该类别名称已存在":
			utils.Error(c, 400, err.Error())
		default:
			utils.InternalError(c, "添加类别失败: "+err.Error())
		}
		return
	}

	// 返回创建成功的响应
	utils.Success(c, category)
}

// @Summary 获取服务类别列表
// @Description 获取当前商家的所有服务类别
// @Tags 服务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "BearerToken" default(Bearer <token>)
// @Success 200 {array} models.ServiceCategory "成功返回类别列表"
// @Failure 401 {object} utils.Response "未授权"
// @Failure 500 {object} utils.Response "获取类别失败"
// @Router /api/merchant/categories [get]
func GetServiceCategories(c *gin.Context) {
	// 从认证信息中获取商家ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	merchantIDUint, ok := merchantID.(uint)
	if !ok {
		utils.InternalError(c, "商家ID格式错误")
		return
	}

	// 获取服务类别
	categories, err := models.GetCategoriesByMerchant(merchantIDUint)
	if err != nil {
		utils.InternalError(c, "获取类别失败: "+err.Error())
		return
	}

	utils.Success(c, categories)
}

// 定义请求结构体
type UpdateServiceCategoryRequest struct {
	Name string `json:"name"`
	Sort *int   `json:"sort"` // 使用指针类型区分0值和未传值
}

// @Summary 更新服务类别
// @Description 更新指定服务类别信息
// @Tags 服务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "BearerToken" default(Bearer <token>)
// @Param id path int true "类别ID"
// @Param body body UpdateServiceCategoryRequest true "更新信息"
// @Success 200 {object} models.ServiceCategory "成功返回更新的类别"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 401 {object} utils.Response "未授权"
// @Failure 404 {object} utils.Response "类别不存在"
// @Failure 409 {object} utils.Response "类别名称已存在"
// @Failure 500 {object} utils.Response "更新类别失败"
// @Router /api/merchant/categories/{id} [put]
func UpdateServiceCategory(c *gin.Context) {
	// 从认证信息中获取商家ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	merchantIDUint, ok := merchantID.(uint)
	if !ok {
		utils.InternalError(c, "商家ID格式错误")
		return
	}

	// 获取类别ID
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的类别ID")
		return
	}

	var req UpdateServiceCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 处理排序值
	sort := -1 // 默认值表示不更新
	if req.Sort != nil {
		sort = *req.Sort
	}

	// 调用模型层更新类别
	category, err := models.UpdateCategory(uint(categoryID), merchantIDUint, req.Name, sort)
	if err != nil {
		switch {
		case err.Error() == "服务类别不存在":
			utils.NotFound(c, err.Error())
		case err.Error() == "该类别名称已被使用":
			utils.Error(c, 400, err.Error())
		default:
			utils.InternalError(c, "更新类别失败: "+err.Error())
		}
		return
	}

	utils.Success(c, category)
}

// @Summary 删除服务类别
// @Description 删除指定服务类别
// @Tags 服务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "BearerToken" default(Bearer <token>)
// @Param id path int true "类别ID"
// @Success 200 {string} string "删除成功"
// @Failure 400 {object} utils.Response "无效的类别ID"
// @Failure 401 {object} utils.Response "未授权"
// @Failure 403 {object} utils.Response "类别下存在服务，无法删除"
// @Failure 404 {object} utils.Response "类别不存在"
// @Failure 500 {object} utils.Response "删除类别失败"
// @Router /api/merchant/categories/{id} [delete]
func DeleteServiceCategory(c *gin.Context) {
	// 从认证信息中获取商家ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	merchantIDUint, ok := merchantID.(uint)
	if !ok {
		utils.InternalError(c, "商家ID格式错误")
		return
	}

	// 获取类别ID
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的类别ID")
		return
	}

	// 调用模型层删除类别
	if err := models.DeleteCategory(uint(categoryID), merchantIDUint); err != nil {
		switch {
		case err.Error() == "服务类别不存在":
			utils.NotFound(c, err.Error())
		case err.Error() == "该类别下存在服务，无法删除":
			utils.Forbidden(c, err.Error())
		default:
			utils.InternalError(c, "删除类别失败: "+err.Error())
		}
		return
	}

	utils.Success(c, "类别删除成功")
}
