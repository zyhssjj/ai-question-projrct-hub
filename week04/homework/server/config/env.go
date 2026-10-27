package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

// 全局配置变量
var (
	AIApiKey               string
	AIBaseURL              string
	AIModel                string
	DatabaseURL            string
	SupabaseURL            string
	SupabaseServiceRoleKey string
	SupabaseBucket         string
	TesseractPath          string
	PopplerBinPath         string
	FrontendOrigin         string
	ServerPort             string
	AutoMigrate            bool
	CookieSecure           bool
)

// LoadEnv 加载环境变量
func LoadEnv() {
	// 尝试从.env文件加载
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env文件未找到，使用系统环境变量")
	}

	// 读取环境变量
	AIApiKey = os.Getenv("AI_API_KEY")
	AIBaseURL = os.Getenv("AI_BASE_URL")
	AIModel = os.Getenv("AI_MODEL")
	DatabaseURL = os.Getenv("DATABASE_URL")
	SupabaseURL = os.Getenv("SUPABASE_URL")
	SupabaseServiceRoleKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	SupabaseBucket = os.Getenv("SUPABASE_BUCKET")
	TesseractPath = os.Getenv("TESSERACT_PATH")
	PopplerBinPath = os.Getenv("POPPLER_BIN_PATH")
	FrontendOrigin = os.Getenv("FRONTEND_ORIGIN")
	ServerPort = os.Getenv("PORT")
	AutoMigrate = strings.ToLower(os.Getenv("AUTO_MIGRATE")) != "false"
	CookieSecure = strings.ToLower(os.Getenv("COOKIE_SECURE")) == "true"

	// 设置默认值
	if ServerPort == "" {
		ServerPort = "8080"
	}
	if AIModel == "" {
		AIModel = "deepseek-v4-flash"
	}
	if FrontendOrigin == "" {
		FrontendOrigin = "http://localhost:5173"
	}

	// 验证必需的环境变量
	if AIApiKey == "" {
		log.Fatal("❌ 请设置AI_API_KEY环境变量")
	}

	log.Println("✅ 环境变量加载成功")
}
