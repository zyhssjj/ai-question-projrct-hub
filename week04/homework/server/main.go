package main

import (
    "fmt"
    "log"
    "net/http"
    "question-bank/config"
    "question-bank/router"
)

func main() {
    fmt.Println("🚀 启动题库管理系统后端服务...")
    
    // 1. 加载环境变量
    config.LoadEnv()
    
    // 2. 初始化数据库
    config.InitDB()
    
    // 3. 设置路由
    r := router.SetupRouter()
    
    // 4. 启动HTTP服务器
    serverAddr := ":" + config.ServerPort
    log.Printf("✅ 后端服务启动成功！")
    log.Printf("📡 服务地址：http://localhost%s", serverAddr)
    log.Printf("📚 API文档：http://localhost%s/api", serverAddr)
    log.Println("🛑 按 Ctrl+C 停止服务")
    
    // 启动服务器
    if err := http.ListenAndServe(serverAddr, r); err != nil {
        log.Fatalf("❌ 服务器启动失败：%v", err)
    }
}