package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

// 全局配置变量
var (
    AIApiKey   string
    AIBaseURL  string
    ServerPort string
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
    ServerPort = os.Getenv("PORT")
    
    // 设置默认值
    if ServerPort == "" {
        ServerPort = "8080"
    }
    
    // 验证必需的环境变量
    if AIApiKey == "" {
        log.Fatal("❌ 请设置AI_API_KEY环境变量")
    }
    
    log.Println("✅ 环境变量加载成功")
}