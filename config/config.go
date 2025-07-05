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
	"os"
	"path/filepath"
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

// 添加全局初始化标志
var initialized bool

// 配置初始化（改为可调用的函数）
func Init(configPath string) {
	// 如果已经初始化则跳过
	if initialized {
		return
	}

	// 1. 确定最终配置文件路径
	finalPath := resolveConfigPath(configPath)
	//log.Printf("🔧 加载配置文件: %s", finalPath)

	// 2. 读取配置文件
	yamlFile, err := os.ReadFile(finalPath)
	if err != nil {
		//log.Fatalf("❌ 读取配置文件失败: %s | %v", finalPath, err)
	}

	// 3. 解析配置
	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		//log.Fatalf("❌ 解析配置文件失败: %v", err)
	}

	//log.Printf("✅ 配置文件加载成功")
	initialized = true
}

// 解析配置文件路径
func resolveConfigPath(userPath string) string {
	// 1. 优先使用用户指定的路径
	if userPath != "" {
		return userPath
	}

	// 2. 尝试环境变量指定路径
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		return envPath
	}

	// 3. 尝试当前工作目录
	if cwd, err := os.Getwd(); err == nil {
		defaultPath := filepath.Join(cwd, "config.yaml")
		if _, err := os.Stat(defaultPath); err == nil {
			return defaultPath
		}
	}

	// 4. 尝试可执行文件所在目录
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		defaultPath := filepath.Join(exeDir, "config.yaml")
		if _, err := os.Stat(defaultPath); err == nil {
			return defaultPath
		}
	}

	// 5. 尝试常用位置
	commonPaths := []string{
		"/etc/app/config.yaml",
		"/app/config.yaml",
		"/config/config.yaml",
	}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	//log.Fatal("❌ 无法找到配置文件，请通过环境变量 CONFIG_PATH 指定")
	return "" // 不会执行到这里
}

//var Config *config
//
//// 配置初始化
//func init() {
//	yamlFile, err := ioutil.ReadFile("./config.yaml")
//	// 有错就down机
//	if err != nil {
//		panic(err)
//	}
//	// 绑定值
//	err = yaml.Unmarshal(yamlFile, &Config)
//	if err != nil {
//		panic(err)
//	}
//}

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
