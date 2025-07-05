package main

import (
	"admin-api/database"
	_ "admin-api/docs"
	"admin-api/middlewares"
	"admin-api/routes"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	//"fmt"
	"context"
	"github.com/gin-gonic/gin"
	"os"
)

// @title Appointment System API
// @version 1.0
// @description 预约系统API文档
// @host https://user-go-api-171613-8-1367826874.sh.run.tcloudbase.com
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// 1. 从环境变量获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("⚠️ 使用默认端口: %s", port)
	} else {
		log.Printf("✅ 使用环境变量端口: %s", port)
	}

	// 2. 创建唯一的路由器实例
	router := gin.Default()

	router.Use(middlewares.Cors())

	// 3. 添加健康检查端点（必须放在最前面）
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "admin-api"})
	})

	router.GET("/health", func(c *gin.Context) {
		// 添加数据库健康检查
		dbStatus := "ok"
		if err := database.DB.Exec("SELECT 1").Error; err != nil {
			dbStatus = "error: " + err.Error()
		}
		c.JSON(200, gin.H{
			"status":   "ok",
			"database": dbStatus,
		})
	})

	//router.Use(func(c *gin.Context) {
	//	if c.Request.Header.Get("X-Forwarded-Proto") == "http" {
	//		target := "https://" + c.Request.Host + c.Request.URL.Path
	//		c.Redirect(http.StatusMovedPermanently, target)
	//		return
	//	}
	//	c.Next()
	//})

	//router.Use(func(c *gin.Context) {
	//	// 跳过所有 API 路由（只重定向非 API 请求）
	//	if strings.HasPrefix(c.Request.URL.Path, "/api") {
	//		c.Next()
	//		return
	//	}
	//
	//	if c.Request.Header.Get("X-Forwarded-Proto") == "http" {
	//		target := "https://" + c.Request.Host + c.Request.URL.Path
	//		c.Redirect(http.StatusMovedPermanently, target)
	//		return
	//	}
	//	c.Next()
	//})

	// 4. 初始化数据库
	database.InitDB()
	log.Println("✅ 数据库初始化完成")

	// 5. 设置路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	routes.SetupCustomerRoutes(router) // 使用同一个router实例
	routes.SetupMerchantRoutes(router)
	routes.SetupInternalRoutes(router)

	// 6. 创建HTTP服务器
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router, // 使用同一个router
		// 超时设置保持不变...
	}

	// 7. 启动服务器
	go func() {
		log.Printf("🚀 服务启动在 http://0.0.0.0:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	log.Println("===== 注册的路由 =====")
	for _, route := range router.Routes() {
		log.Printf("%-6s %s", route.Method, route.Path)
	}
	log.Println("======================")

	// 8. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 接收到关闭信号，开始优雅关闭...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ 服务关闭失败: %v", err)
	}
	log.Println("✅ 服务已优雅停止")
}

//func main() {
//	// 1. 从环境变量获取端口
//	port := os.Getenv("PORT")
//	if port == "" {
//		port = "80"
//	}
//
//	r := gin.Default()
//
//	// 仅保留健康检查
//	r.GET("/", func(c *gin.Context) {
//		c.JSON(200, gin.H{"status": "ok"})
//	})
//
//	//log.Printf("🚀 简易服务启动在 :%s", port)
//	//if err := r.Run(":" + port); err != nil {
//	//	log.Fatalf("❌ 服务启动失败: %v", err)
//	//}
//
//	database.InitDB()
//	mainRouter := gin.Default()
//	mainRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
//	routes.SetupCustomerRoutes(mainRouter)
//	routes.SetupMerchantRoutes(mainRouter)
//	routes.SetupInternalRoutes(mainRouter)
//
//	server := &http.Server{
//		Addr:         ":" + port, // 关键修改：使用环境变量端口
//		Handler:      mainRouter,
//		ReadTimeout:  15 * time.Second,
//		WriteTimeout: 30 * time.Second,
//		IdleTimeout:  60 * time.Second,
//	}
//
//	go func() {
//		log.Printf("🚀 服务启动在 http://0.0.0.0:%s", port)
//		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
//			log.Fatalf("❌ 服务器启动失败: %v", err)
//		}
//	}()
//
//	quit := make(chan os.Signal, 1)
//	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//	<-quit
//	log.Println("🛑 接收到关闭信号，开始优雅关闭...")
//
//	//ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
//	//defer cancel()
//	//
//	//if err := server.Shutdown(ctx); err != nil {
//	//	log.Fatalf("❌ 服务器强制关闭: %v", err)
//	//}
//	//log.Println("✅ 服务器已优雅关闭")
//
//	//port := os.Getenv("PORT")
//	//if port == "" {
//	//	port = "80" // 本地开发默认端口
//	//	log.Printf("⚠️ PORT环境变量未设置，使用默认端口: %s", port)
//	//} else {
//	//	log.Printf("✅ 使用环境变量PORT: %s", port)
//	//}
//	//
//	//// 2. 初始化数据库（确保database.InitDB()内部使用环境变量）
//	//database.InitDB()
//	//
//	//// 3. 初始化Redis（确保redis.SetupRedisDb()内部使用环境变量）
//	////err := redis.SetupRedisDb()
//	////if err != nil {
//	////	log.Fatalf("❌ Redis初始化失败: %v", err)
//	////} else {
//	////	log.Println("✅ Redis初始化成功")
//	////}
//	//
//	//// 4. 创建路由
//	//mainRouter := gin.Default()
//	//
//	//// 5. 添加Swagger路由
//	//mainRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
//	//log.Printf("🔍 Swagger UI 可用: http://0.0.0.0:%s/swagger/index.html", port)
//	//
//	//// 6. 设置路由
//	//routes.SetupCustomerRoutes(mainRouter)
//	//routes.SetupMerchantRoutes(mainRouter)
//	//routes.SetupInternalRoutes(mainRouter)
//	//
//	//// 7. 添加健康检查端点（云托管需要）
//	//mainRouter.GET("/health", func(c *gin.Context) {
//	//	c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
//	//})
//	//
//	//// 8. 添加根路径健康检查（云托管默认检查）
//	//mainRouter.GET("/", func(c *gin.Context) {
//	//	c.JSON(200, gin.H{"status": "ok"})
//	//})
//	//
//	//// 9. 启动HTTP服务器（使用环境变量端口）
//	//server := &http.Server{
//	//	Addr:         ":" + port, // 关键修改：使用环境变量端口
//	//	Handler:      mainRouter,
//	//	ReadTimeout:  15 * time.Second,
//	//	WriteTimeout: 30 * time.Second,
//	//	IdleTimeout:  60 * time.Second,
//	//}
//	//
//	//go func() {
//	//	log.Printf("🚀 服务启动在 http://0.0.0.0:%s", port)
//	//	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
//	//		log.Fatalf("❌ 服务器启动失败: %v", err)
//	//	}
//	//}()
//	//
//	//// 10. 优雅关闭
//	//quit := make(chan os.Signal, 1)
//	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//	//<-quit
//	//log.Println("🛑 接收到关闭信号，开始优雅关闭...")
//	//
//	//ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
//	//defer cancel()
//	//
//	//if err := server.Shutdown(ctx); err != nil {
//	//	log.Fatalf("❌ 服务器强制关闭: %v", err)
//	//}
//	//log.Println("✅ 服务器已优雅关闭")
//}
