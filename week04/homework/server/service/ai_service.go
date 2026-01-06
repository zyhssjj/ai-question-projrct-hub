package service

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "question-bank/config"
    "question-bank/model"
    "strings"
    "time"
    "log"
)

// 阿里云百炼请求结构体
type BailianRequest struct {
    Model    string                 `json:"model"`
    Input    BailianInput           `json:"input"`
    Parameters BailianParameters    `json:"parameters"`
}

type BailianInput struct {
    Messages []BailianMessage `json:"messages"`
}

type BailianMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type BailianParameters struct {
    Temperature float64 `json:"temperature"`
    TopP        float64 `json:"top_p"`
    MaxTokens   int     `json:"max_tokens"`
}

// 阿里云百炼响应结构体
type BailianResponse struct {
    Output struct {
        Text        string `json:"text"`
        FinishReason string `json:"finish_reason"`
    } `json:"output"`
    Usage struct {
        InputTokens  int `json:"input_tokens"`
        OutputTokens int `json:"output_tokens"`
        TotalTokens  int `json:"total_tokens"`
    } `json:"usage"`
    RequestID string `json:"request_id"`
}

// GenerateQuestions 调用阿里云百炼API生成题目
func GenerateQuestions(params model.Question, count int) ([]model.Question, error) {
    log.Printf("🚀 调用阿里云百炼API生成题目，参数: %+v, 数量: %d", params, count)
    
    // 构建适合阿里云百炼的提示词
    prompt := buildBailianPrompt(params, count)
    
    // 准备请求
    bailianReq := BailianRequest{
        Model: "qwen-max", // 使用通义千问模型
        Input: BailianInput{
            Messages: []BailianMessage{
                {
                    Role:    "user",
                    Content: prompt,
                },
            },
        },
        Parameters: BailianParameters{
            Temperature: 0.7,
            TopP:        0.8,
            MaxTokens:   2000,
        },
    }
    
    jsonData, err := json.Marshal(bailianReq)
    if err != nil {
        log.Printf("❌ JSON编码失败: %v", err)
        return nil, fmt.Errorf("JSON编码失败: %v", err)
    }
    
    log.Printf("📤 发送请求到阿里云百炼, 数据长度: %d", len(jsonData))
    
    // 阿里云百炼API地址 (根据官方文档)
    apiURL := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
    
    // 创建HTTP请求
    req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("❌ 创建请求失败: %v", err)
        return nil, fmt.Errorf("创建请求失败: %v", err)
    }
    
    // 设置请求头 (阿里云百炼要求)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+config.AIApiKey)
    req.Header.Set("X-DashScope-SSE", "disable") // 禁用流式输出
    
    // 发送请求
    client := &http.Client{Timeout: 60 * time.Second} // 延长超时时间
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("❌ 网络请求失败: %v", err)
        return nil, fmt.Errorf("AI请求失败: %v", err)
    }
    defer resp.Body.Close()
    
    // 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("❌ 读取响应失败: %v", err)
        return nil, fmt.Errorf("读取响应失败: %v", err)
    }
    
    log.Printf("📥 收到响应, 状态码: %d, 长度: %d", resp.StatusCode, len(body))
    
    // 检查状态码
    if resp.StatusCode != 200 {
        log.Printf("❌ 阿里云百炼返回错误: %s", string(body))
        return nil, fmt.Errorf("阿里云百炼API错误: 状态码 %d, 响应: %s", resp.StatusCode, string(body))
    }
    
    // 解析响应
    var bailianResp BailianResponse
    if err := json.Unmarshal(body, &bailianResp); err != nil {
        log.Printf("❌ 响应解析失败: %v, 原始响应: %s", err, string(body))
        return nil, fmt.Errorf("AI响应解析失败: %v", err)
    }
    
    log.Printf("✅ 阿里云百炼API调用成功, 生成文本长度: %d", len(bailianResp.Output.Text))
    
    // 解析生成的题目
    return parseBailianResponse(bailianResp.Output.Text, params, count)
}

// buildBailianPrompt 构建阿里云百炼的提示词
func buildBailianPrompt(params model.Question, count int) string {
    var prompt strings.Builder
    
    prompt.WriteString(fmt.Sprintf("你是一个专业的编程教育专家，请生成%d道%s题目。\n\n", count, params.Type))
    prompt.WriteString("具体要求：\n")
    prompt.WriteString(fmt.Sprintf("1. 题型：%s\n", params.Type))
    prompt.WriteString(fmt.Sprintf("2. 难度：%s\n", params.Difficulty))
    prompt.WriteString(fmt.Sprintf("3. 编程语言：%s\n", params.Language))
    
    if params.Type == "单选题" || params.Type == "多选题" {
        prompt.WriteString("\n请按以下格式返回每道题目：\n")
        prompt.WriteString("题目：[题目内容]\n")
        prompt.WriteString("选项：\nA. [选项A]\nB. [选项B]\nC. [选项C]\nD. [选项D]\n")
        prompt.WriteString("答案：[正确答案，多选题用逗号分隔，如A,B]\n")
        prompt.WriteString("---\n")
        
        prompt.WriteString("\n示例：\n")
        prompt.WriteString("题目：在Go语言中，下面哪个关键字用于声明一个变量？\n")
        prompt.WriteString("选项：\nA. var\nB. let\nC. const\nD. def\n")
        prompt.WriteString("答案：A\n")
        prompt.WriteString("---\n")
    } else if params.Type == "编程题" {
        prompt.WriteString("\n请按以下格式返回每道题目：\n")
        prompt.WriteString("题目：[题目描述]\n")
        prompt.WriteString("要求：[具体编程要求]\n")
        prompt.WriteString("示例输入：[示例输入]\n")
        prompt.WriteString("示例输出：[示例输出]\n")
        prompt.WriteString("---\n")
        
        prompt.WriteString("\n示例：\n")
        prompt.WriteString("题目：实现一个函数计算两个数的和\n")
        prompt.WriteString("要求：使用Go语言实现add函数，接收两个整数参数，返回它们的和\n")
        prompt.WriteString("示例输入：3, 5\n")
        prompt.WriteString("示例输出：8\n")
        prompt.WriteString("---\n")
    }
    
    prompt.WriteString("\n请确保题目质量高、难度适中、描述清晰。每道题目用'---'分隔。")
    
    return prompt.String()
}

// parseBailianResponse 解析阿里云百炼的响应
func parseBailianResponse(content string, params model.Question, count int) ([]model.Question, error) {
	log.Printf("🔍 开始解析AI响应，内容长度: %d", len(content))
	
	// 按'---'分隔题目
	questionBlocks := strings.Split(content, "---")
	var questions []model.Question
	
	for i, block := range questionBlocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		
		question := model.Question{
			Type:       params.Type,
			Difficulty: params.Difficulty,
			Language:   params.Language,
		}
		
		lines := strings.Split(block, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			
			if strings.HasPrefix(line, "题目：") {
				question.Content = strings.TrimPrefix(line, "题目：")
			} else if strings.HasPrefix(line, "要求：") {
				if question.Content != "" {
					question.Content += "\n" + line
				} else {
					question.Content = line
				}
			} else if strings.HasPrefix(line, "选项：") {
				// 处理后续的选项行
			} else if strings.HasPrefix(line, "A.") || strings.HasPrefix(line, "B.") ||
					  strings.HasPrefix(line, "C.") || strings.HasPrefix(line, "D.") {
				// 收集选项
				var options []string
				for _, optLine := range lines {
					optLine = strings.TrimSpace(optLine)
					if strings.HasPrefix(optLine, "A.") || strings.HasPrefix(optLine, "B.") ||
					   strings.HasPrefix(optLine, "C.") || strings.HasPrefix(optLine, "D.") {
						options = append(options, optLine)
					}
				}
				if len(options) > 0 {
					optsJSON, _ := json.Marshal(options)
					question.Options = string(optsJSON)
				}
			} else if strings.HasPrefix(line, "答案：") {
				question.Answer = strings.TrimPrefix(line, "答案：")
			} else if strings.HasPrefix(line, "示例输入：") || strings.HasPrefix(line, "示例输出：") {
				// 对于编程题，将示例信息添加到题目内容中
				if question.Content != "" {
					question.Content += "\n" + line
				}
			}
		}
		
		// 如果没有解析到内容，添加默认内容
		if question.Content == "" {
			question.Content = fmt.Sprintf("第%d道%s题目，请完善内容", i+1, params.Type)
		}
		
		questions = append(questions, question)
		
		// 如果已经达到要求的数量，就停止
		if len(questions) >= count {
			break
		}
	}
	
	// 如果解析失败，使用模拟数据
	if len(questions) == 0 {
		log.Println("⚠️ AI响应解析失败，使用模拟数据")
		return generateBailianMockQuestions(params, count), nil
	}
	
	log.Printf("✅ 成功解析 %d 道题目", len(questions))
	return questions, nil
}

// generateBailianMockQuestions 阿里云百炼失败时的模拟数据
func generateBailianMockQuestions(params model.Question, count int) []model.Question {
    log.Printf("🎨 生成模拟题目用于演示")
    
    questions := make([]model.Question, count)
    
    for i := 0; i < count; i++ {
        questions[i] = model.Question{
            Type:       params.Type,
            Difficulty: params.Difficulty,
            Language:   params.Language,
        }
        
        switch params.Type {
        case "单选题":
            questions[i].Content = fmt.Sprintf("关于%s的%s单选题示例 %d", params.Language, params.Difficulty, i+1)
            options := []string{
                fmt.Sprintf("A. %s的正确用法", params.Language),
                fmt.Sprintf("B. %s的错误用法1", params.Language),
                fmt.Sprintf("C. %s的错误用法2", params.Language),
                fmt.Sprintf("D. %s的错误用法3", params.Language),
            }
            optsJSON, _ := json.Marshal(options)
            questions[i].Options = string(optsJSON)
            questions[i].Answer = "A"
            
        case "多选题":
            questions[i].Content = fmt.Sprintf("关于%s的%s多选题示例 %d", params.Language, params.Difficulty, i+1)
            options := []string{
                fmt.Sprintf("A. %s的正确特性1", params.Language),
                fmt.Sprintf("B. %s的正确特性2", params.Language),
                fmt.Sprintf("C. %s的错误特性", params.Language),
                fmt.Sprintf("D. %s的错误用法", params.Language),
            }
            optsJSON, _ := json.Marshal(options)
            questions[i].Options = string(optsJSON)
            questions[i].Answer = "A,B"
            
        case "编程题":
            questions[i].Content = fmt.Sprintf("%s编程题示例 %d\n\n请使用%s语言实现一个函数，完成指定功能。\n\n要求：\n1. 代码规范\n2. 考虑边界情况\n3. 添加适当注释", 
                params.Difficulty, i+1, params.Language)
            questions[i].Answer = "参考实现代码"
        }
    }
    
    return questions
}