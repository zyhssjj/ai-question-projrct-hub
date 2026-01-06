package controller

import (
    "net/http"
    "os"
    "path/filepath"
    "github.com/gin-gonic/gin"
)

// GetStudyNote 读取学习心得.md文件
func GetStudyNote(c *gin.Context) {
    // 获取当前工作目录
    currentDir, err := os.Getwd()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "msg":  "获取当前目录失败：" + err.Error(),
            "data": "",
        })
        return
    }
    
    // 构建学习心得.md的完整路径（在当前目录的上级目录中）
    notePath := filepath.Join(currentDir, "..", "学习心得.md")
    
    // 检查文件是否存在
    if _, err := os.Stat(notePath); os.IsNotExist(err) {
        // 如果文件不存在，创建默认的学习心得文件
        defaultContent := `# 学习心得

## 个人信息
- 学校：华南农业大学
- 姓名：[你的姓名]
- 学号：[你的学号]

## 课程收获
1. 学习了Go语言的基础语法和Web开发框架Gin
2. 掌握了React前端框架和状态管理
3. 理解了前后端分离的开发模式
4. 学会了SQLite数据库的使用

## 项目心得
通过本次题库管理系统的开发，我深入理解了：
- 全栈开发的完整流程
- RESTful API的设计原则
- 数据库表设计的基本原则
- 组件化开发的思想

## 遇到的问题
1. 初次接触SQLite时对嵌入式数据库不太了解
2. React状态管理需要一定时间适应
3. 跨域问题的解决需要前后端配合

## 未来计划
1. 深入学习Go语言的并发编程
2. 学习更复杂的前端框架
3. 尝试开发更大型的项目`

        // 创建文件并写入默认内容
        if err := os.WriteFile(notePath, []byte(defaultContent), 0644); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "code": 500,
                "msg":  "创建学习心得文件失败：" + err.Error(),
                "data": "",
            })
            return
        }
    }
    
    // 读取文件内容
    content, err := os.ReadFile(notePath)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "msg":  "读取学习心得失败：" + err.Error(),
            "data": "",
        })
        return
    }
    
    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "msg":  "success",
        "data": string(content),
    })
}