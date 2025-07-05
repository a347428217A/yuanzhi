package routes

import (
	"admin-api/controllers/merchant"
	"admin-api/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupMerchantRoutes(r *gin.Engine) {
	// 公共路由组 (无需认证)
	public := r.Group("/api/merchant")
	{
		public.GET("/captcha", merchant.GetCaptcha)
		public.POST("/login", middlewares.LoginGuardMiddleware(), merchant.Login)
	}

	// 认证路由组 (需要商家认证)
	auth := r.Group("/api/merchant")
	auth.Use(middlewares.MerchantAuthMiddleware())
	{
		// 支付管理
		paymentGroup := auth.Group("/payments")
		{
			paymentGroup.GET("", merchant.GetMerchantPayments)
			paymentGroup.GET("/:paymentId", merchant.GetPaymentDetail)
		}

		// 服务类别
		categoryGroup := auth.Group("/categories")
		{
			categoryGroup.POST("", merchant.AddServiceCategory)
			categoryGroup.GET("", merchant.GetServiceCategories)
			categoryGroup.PUT("/:id", merchant.UpdateServiceCategory)
			categoryGroup.DELETE("/:id", merchant.DeleteServiceCategory)
		}

		// 预约管理
		appointmentGroup := auth.Group("/appointments")
		{
			appointmentGroup.GET("", merchant.GetMerchantAppointments)

			// 特定预约操作
			specificAppointment := appointmentGroup.Group("/:appointmentId")
			{
				specificAppointment.PUT("/status", merchant.UpdateAppointmentStatus)
				specificAppointment.POST("/refund", merchant.InitiateRefund)
			}
		}

		// 服务管理
		serviceGroup := auth.Group("/services")
		{
			serviceGroup.GET("", merchant.GetMerchantServices)
			serviceGroup.POST("", merchant.CreateService)

			// 特定服务操作
			specificService := serviceGroup.Group("/:serviceId")
			{
				specificService.PUT("", merchant.UpdateService)
				specificService.DELETE("", merchant.DeleteService)
			}
		}

		// 员工管理
		staffGroup := auth.Group("/staff")
		{
			staffGroup.GET("", merchant.GetMerchantStaff)
			staffGroup.POST("", merchant.CreateStaff)

			// 特定员工操作
			specificStaff := staffGroup.Group("/:staffId")
			{
				specificStaff.PUT("", merchant.UpdateStaff)
				specificStaff.DELETE("", merchant.DeleteStaff)
			}
		}

		// 时间槽管理
		timeslotGroup := auth.Group("/timeslots")
		{
			timeslotGroup.GET("", merchant.GetTimeSlots)
			timeslotGroup.POST("/:staffId/batch", merchant.BatchCreateTimeSlots)

			// 特定时间槽操作
			specificTimeslot := timeslotGroup.Group("/:timeslotId")
			{
				specificTimeslot.DELETE("", merchant.DeleteTimeSlot)
			}
		}

		// 优惠券管理
		couponGroup := auth.Group("/coupons")
		{
			couponGroup.GET("", merchant.GetCouponTemplates)
			couponGroup.POST("", merchant.CreateCouponTemplate)

			// 特定优惠券操作
			specificCoupon := couponGroup.Group("/:couponTemplateId")
			{
				specificCoupon.PUT("", merchant.UpdateCouponTemplate)
				specificCoupon.DELETE("", merchant.DeleteCouponTemplate)
			}
		}

		// 数据统计
		statsGroup := auth.Group("/stats")
		{
			statsGroup.GET("/appointments", merchant.GetAppointmentStats)
			statsGroup.GET("/revenue", merchant.GetRevenueStats)
		}
	}
}
