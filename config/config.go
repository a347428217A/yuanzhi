//package config
//
//import (
//	"io/ioutil"
//	"log"
//	"os"
//
//	"github.com/joho/godotenv"
//)

package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type WechatPayConfig struct {
	AppID       string `yaml:"app_id"`
	MchID       string `yaml:"mch_id"`
	APIKey      string `yaml:"api_key"`
	CertPath    string `yaml:"cert_path"`
	KeyPath     string `yaml:"key_path"`
	NotifyURL   string `yaml:"notify_url"`
	UseSimulate bool   `yaml:"use_simulate"`
}

// 总配文件
type config struct {
	Server        server          `yaml:"server"`
	Db            db              `yaml:"db"`
	Redis         redis           `yaml:"redis"`
	ImageSettings imageSettings   `yaml:"imageSettings"`
	Log           log             `yaml:"log"`
	WechatPay     WechatPayConfig `yaml:"wechat_pay"`
}

// 项目端口配置
type server struct {
	Address string `yaml:"address"`
	Model   string `yaml:"model"`
}

// 数据库配置
type db struct {
	Dialects string `yaml:"dialects"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Db       string `yaml:"db"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Charset  string `yaml:"charset"`
	MaxIdle  int    `yaml:"maxIdle"`
	MaxOpen  int    `yaml:"maxOpen"`
}

// redis配置
type redis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

// imageSettings图片上传配置
type imageSettings struct {
	UploadDir string `yaml:"uploadDir"`
	ImageHost string `yaml:"imageHost"`
}

// log日志配置
type log struct {
	Path  string `yaml:"path"`
	Name  string `yaml:"name"`
	Model string `yaml:"model"`
}

type payment struct {
	UseSimulate bool `yaml:"use_simulate"` // 新增模拟支付开关
}

var Config *config

// 配置初始化
func init() {
	yamlFile, err := ioutil.ReadFile("./config.yaml")
	// 有错就down机
	if err != nil {
		panic(err)
	}
	// 绑定值
	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		panic(err)
	}
}

//type Config struct {
//	DBHost         string
//	DBPort         string
//	DBUser         string
//	DBPassword     string
//	DBName         string
//	Port           string
//	JWTSecret      string
//	WXAppID        string
//	WXAppSecret    string
//	SMSServiceURL  string
//	SMSServiceKey  string
//	AllowedOrigins string
//	Env            string
//}

//var AppConfig *Config
//
//func LoadConfig() {
//	if err := godotenv.Load(); err != nil {
//		log.Println("No .env file found, using system environment variables")
//	}
//
//	AppConfig = &Config{
//		DBHost:         getEnv("DB_HOST", "localhost"),
//		DBPort:         getEnv("DB_PORT", "3306"),
//		DBUser:         getEnv("DB_USER", "root"),
//		DBPassword:     getEnv("DB_PASSWORD", "a893782064A."),
//		DBName:         getEnv("DB_NAME", "appointment_db"),
//		Port:           getEnv("PORT", "8080"),
//		JWTSecret:      getEnv("JWT_SECRET", "default_secret"),
//		WXAppID:        getEnv("WX_APP_ID", "wx10ca8858028379ec"),
//		WXAppSecret:    getEnv("WX_APP_SECRET", "232fc9c655456253aed21efb6b230df3"),
//		SMSServiceURL:  getEnv("SMS_SERVICE_URL", ""),
//		SMSServiceKey:  getEnv("SMS_SERVICE_KEY", ""),
//		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
//		Env:            getEnv("ENV", "development"),
//	}
//}
//
//func getEnv(key, defaultValue string) string {
//	if value, exists := os.LookupEnv(key); exists {
//		return value
//	}
//	return defaultValue
//}
