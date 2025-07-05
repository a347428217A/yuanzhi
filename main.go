package main

import (
	_ "admin-api/docs"
	//"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

// @title Appointment System API
// @version 1.0
// @description 预约系统API文档
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// 1. 从环境变量获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	r := gin.Default()

	// 仅保留健康检查
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("🚀 简易服务启动在 :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}

	//port := os.Getenv("PORT")
	//if port == "" {
	//	port = "80" // 本地开发默认端口
	//	log.Printf("⚠️ PORT环境变量未设置，使用默认端口: %s", port)
	//} else {
	//	log.Printf("✅ 使用环境变量PORT: %s", port)
	//}
	//
	//// 2. 初始化数据库（确保database.InitDB()内部使用环境变量）
	//database.InitDB()
	//
	//// 3. 初始化Redis（确保redis.SetupRedisDb()内部使用环境变量）
	////err := redis.SetupRedisDb()
	////if err != nil {
	////	log.Fatalf("❌ Redis初始化失败: %v", err)
	////} else {
	////	log.Println("✅ Redis初始化成功")
	////}
	//
	//// 4. 创建路由
	//mainRouter := gin.Default()
	//
	//// 5. 添加Swagger路由
	//mainRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//log.Printf("🔍 Swagger UI 可用: http://0.0.0.0:%s/swagger/index.html", port)
	//
	//// 6. 设置路由
	//routes.SetupCustomerRoutes(mainRouter)
	//routes.SetupMerchantRoutes(mainRouter)
	//routes.SetupInternalRoutes(mainRouter)
	//
	//// 7. 添加健康检查端点（云托管需要）
	//mainRouter.GET("/health", func(c *gin.Context) {
	//	c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	//})
	//
	//// 8. 添加根路径健康检查（云托管默认检查）
	//mainRouter.GET("/", func(c *gin.Context) {
	//	c.JSON(200, gin.H{"status": "ok"})
	//})
	//
	//// 9. 启动HTTP服务器（使用环境变量端口）
	//server := &http.Server{
	//	Addr:         ":" + port, // 关键修改：使用环境变量端口
	//	Handler:      mainRouter,
	//	ReadTimeout:  15 * time.Second,
	//	WriteTimeout: 30 * time.Second,
	//	IdleTimeout:  60 * time.Second,
	//}
	//
	//go func() {
	//	log.Printf("🚀 服务启动在 http://0.0.0.0:%s", port)
	//	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	//		log.Fatalf("❌ 服务器启动失败: %v", err)
	//	}
	//}()
	//
	//// 10. 优雅关闭
	//quit := make(chan os.Signal, 1)
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	//<-quit
	//log.Println("🛑 接收到关闭信号，开始优雅关闭...")
	//
	//ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	//defer cancel()
	//
	//if err := server.Shutdown(ctx); err != nil {
	//	log.Fatalf("❌ 服务器强制关闭: %v", err)
	//}
	//log.Println("✅ 服务器已优雅关闭")
}
