package controller

import (
    "net/http"
    "strconv"
    "question-bank/config"
    "question-bank/model"
    "question-bank/service"
    "github.com/gin-gonic/gin"
    "log"
)

// GetQuestionList 查询题目列表（分页+筛选+搜索）
func GetQuestionList(c *gin.Context) {
    // 1. 获取查询参数
    pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
    questionType := c.Query("type")
    keyword := c.Query("keyword")
    
    // 2. 参数验证
    if pageNum < 1 {
        pageNum = 1
    }
    if pageSize < 1 {
        pageSize = 10
    }
    
    // 3. 构建查询条件
    db := config.DB.Model(&model.Question{})
    
    // 3.1 按题型筛选
    if questionType != "" && questionType != "全部" {
        db = db.Where("type = ?", questionType)
    }
    
    // 3.2 按关键词搜索（题目内容模糊匹配）
    if keyword != "" {
        db = db.Where("content LIKE ?", "%"+keyword+"%")
    }
    
    // 4. 查询总数
    var total int64
    db.Count(&total)
    
    // 5. 分页查询
    var questions []model.Question
    offset := (pageNum - 1) * pageSize
    db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&questions)
    
    // 6. 处理选项字段（JSON字符串转数组）
    for i := range questions {
        if questions[i].Options != "" {
            // 这里前端会直接使用JSON字符串，所以不需要转换
        }
    }
    
    // 7. 返回响应
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "msg":  "查询成功",
        "data": gin.H{
            "list":  questions,
            "total": total,
            "pageNum": pageNum,
            "pageSize": pageSize,
        },
    })
}

// AddQuestion 添加单个题目（手工出题）
func AddQuestion(c *gin.Context) {
    var question model.Question
    
    // 1. 绑定请求参数
    if err := c.ShouldBindJSON(&question); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "请求参数错误：" + err.Error(),
        })
        return
    }
    
    // 2. 验证必填字段
    if question.Content == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "题目内容不能为空",
        })
        return
    }
    
    // 3. 保存到数据库
    if err := config.DB.Create(&question).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "msg":  "保存题目失败：" + err.Error(),
        })
        return
    }
    
    // 4. 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "msg":  "题目添加成功",
        "data": question,
    })
}

// BatchAddQuestion 批量添加题目（AI出题后确认添加）
func BatchAddQuestion(c *gin.Context) {
    var questions []model.Question
    
    // 1. 绑定请求参数
    if err := c.ShouldBindJSON(&questions); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "请求参数错误：" + err.Error(),
        })
        return
    }
    
    // 2. 验证数据
    if len(questions) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "题目列表不能为空",
        })
        return
    }
    
    // 3. 批量保存到数据库
    if err := config.DB.Create(&questions).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "msg":  "批量保存题目失败：" + err.Error(),
        })
        return
    }
    
    // 4. 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "msg":  "批量添加成功",
        "data": gin.H{
            "count": len(questions),
        },
    })
}

// UpdateQuestion 更新题目（编辑）
func UpdateQuestion(c *gin.Context) {
    var question model.Question
    
    // 1. 绑定请求参数
    if err := c.ShouldBindJSON(&question); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "请求参数错误：" + err.Error(),
        })
        return
    }
    
    // 2. 验证ID是否存在
    if question.ID == 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "题目ID不能为空",
        })
        return
    }
    
    // 3. 检查题目是否存在
    var existingQuestion model.Question
    if err := config.DB.First(&existingQuestion, question.ID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "code": 404,
            "msg":  "题目不存在",
        })
        return
    }
    
    // 4. 更新数据库
    if err := config.DB.Model(&existingQuestion).Updates(question).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "msg":  "更新题目失败：" + err.Error(),
        })
        return
    }
    
    // 5. 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "msg":  "题目更新成功",
        "data": existingQuestion,
    })
}

// DeleteQuestion 删除题目（支持单个和批量）
func DeleteQuestion(c *gin.Context) {
    var request struct {
        IDs []uint `json:"ids" binding:"required"`
    }
    
    // 1. 绑定请求参数
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "请求参数错误：" + err.Error(),
        })
        return
    }
    
    // 2. 验证ID列表
    if len(request.IDs) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "请选择要删除的题目",
        })
        return
    }
    
    // 3. 批量删除（软删除）
    result := config.DB.Delete(&model.Question{}, request.IDs)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "msg":  "删除题目失败：" + result.Error.Error(),
        })
        return
    }
    
    // 4. 检查是否成功删除
    if result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{
            "code": 404,
            "msg":  "未找到对应的题目",
        })
        return
    }
    
    // 5. 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "msg":  "删除成功",
        "data": gin.H{
            "deletedCount": result.RowsAffected,
        },
    })
}

// // AIGenerateQuestion AI生成题目
// func AIGenerateQuestion(c *gin.Context) {
//     var params struct {
//         Type       string `json:"type" binding:"required"`
//         Count      int    `json:"count" binding:"required,min=1,max=10"`
//         Difficulty string `json:"difficulty" binding:"required"`
//         Language   string `json:"language" binding:"required"`
//     }
    
//     // 1. 绑定请求参数
//     if err := c.ShouldBindJSON(&params); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "code": 400,
//             "msg":  "请求参数错误：" + err.Error(),
//         })
//         return
//     }
    
//     // 2. 验证参数范围
//     if params.Count < 1 || params.Count > 10 {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "code": 400,
//             "msg":  "题目数量必须在1-10之间",
//         })
//         return
//     }
    
//     // 3. 调用AI服务生成题目
//     baseQuestion := model.Question{
//         Type:       params.Type,
//         Difficulty: params.Difficulty,
//         Language:   params.Language,
//     }
    
//     questions, err := service.GenerateQuestions(baseQuestion, params.Count)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{
//             "code": 500,
//             "msg":  "AI生成题目失败：" + err.Error(),
//         })
//         return
//     }
    
//     // 4. 返回生成的题目
//     c.JSON(http.StatusOK, gin.H{
//         "code": 200,
//         "msg":  "AI生成成功",
//         "data": questions,
//     })
// }
// 6. AI 生成题目 - 阿里云百炼版本
func AIGenerateQuestion(c *gin.Context) {
    var params struct {
        Type       string `json:"type" binding:"required"`
        Count      int    `json:"count" binding:"required,min=1,max=10"`
        Difficulty string `json:"difficulty" binding:"required"`
        Language   string `json:"language" binding:"required"`
    }
    
    // 1. 绑定请求参数
    if err := c.ShouldBindJSON(&params); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "参数错误：" + err.Error(),
        })
        return
    }
    
    // 2. 验证参数范围
    if params.Count < 1 || params.Count > 10 {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "msg":  "题目数量必须在1-10之间",
        })
        return
    }
    
    log.Printf("🎯 收到AI生成请求: type=%s, count=%d, difficulty=%s, language=%s", 
        params.Type, params.Count, params.Difficulty, params.Language)
    
    // 3. 调用阿里云百炼服务生成题目
    baseQuestion := model.Question{
        Type:       params.Type,
        Difficulty: params.Difficulty,
        Language:   params.Language,
    }
    
    questions, err := service.GenerateQuestions(baseQuestion, params.Count)
    if err != nil {
        log.Printf("❌ AI生成题目失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "msg":  "AI生成题目失败：" + err.Error(),
            "data": []interface{}{}, // 返回空数组而不是nil
        })
        return
    }
    
    log.Printf("✅ AI生成成功，返回 %d 道题目", len(questions))
    
    // 4. 返回生成的题目
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "msg":  "生成成功",
        "data": questions,
    })
}