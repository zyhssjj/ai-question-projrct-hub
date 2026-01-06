package router

import (
    "net/http"
    "question-bank/controller"
    "github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
    // 创建Gin引擎（默认中间件）
    r := gin.Default()
    
    // 1. 全局中间件
    // CORS中间件（允许前端跨域请求）
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        
        c.Next()
    })
    
    // 2. 静态文件服务（用于生产环境托管前端文件）
    r.Static("/static", "./static")
    r.StaticFile("/", "./static/index.html")
    r.StaticFile("/index.html", "./static/index.html")
    
    // 3. API路由组（所有API以/api开头）
    api := r.Group("/api")
    {
        // 3.1 学习心得相关接口
        api.GET("/note/get", controller.GetStudyNote)
        
        // 3.2 题目管理相关接口
        api.GET("/question/list", controller.GetQuestionList)      // 查询题目列表
        api.POST("/question/add", controller.AddQuestion)          // 添加单个题目
        api.POST("/question/add-batch", controller.BatchAddQuestion) // 批量添加题目
        api.POST("/question/update", controller.UpdateQuestion)    // 更新题目
        api.POST("/question/delete", controller.DeleteQuestion)    // 删除题目
        api.POST("/question/ai-generate", controller.AIGenerateQuestion) // AI生成题目
    }
    
    // 4. 404处理
    r.NoRoute(func(c *gin.Context) {
        c.JSON(http.StatusNotFound, gin.H{
            "code": 404,
            "msg":  "请求的资源不存在",
        })
    })
    
    return r
}