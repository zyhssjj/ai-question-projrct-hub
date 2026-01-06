package model

import (
    "gorm.io/gorm"
    "time"
)

// Question 题目数据模型
type Question struct {
    ID         uint           `json:"id" gorm:"primaryKey;autoIncrement"`      // 主键ID，自增
    Type       string         `json:"type" gorm:"type:varchar(20);not null"`    // 题型：单选题/多选题/编程题
    Content    string         `json:"content" gorm:"type:text;not null"`        // 题目内容
    Options    string         `json:"options" gorm:"type:text"`                 // 选项（JSON格式存储）
    Answer     string         `json:"answer" gorm:"type:varchar(100)"`          // 正确答案
    Difficulty string         `json:"difficulty" gorm:"type:varchar(10)"`       // 难度：简单/中等/困难
    Language   string         `json:"language" gorm:"type:varchar(20)"`         // 编程语言
    CreatedAt  time.Time      `json:"createdAt"`                                // 创建时间
    UpdatedAt  time.Time      `json:"updatedAt"`                                // 更新时间
    DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`                           // 软删除时间（不返回给前端）
}

// TableName 自定义表名
func (Question) TableName() string {
    return "questions"
}