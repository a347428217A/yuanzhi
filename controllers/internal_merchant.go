package internal

import (
	"admin-api/config"
	"errors"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"

	"admin-api/models"
	"admin-api/utils"
	"github.com/gin-gonic/gin"
)

type RegisterMerchantRequest struct {
	Name         string `json:"name" binding:"required"`
	Address      string `json:"address" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Description  string `json:"description"`
	Logo         string `json:"logo"`
	BusinessHour string `json:"business_hour"`
}

// @Summary 注册新商家
// @Description 内部接口：注册一个新的商家账户（不对普通用户开放）
// @Tags 内部管理
// @Accept json
// @Produce json
// @Param body body RegisterMerchantRequest true "商家注册信息"
// @Success 201 {object} models.Merchant "成功返回创建的商家信息"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 409 {object} utils.Response "商家名称或电话已存在"
// @Failure 500 {object} utils.Response "创建商家失败"
// @Router /api/internal/merchants [post] // 修正路径
func RegisterMerchant(c *gin.Context) {
	// 定义请求结构体

	var req RegisterMerchantRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 可选：添加额外验证
	if len(req.Phone) < 6 || len(req.Phone) > 20 {
		utils.BadRequest(c, "联系电话格式无效")
		return
	}

	// 调用模型层创建商家
	merchant, err := models.CreateMerchant(
		req.Name,
		req.Address,
		req.Phone,
		req.Description,
		req.Logo,
		req.BusinessHour,
	)

	if err != nil {
		// 根据错误类型返回不同响应
		if strings.Contains(err.Error(), "已存在") {
			utils.Error(c, 400, "创建商家失败: "+err.Error())
		} else {
			utils.InternalError(c, "创建商家失败: "+err.Error())
		}
		return
	}

	// 返回创建成功的响应
	utils.Success(c, merchant)
}

type CreateMerchantAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin staff"`
}

// @Summary 创建商家管理员
// @Description 内部接口：为指定商家创建管理员账户
// @Tags 内部管理
// @Accept json
// @Produce json
// @Param merchantId path int true "商家ID"
// @Param body body CreateMerchantAdminRequest true "管理员信息"
// @Success 201 {object} models.MerchantAdmin "成功返回创建的管理员信息"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 404 {object} utils.Response "商家不存在"
// @Failure 409 {object} utils.Response "用户名已被使用"
// @Failure 500 {object} utils.Response "创建管理员失败"
// @Router /api/internal/merchants/{merchantId}/admins [post] // 修正路径
func CreateMerchantAdmin(c *gin.Context) {
	// 获取路径参数
	merchantID, err := strconv.Atoi(c.Param("merchantId"))
	if err != nil {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	var req CreateMerchantAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	//密码强度验证
	if err := utils.ValidatePasswordPolicy(req.Password); err != nil {
		utils.BadRequest(c, "密码强度不足: "+err.Error())
		return
	}

	hashedPassword, err := utils.GeneratePasswordHash(req.Password)
	if err != nil {
		utils.InternalError(c, "密码加密失败: "+err.Error())
		return
	}

	// 调用模型层创建商家管理员
	admin, err := models.CreateMerchantAdmin(
		uint(merchantID),
		req.Username,
		hashedPassword,
		req.Role,
	)
	if err != nil {
		switch {
		case err.Error() == "商家不存在":
			utils.NotFound(c, err.Error())
		case err.Error() == "用户名已被使用":
			utils.Error(c, 400, err.Error())
		case err.Error() == "无效的角色类型":
			utils.BadRequest(c, err.Error())
		default:
			utils.InternalError(c, "创建管理员失败: "+err.Error())
		}
		return
	}

	// 返回创建成功的响应
	utils.Success(c, admin)
}

// 查询参数结构体
type GetMerchantsQuery struct {
	Name  string `form:"name"`                          // 商家名称模糊查询
	Phone string `form:"phone"`                         // 联系电话精确查询
	Page  int    `form:"page" binding:"min=1"`          // 页码
	Limit int    `form:"limit" binding:"min=1,max=100"` // 每页数量
}

// @Summary 获取商家列表
// @Description 内部接口：获取商家列表（支持分页和搜索）
// @Tags 内部管理
// @Accept json
// @Produce json
// @Param name query string false "商家名称(模糊匹配)"
// @Param phone query string false "联系电话(精确匹配)"
// @Param page query int true "页码" default(1)
// @Param limit query int true "每页数量" default(10)
// @Success 200 {object} utils.PaginatedResponse{data=[]models.Merchant} "成功返回商家列表"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/internal/merchants [get]
func GetMerchants(c *gin.Context) {
	var query GetMerchantsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 10
	}

	// 调用模型层
	merchants, total, err := models.GetMerchants(query.Name, query.Phone, query.Page, query.Limit)
	if err != nil {
		utils.InternalError(c, "获取商家列表失败: "+err.Error())
		return
	}

	// 返回分页响应
	utils.PaginatedSuccess(c, merchants, total, query.Page, query.Limit)
}

// @Summary 获取单个商家
// @Description 内部接口：获取指定商家的详细信息
// @Tags 内部管理
// @Accept json
// @Produce json
// @Param merchantId path int true "商家ID"
// @Success 200 {object} models.Merchant "成功返回商家信息"
// @Failure 400 {object} utils.Response "无效的商家ID"
// @Failure 404 {object} utils.Response "商家不存在"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/internal/merchants/{merchantId} [get]
func GetMerchant(c *gin.Context) {
	merchantID, err := strconv.Atoi(c.Param("merchantId"))
	if err != nil || merchantID <= 0 {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	merchant, err := models.GetMerchantByID(uint(merchantID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.NotFound(c, "商家不存在")
		} else {
			utils.InternalError(c, "获取商家信息失败: "+err.Error())
		}
		return
	}

	utils.Success(c, merchant)
}

// 管理员查询参数
type GetAdminsQuery struct {
	Username string `form:"username"`                      // 用户名模糊查询
	Role     string `form:"role"`                          // 角色过滤
	Page     int    `form:"page" binding:"min=1"`          // 页码
	Limit    int    `form:"limit" binding:"min=1,max=100"` // 每页数量
}

// @Summary 获取商家管理员列表
// @Description 内部接口：获取指定商家的管理员列表
// @Tags 内部管理
// @Accept json
// @Produce json
// @Param merchantId path int true "商家ID"
// @Param username query string false "用户名(模糊匹配)"
// @Param role query string false "角色过滤"
// @Param page query int true "页码" default(1)
// @Param limit query int true "每页数量" default(10)
// @Success 200 {object} utils.PaginatedResponse{data=[]models.MerchantAdmin} "成功返回管理员列表"
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 404 {object} utils.Response "商家不存在"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/internal/merchants/{merchantId}/admins [get]
func GetMerchantAdmins(c *gin.Context) {
	merchantID, err := strconv.Atoi(c.Param("merchantId"))
	if err != nil || merchantID <= 0 {
		utils.BadRequest(c, "无效的商家ID")
		return
	}

	var query GetAdminsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 10
	}

	// 调用模型层
	admins, total, err := models.GetMerchantAdmins(
		uint(merchantID),
		query.Username,
		query.Role,
		query.Page,
		query.Limit,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.NotFound(c, "商家不存在")
		} else {
			utils.InternalError(c, "获取管理员列表失败: "+err.Error())
		}
		return
	}

	// 返回分页响应
	utils.PaginatedSuccess(c, admins, total, query.Page, query.Limit)
}

// @Summary 获取单个管理员
// @Description 内部接口：获取指定商家管理员的详细信息
// @Tags 内部管理
// @Accept json
// @Produce json
// @Param merchantId path int true "商家ID"
// @Param adminId path int true "管理员ID"
// @Success 200 {object} models.MerchantAdmin "成功返回管理员信息"
// @Failure 400 {object} utils.Response "无效的ID"
// @Failure 404 {object} utils.Response "管理员不存在"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/internal/merchants/{merchantId}/admins/{adminId} [get]
func GetMerchantAdmin(c *gin.Context) {
	merchantID, err1 := strconv.Atoi(c.Param("merchantId"))
	adminID, err2 := strconv.Atoi(c.Param("adminId"))

	if err1 != nil || merchantID <= 0 || err2 != nil || adminID <= 0 {
		utils.BadRequest(c, "无效的ID参数")
		return
	}

	admin, err := models.GetMerchantAdminByID(uint(adminID), uint(merchantID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.NotFound(c, "管理员不存在")
		} else {
			utils.InternalError(c, "获取管理员信息失败: "+err.Error())
		}
		return
	}

	// 敏感信息处理
	admin.Password = "" // 清除密码字段

	utils.Success(c, admin)
}

type CreateBannerRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url" binding:"required"`
	LinkURL     string    `json:"link_url"`
	Position    string    `json:"position" binding:"required"` // home, category
	Platform    string    `json:"platform" binding:"required"` // all, weapp, h5, app
	Sort        int       `json:"sort"`
	Status      int       `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

type UpdateBannerRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	LinkURL     string    `json:"link_url"`
	Position    string    `json:"position"`
	Platform    string    `json:"platform"`
	Sort        int       `json:"sort"`
	Status      int       `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

// @Summary 创建轮播图
// @Description 内部接口：创建轮播图
// @Tags 轮播图管理
// @Accept json
// @Produce json
// @Param input body CreateBannerRequest true "轮播图信息"
// @Success 201 {object} models.Banner
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 500 {object} utils.Response "创建失败"
// @Router /api/internal/banners [post]
func CreateBanner(c *gin.Context) {
	var req CreateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.StartTime.IsZero() {
		req.StartTime = time.Now()
	}
	if req.EndTime.IsZero() {
		req.EndTime = time.Now().AddDate(0, 1, 0) // 默认一个月后
	}
	if req.Sort == 0 {
		req.Sort = 99 // 默认排序值
	}

	banner := models.Banner{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		LinkURL:     req.LinkURL,
		Position:    req.Position,
		Platform:    req.Platform,
		Sort:        req.Sort,
		Status:      req.Status,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	if err := models.CreateBanner(&banner); err != nil {
		utils.InternalError(c, "创建轮播图失败: "+err.Error())
		return
	}

	utils.Success(c, banner)
}

// @Summary 更新轮播图
// @Description 内部接口：更新轮播图
// @Tags 轮播图管理
// @Accept json
// @Produce json
// @Param id path int true "轮播图ID"
// @Param input body UpdateBannerRequest true "轮播图信息"
// @Success 200 {object} models.Banner
// @Failure 400 {object} utils.Response "参数错误"
// @Failure 404 {object} utils.Response "轮播图不存在"
// @Failure 500 {object} utils.Response "更新失败"
// @Router /api/internal/banners/{id} [put]
func UpdateBanner(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的轮播图ID")
		return
	}

	var req UpdateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	banner, err := models.GetBannerByID(uint(id))
	if err != nil {
		utils.NotFound(c, "轮播图不存在")
		return
	}

	// 更新字段
	if req.Title != "" {
		banner.Title = req.Title
	}
	if req.Description != "" {
		banner.Description = req.Description
	}
	if req.ImageURL != "" {
		banner.ImageURL = req.ImageURL
	}
	if req.LinkURL != "" {
		banner.LinkURL = req.LinkURL
	}
	if req.Position != "" {
		banner.Position = req.Position
	}
	if req.Platform != "" {
		banner.Platform = req.Platform
	}
	if req.Sort != 0 {
		banner.Sort = req.Sort
	}
	if req.Status != 0 {
		banner.Status = req.Status
	}
	if !req.StartTime.IsZero() {
		banner.StartTime = req.StartTime
	}
	if !req.EndTime.IsZero() {
		banner.EndTime = req.EndTime
	}

	if err := models.UpdateBanner(banner); err != nil {
		utils.InternalError(c, "更新轮播图失败: "+err.Error())
		return
	}

	utils.Success(c, banner)
}

// @Summary 删除轮播图
// @Description 内部接口：删除轮播图
// @Tags 轮播图管理
// @Accept json
// @Produce json
// @Param id path int true "轮播图ID"
// @Success 200 {object} utils.Response "删除成功"
// @Failure 400 {object} utils.Response "无效的轮播图ID"
// @Failure 500 {object} utils.Response "删除失败"
// @Router /api/internal/banners/{id} [delete]
func DeleteBanner(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的轮播图ID")
		return
	}

	if err := models.DeleteBanner(uint(id)); err != nil {
		utils.InternalError(c, "删除轮播图失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "轮播图删除成功"})
}

// @Summary 获取轮播图列表
// @Description 内部接口：获取轮播图列表
// @Tags 轮播图管理
// @Accept json
// @Produce json
// @Param position query string false "位置筛选"
// @Param platform query string false "平台筛选"
// @Param status query int false "状态筛选(0-禁用,1-启用)"
// @Success 200 {array} models.Banner
// @Router /api/internal/banners [get]
func GetBanners(c *gin.Context) {
	position := c.Query("position")
	platform := c.Query("platform")
	statusStr := c.Query("status")

	var status int
	if statusStr != "" {
		status, _ = strconv.Atoi(statusStr)
	}

	banners, err := models.GetBanners(position, platform, status)
	if err != nil {
		utils.InternalError(c, "获取轮播图失败: "+err.Error())
		return
	}

	utils.Success(c, banners)
}

// @Summary 上传轮播图图片
// @Description 内部接口：上传轮播图图片
// @Tags 轮播图管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} utils.Response "上传成功"
// @Failure 400 {object} utils.Response "文件错误"
// @Failure 500 {object} utils.Response "上传失败"
// @Router /api/internal/banners/upload [post]
func UploadBannerImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "获取文件失败: "+err.Error())
		return
	}

	// 验证文件类型
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		utils.BadRequest(c, "不支持的文件类型")
		return
	}

	// 生成唯一文件名
	newFilename := utils.GenerateFilename(file.Filename)
	filePath := config.Config.ImageSettings.UploadDir + "/banners/" + newFilename

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.InternalError(c, "保存文件失败: "+err.Error())
		return
	}

	// 返回图片URL
	imageURL := config.Config.Server.Address + "/uploads/banners/" + newFilename
	utils.Success(c, gin.H{"image_url": imageURL})
}
