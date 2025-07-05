package routes

import (
	"admin-api/controllers/merchant"
	"admin-api/middlewares"
	"github.com/gin-gonic/gin"
)

//func InitRouterMerchant() *gin.Engine {
//	router := gin.New()
//	// 跌机恢复
//	router.Use(gin.Recovery())
//	router.Use(middlewares.Cors())
//	router.StaticFS(config.Config.ImageSettings.UploadDir, http.Dir(config.Config.ImageSettings.UploadDir))
//	//router.Use(middlewares.Logger())
//	SetupMerchantRoutes(router)
//	return router
//}

func SetupMerchantRoutes(r *gin.Engine) {
	// 商家端认证
	auth := r.Group("/api/merchant")
	{
		auth.GET("/captcha", merchant.GetCaptcha)

		// 登录路由添加防护
		auth.POST("/login", middlewares.LoginGuardMiddleware(), merchant.Login)
	}

	// 需要认证的路由
	admin := auth.Group("")
	admin.Use(middlewares.MerchantAuthMiddleware())
	{
		// 新增支付管理路由组
		paymentGroup := admin.Group("/payments")
		{
			paymentGroup.GET("", merchant.GetMerchantPayments)         // 获取商家支付记录
			paymentGroup.GET("/:paymentId", merchant.GetPaymentDetail) // 获取支付详情
		}

		// 服务类别管理路由组 - 新增部分
		categoryGroup := admin.Group("/categories")
		{
			categoryGroup.POST("", merchant.AddServiceCategory)          // POST /api/merchant/categories
			categoryGroup.GET("", merchant.GetServiceCategories)         // GET /api/merchant/categories
			categoryGroup.PUT("/:id", merchant.UpdateServiceCategory)    // PUT /api/merchant/categories/{id}
			categoryGroup.DELETE("/:id", merchant.DeleteServiceCategory) // DELETE /api/merchant/categories/{id}
		}

		// 原有的其他路由组保持不变...
		appointmentGroup := admin.Group("/appointments")
		{
			appointmentGroup.GET("/", merchant.GetMerchantAppointments) // GET /api/merchant/appointments

			// 特定预约
			specificAppointment := appointmentGroup.Group("/:appointmentId")
			{
				specificAppointment.PUT("/status", merchant.UpdateAppointmentStatus) // PUT /api/merchant/appointments/{appointmentId}/status

				specificAppointment.POST("/refund", merchant.InitiateRefund) // 发起退款
			}
		}

		// 服务管理路由组 - 统一使用 :serviceId
		serviceGroup := admin.Group("/services")
		{
			serviceGroup.GET("", merchant.GetMerchantServices) // GET /api/merchant/services
			serviceGroup.POST("", merchant.CreateService)      // POST /api/merchant/services

			// 特定服务
			specificService := serviceGroup.Group("/:serviceId")
			{
				specificService.PUT("", merchant.UpdateService)    // PUT /api/merchant/services/{serviceId}
				specificService.DELETE("", merchant.DeleteService) // DELETE /api/merchant/services/{serviceId}
			}
		}

		// 员工管理路由组 - 统一使用 :staffId
		staffGroup := admin.Group("/staff")
		{
			staffGroup.GET("", merchant.GetMerchantStaff) // GET /api/merchant/staff
			staffGroup.POST("", merchant.CreateStaff)     // POST /api/merchant/staff

			// 特定员工
			specificStaff := staffGroup.Group("/:staffId")
			{
				specificStaff.PUT("", merchant.UpdateStaff)    // PUT /api/merchant/staff/{staffId}
				specificStaff.DELETE("", merchant.DeleteStaff) // DELETE /api/merchant/staff/{staffId}
			}
		}

		// 时间管理路由组 - 统一使用 :timeslotId
		timeslotGroup := admin.Group("/timeslots")
		{
			timeslotGroup.GET("", merchant.GetTimeSlots)                         // GET /api/merchant/timeslots
			timeslotGroup.POST("/:staffId/batch", merchant.BatchCreateTimeSlots) // POST /api/merchant/timeslots/batch

			// 特定时间槽
			specificTimeslot := timeslotGroup.Group("/:timeslotId")
			{
				specificTimeslot.DELETE("", merchant.DeleteTimeSlot) // DELETE /api/merchant/timeslots/{timeslotId}
			}
		}

		// 优惠券管理路由组 - 统一使用 :couponTemplateId
		couponGroup := admin.Group("/coupons")
		{
			couponGroup.GET("", merchant.GetCouponTemplates)    // GET /api/merchant/coupons
			couponGroup.POST("", merchant.CreateCouponTemplate) // POST /api/merchant/coupons

			// 特定优惠券模板
			specificCoupon := couponGroup.Group("/:couponTemplateId")
			{
				specificCoupon.PUT("", merchant.UpdateCouponTemplate)    // PUT /api/merchant/coupons/{couponTemplateId}
				specificCoupon.DELETE("", merchant.DeleteCouponTemplate) // DELETE /api/merchant/coupons/{couponTemplateId}
			}
		}

		// 数据统计路由组
		statsGroup := admin.Group("/stats")
		{
			statsGroup.GET("/appointments", merchant.GetAppointmentStats) // GET /api/merchant/stats/appointments
			statsGroup.GET("/revenue", merchant.GetRevenueStats)          // GET /api/merchant/stats/revenue
		}
	}
}
