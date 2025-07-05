package routes

import (
	"admin-api/controllers/customer"
	"admin-api/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupCustomerRoutes(r *gin.Engine) {
	// 公共路由组 (无需认证)
	public := r.Group("/api/customer")
	{
		// 认证相关
		public.POST("/login", customer.WechatLogin)

		// 优惠券
		public.GET("/coupons/available", customer.GetAvailableCoupons)

		// 横幅
		public.GET("/banners", customer.GetBanners)

		// 支付回调
		public.POST("/payments/notify", customer.HandlePaymentNotify)
		public.POST("/payments/simulate-notify", customer.HandleSimulatePaymentNotify)

		// 商家相关
		merchantGroup := public.Group("/merchants")
		{
			merchantGroup.GET("", customer.GetRecommendedMerchants)

			// 特定商家
			specificMerchant := merchantGroup.Group("/:merchantId")
			{
				specificMerchant.GET("", customer.GetMerchantDetail)
				specificMerchant.GET("/categories", customer.GetMerchantServiceCategories)
				specificMerchant.GET("/services", customer.GetMerchantServices)
			}
		}

		// 服务相关
		serviceGroup := public.Group("/services")
		{
			serviceGroup.GET("/:serviceId/staff", customer.GetServiceAvailableStaff)
		}
	}

	// 认证路由组 (需要客户认证)
	auth := r.Group("/api/customer")
	auth.Use(middlewares.CustomerAuthMiddleware())
	{
		// 用户资料
		auth.PUT("/phone", customer.UpdatePhone)
		auth.GET("/profile", customer.GetUserProfile)

		// 支付管理
		paymentGroup := auth.Group("/payments")
		{
			paymentGroup.POST("", customer.CreatePayment)
			paymentGroup.GET("/:paymentId", customer.GetPayment)
		}

		// 预约管理
		appointmentGroup := auth.Group("/appointments")
		{
			appointmentGroup.GET("", customer.GetUserAppointments)
			appointmentGroup.POST("", customer.CreateAppointment)

			// 特定预约操作
			specificAppointment := appointmentGroup.Group("/:appointmentId")
			{
				specificAppointment.GET("", customer.GetAppointmentDetail)
				specificAppointment.PUT("/cancel", customer.CancelAppointment)
				specificAppointment.POST("/pay", customer.PayForAppointment)
			}
		}

		// 时间槽
		timeslotGroup := auth.Group("/timeslots")
		{
			timeslotGroup.GET("/dates", customer.GetAvailableDates)
			timeslotGroup.GET("", customer.GetTimeSlots)
		}

		// 优惠券
		couponGroup := auth.Group("/coupons")
		{
			couponGroup.GET("", customer.GetUserCoupons)
			couponGroup.POST("/:couponTemplateId/claim", customer.ClaimCoupon)
		}
	}
}
