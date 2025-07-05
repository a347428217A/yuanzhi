package routes

import (
	"admin-api/controllers"
	"github.com/gin-gonic/gin"
)

//func InitRouterInternal() *gin.Engine {
//	router := gin.New()
//	router.Use(gin.Recovery())
//
//	// 注意：移除了全局CORS中间件，因为Nginx层已处理跨域
//	// router.Use(middlewares.Cors())
//
//	router.StaticFS(config.Config.ImageSettings.UploadDir, http.Dir(config.Config.ImageSettings.UploadDir))
//	SetupInternalRoutes(router)
//	return router
//}

func SetupInternalRoutes(r *gin.Engine) {
	// 主内部路由组
	internals := r.Group("/api/internal")

	// 商家管理路由组
	merchantGroup := internals.Group("/merchants")
	{
		merchantGroup.POST("", internal.RegisterMerchant)
		merchantGroup.GET("", internal.GetMerchants)            // 新增：获取商家列表
		merchantGroup.GET("/:merchantId", internal.GetMerchant) // 新增：获取单个商家
		merchantGroup.POST("/:merchantId/admins", internal.CreateMerchantAdmin)
		merchantGroup.GET("/:merchantId/admins", internal.GetMerchantAdmins)         // 新增：获取商家管理员列表
		merchantGroup.GET("/:merchantId/admins/:adminId", internal.GetMerchantAdmin) // 新增：获取单个管理员
	}

	//merchantGroup := internals.Group("/merchants")
	//{
	//	merchantGroup.POST("", internal.RegisterMerchant)
	//	merchantGroup.POST("/:merchantId/admins", internal.CreateMerchantAdmin)
	//}

	bannerGroup := internals.Group("/banners")
	{
		bannerGroup.POST("", internal.CreateBanner)
		bannerGroup.GET("", internal.GetBanners)
		bannerGroup.PUT("/:id", internal.UpdateBanner)
		bannerGroup.DELETE("/:id", internal.DeleteBanner)
		bannerGroup.POST("/upload", internal.UploadBannerImage)
	}
}
