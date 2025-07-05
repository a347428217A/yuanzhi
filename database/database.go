package database

import (
	//"admin-api/models"
	//"admin-api/models"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	//cfg := config.AppConfig

	dbHost := "sh-cynosdbmysql-grp-71t9co2k.sql.tencentcdb.com"
	dbPort := "27308"
	dbUser := "zoufy"
	dbPass := "a893782064A."
	dbName := "appintment_db"

	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	//var dbConfig = config.Config.Db
	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
	//	dbConfig.Username,
	//	dbConfig.Password,
	//	dbConfig.Host,
	//	dbConfig.Port,
	//	dbConfig.Db,
	//	dbConfig.Charset)

	gormConfig := &gorm.Config{}

	//if cfg.Env == "development" {
	//	gormConfig.Logger = logger.Default.LogMode(logger.Info)
	//}

	log.Printf("📡 尝试连接数据库: %s@%s:%s", dbUser, dbHost, dbPort)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	//if err != nil {
	//	log.Fatalf("Failed to get DB instance: %v", err)
	//}

	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	log.Println("✅ 数据库连接成功")

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connected successfully")

	// Auto migrate models
	//autoMigrate()
}

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
