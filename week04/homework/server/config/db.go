package config

import (
	"github.com/glebarez/sqlite" // 纯Go的SQLite驱动，不需要CGO
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"question-bank/model" // 导入模型
	"time"
)

// DB 全局数据库连接变量
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	var err error
	if DatabaseURL != "" {
		DB, err = gorm.Open(postgres.Open(DatabaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	} else {
		DB, err = gorm.Open(sqlite.Open("question_bank.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	}
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("❌ 获取数据库连接池失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	if DatabaseURL != "" {
		if err = DB.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
			log.Fatalf("❌ pgvector 扩展初始化失败: %v", err)
		}
	}
	// UserProfile is introduced after deployments commonly disabled full auto migration.
	// Migrating this additive table is idempotent and does not alter existing learning data.
	if err = DB.AutoMigrate(&model.UserProfile{}); err != nil {
		log.Fatalf("❌ 个人资料表初始化失败: %v", err)
	}

	if AutoMigrate {
		err = DB.AutoMigrate(&model.AnonymousSession{}, &model.UserProfile{}, &model.Material{}, &model.DocumentChunk{}, &model.GenerationJob{}, &model.AgentConversation{}, &model.AgentMessage{}, &model.Exam{}, &model.ExamQuestion{}, &model.ExamAttempt{}, &model.AttemptAnswer{}, &model.WrongQuestion{})
		if err != nil {
			log.Fatalf("❌ 数据库表创建失败: %v", err)
		}
	}
	if DatabaseURL != "" && AutoMigrate {
		if err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding ON document_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 10)").Error; err != nil {
			log.Fatalf("❌ pgvector 索引创建失败: %v", err)
		}
	}

	if os.Getenv("DATABASE_URL") != "" {
		log.Println("✅ PostgreSQL 数据库初始化成功")
	} else {
		log.Println("✅ SQLite 数据库初始化成功：question_bank.db")
	}
}
