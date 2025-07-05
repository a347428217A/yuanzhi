package routes

import (
	"admin-api/controllers/customer"
	"admin-api/middlewares"
	"github.com/gin-gonic/gin"
)

// 初始化路由
//func InitRouterCustomer() *gin.Engine {
//	//router := gin.New()
//	//// 跌机恢复
//	//router.Use(gin.Recovery())
//	//router.Use(middlewares.Cors())
//	//router.StaticFS(config.Config.ImageSettings.UploadDir, http.Dir(config.Config.ImageSettings.UploadDir))
//	////router.Use(middlewares.Logger())
//	//SetupCustomerRoutes(router)
//	//return router
//
//	router := gin.New()
//	router.Use(gin.Recovery())
//
//	// 注意：移除了全局CORS中间件，因为Nginx层已处理跨域
//	// router.Use(middlewares.Cors())
//
//	router.StaticFS(config.Config.ImageSettings.UploadDir, http.Dir(config.Config.ImageSettings.UploadDir))
//	SetupCustomerRoutes(router)
//	return router
//}

func SetupCustomerRoutes(r *gin.Engine) {
	// 公共路由
	public := r.Group("/api/customer")
	{
		public.POST("/login", customer.WechatLogin)
		public.GET("/coupons/available", customer.GetAvailableCoupons)
		public.GET("/banners", customer.GetBanners)
		public.POST("/payments/notify", customer.HandlePaymentNotify)

		// 商家相关路由组 - 统一使用 :merchantId
		merchantGroup := public.Group("/merchants")
		{
			merchantGroup.GET("", customer.GetRecommendedMerchants) // GET /api/customer/merchants

			// 特定商家路由组
			specificMerchant := merchantGroup.Group("/:merchantId")
			{
				specificMerchant.GET("", customer.GetMerchantDetail)                       // GET /api/customer/merchants/{merchantId}
				specificMerchant.GET("/categories", customer.GetMerchantServiceCategories) // GET /api/customer/merchants/{merchantId}/categories
				specificMerchant.GET("/services", customer.GetMerchantServices)            // GET /api/customer/merchants/{merchantId}/services
			}
		}

		// 服务相关路由组 - 统一使用 :serviceId
		serviceGroup := public.Group("/services")
		{
			serviceGroup.GET("/:serviceId/staff", customer.GetServiceAvailableStaff) // GET /api/customer/services/{serviceId}/staff
		}
	}

	// 需要认证的路由
	auth := public.Group("")
	auth.Use(middlewares.CustomerAuthMiddleware())
	{
		auth.PUT("/phone", customer.UpdatePhone)
		auth.GET("/profile", customer.GetUserProfile)
		public.POST("/payments/simulate-notify", customer.HandleSimulatePaymentNotify)

		// 新增支付相关路由组
		paymentGroup := auth.Group("/payments")
		{
			paymentGroup.POST("", customer.CreatePayment)        // 创建支付订单
			paymentGroup.GET("/:paymentId", customer.GetPayment) // 查询支付状态
		}

		// 预约路由组 - 统一使用 :appointmentId
		appointmentGroup := auth.Group("/appointments")
		{
			appointmentGroup.GET("", customer.GetUserAppointments) // GET /api/customer/appointments
			appointmentGroup.POST("", customer.CreateAppointment)  // POST /api/customer/appointments

			// 特定预约
			specificAppointment := appointmentGroup.Group("/:appointmentId")
			{
				specificAppointment.GET("", customer.GetAppointmentDetail)     // GET /api/customer/appointments/{appointmentId}
				specificAppointment.PUT("/cancel", customer.CancelAppointment) // PUT /api/customer/appointments/{appointmentId}/cancel

				specificAppointment.POST("/pay", customer.PayForAppointment)
			}
		}

		// 时间槽路由组
		timeslotGroup := auth.Group("/timeslots")
		{
			timeslotGroup.GET("/dates", customer.GetAvailableDates) // GET /api/customer/timeslots/dates
			timeslotGroup.GET("", customer.GetTimeSlots)            // GET /api/customer/timeslots
		}

		// 优惠券路由组 - 统一使用 :couponTemplateId
		couponGroup := auth.Group("/coupons")
		{
			couponGroup.GET("", customer.GetUserCoupons)                       // GET /api/customer/coupons
			couponGroup.POST("/:couponTemplateId/claim", customer.ClaimCoupon) // POST /api/customer/coupons/{couponTemplateId}/claim
		}
	}
}
