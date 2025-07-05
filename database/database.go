package database

import (
	"database/sql"
	//"admin-api/models"
	//"admin-api/models"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// 1. 优先从环境变量获取配置
	dbHost := os.Getenv("DB_HOST")

	dbPort := os.Getenv("DB_PORT")

	dbUser := os.Getenv("DB_USER")

	dbPass := os.Getenv("DB_PASSWORD")

	dbName := os.Getenv("DB_NAME")

	// 2. 构建DSN（添加关键参数）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?"+
		"charset=utf8mb4&parseTime=True&loc=Local&"+
		"timeout=30s&readTimeout=30s&writeTimeout=30s", // 添加超时设置
		dbUser, dbPass, dbHost, dbPort, dbName)

	log.Printf("📡 尝试连接数据库: %s@%s:%s", dbUser, dbHost, dbPort)

	// 3. 添加重试逻辑
	var err error
	var sqlDB *sql.DB
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			// 获取底层连接
			sqlDB, err = DB.DB()
			if err == nil {
				// 验证连接是否真正可用
				if err = sqlDB.Ping(); err == nil {
					break
				}
			}
		}

		if i < 4 {
			log.Printf("⚠️ 连接失败(尝试 %d/5): %v", i+1, err)
			time.Sleep(time.Duration(i+1) * 2 * time.Second) // 指数退避
		}
	}

	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 4. 优化连接池设置（适配云数据库）
	// 腾讯云CynosDB建议设置：
	sqlDB.SetMaxIdleConns(5)                  // 降低空闲连接
	sqlDB.SetMaxOpenConns(20)                 // 降低最大连接
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // 缩短连接生命周期

	log.Printf("✅ 数据库连接成功 (空闲连接:%d, 最大连接:%d)",
		sqlDB.Stats().Idle, sqlDB.Stats().MaxOpenConnections)
}

//func InitDB() {
//	//cfg := config.AppConfig
//
//	dbHost := "sh-cynosdbmysql-grp-71t9co2k.sql.tencentcdb.com"
//	dbPort := "27308"
//	dbUser := "zoufy"
//	dbPass := "a893782064A."
//	dbName := "appintment_db"
//
//	// 构建DSN连接字符串
//	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
//		dbUser, dbPass, dbHost, dbPort, dbName)
//
//	//var dbConfig = config.Config.Db
//	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
//	//	dbConfig.Username,
//	//	dbConfig.Password,
//	//	dbConfig.Host,
//	//	dbConfig.Port,
//	//	dbConfig.Db,
//	//	dbConfig.Charset)
//
//	gormConfig := &gorm.Config{}
//
//	//if cfg.Env == "development" {
//	//	gormConfig.Logger = logger.Default.LogMode(logger.Info)
//	//}
//
//	log.Printf("📡 尝试连接数据库: %s@%s:%s", dbUser, dbHost, dbPort)
//
//	var err error
//	DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
//	if err != nil {
//		log.Fatalf("Failed to connect to database: %v", err)
//	}
//
//	sqlDB, err := DB.DB()
//	//if err != nil {
//	//	log.Fatalf("Failed to get DB instance: %v", err)
//	//}
//
//	if err != nil {
//		log.Fatalf("❌ 数据库连接失败: %v", err)
//	}
//	log.Println("✅ 数据库连接成功")
//
//	// Set connection pool parameters
//	sqlDB.SetMaxIdleConns(10)
//	sqlDB.SetMaxOpenConns(100)
//	sqlDB.SetConnMaxLifetime(time.Hour)
//
//	log.Println("✅ Database connected successfully")
//
//	// Auto migrate models
//	//autoMigrate()
//}

//func autoMigrate() {
//	err := DB.AutoMigrate(
//		&models.Merchant{},
//		&models.Staff{},
//		&models.ServiceCategory{},
//		&models.Service{},
//		&models.User{},
//		&models.TimeSlot{},
//		&models.Appointment{},
//		&models.CouponTemplate{},
//		&models.UserCoupon{},
//		&models.MerchantAdmin{},
//	)
//
//	if err != nil {
//		log.Fatalf("Failed to migrate database: %v", err)
//	}
//
//	log.Println("✅ Database migrated successfully")
//}
