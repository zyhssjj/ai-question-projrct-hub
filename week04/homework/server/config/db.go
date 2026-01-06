package config

import (
    "log"
    "github.com/glebarez/sqlite"  // 纯Go的SQLite驱动，不需要CGO
    "gorm.io/gorm"
    "question-bank/model"  // 导入模型
)

// DB 全局数据库连接变量
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
    var err error
    // 连接SQLite数据库（如果文件不存在会自动创建）
    DB, err = gorm.Open(sqlite.Open("question_bank.db"), &gorm.Config{})
    if err != nil {
        log.Fatalf("❌ 数据库连接失败: %v", err)
    }
    
    // 自动迁移数据库（创建表）
    err = DB.AutoMigrate(&model.Question{})
    if err != nil {
        log.Fatalf("❌ 数据库表创建失败: %v", err)
    }
    
    log.Println("✅ 数据库初始化成功！数据库文件：question_bank.db")
}