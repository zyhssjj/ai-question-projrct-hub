package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"question-bank/config"
	"question-bank/controller"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建Gin引擎（默认中间件）
	r := gin.Default()

	// 1. 全局中间件
	// CORS中间件（允许前端跨域请求）
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.FrontendOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
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
		api.POST("/session/bootstrap", controller.BootstrapSession)
		learning := api.Group("", controller.SessionMiddleware())
		learning.POST("/materials", controller.UploadMaterial)
		learning.GET("/materials", controller.ListMaterials)
		learning.GET("/materials/:id", controller.GetMaterial)
		learning.GET("/jobs/:id", controller.GetJob)
		learning.POST("/agent/conversations", controller.CreateConversation)
		learning.GET("/agent/conversations/:id", controller.GetConversation)
		learning.POST("/agent/conversations/:id/messages", controller.AddMessage)
		learning.POST("/agent/conversations/:id/confirm", controller.ConfirmConversation)
		learning.GET("/exams", controller.ListExams)
		learning.GET("/exams/:id", controller.GetExam)
		learning.POST("/exams/:id/start", controller.StartExam)
		learning.GET("/attempts/:id/questions", controller.GetAttemptQuestions)
		learning.PUT("/attempts/:id/answers/:questionId", controller.SaveAnswer)
		learning.PUT("/attempts/:id/marks/:questionId", controller.MarkQuestion)
		learning.POST("/attempts/:id/submit", controller.SubmitAttempt)
		learning.GET("/attempts/:id/result", controller.AttemptResult)
		learning.POST("/attempts/:id/wrong-questions/retry", controller.RetryWrong)
		learning.GET("/wrong-questions", controller.ListWrongQuestions)
		learning.GET("/wrong-books", controller.ListWrongBooks)
		learning.GET("/profile", controller.GetProfile)
		learning.PUT("/profile", controller.UpdateProfile)
		learning.PUT("/profile/password", controller.ChangePassword)
		learning.GET("/settings/ai", controller.GetAISettings)
		learning.PUT("/settings/ai", controller.UpdateAISettings)
		learning.POST("/settings/ai/test", controller.TestAISettings)
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
