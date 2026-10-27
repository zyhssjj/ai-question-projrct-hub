package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"question-bank/config"
	"strings"
	"time"
)

type GeneratedQuestion struct {
	Type        string   `json:"type"`
	Content     string   `json:"content"`
	Options     []Option `json:"options"`
	Answer      []string `json:"answer"`
	Explanation string   `json:"explanation"`
	Citation    string   `json:"citation"`
}
type Option struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}
type chatRequest struct {
	Model          string            `json:"model"`
	Messages       []chatMessage     `json:"messages"`
	Temperature    float64           `json:"temperature"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type ExamPlan struct {
	Title           string `json:"title"`
	QuestionCount   int    `json:"questionCount"`
	DurationMinutes int    `json:"durationMinutes"`
	Difficulty      string `json:"difficulty"`
	SingleCount     int    `json:"singleCount"`
	MultipleCount   int    `json:"multipleCount"`
	Focus           string `json:"focus"`
}

type PlanResult struct {
	Ready   bool     `json:"ready"`
	Reply   string   `json:"reply"`
	Missing []string `json:"missing"`
	Plan    ExamPlan `json:"plan"`
}

func TestAIConnection() (string, error) {
	return callJSON(`只返回 JSON：{"status":"ok"}`)
}

func BuildExamPlan(history []string, currentPlan string) (PlanResult, error) {
	prompt := fmt.Sprintf(`你是讲义出卷 Agent。根据对话抽取出卷计划；不能生成题目。缺失信息时用 reply 简短追问。
必填：试卷名称、总题量(1-50)、单选数量、多选数量、难度(简单/中等/困难)、时长分钟(1-180)、章节或知识点。
单选数+多选数必须等于总题量。只返回 JSON：{"ready":true,"reply":"计划已完整，请确认","missing":[],"plan":{"title":"...","questionCount":10,"durationMinutes":30,"difficulty":"中等","singleCount":6,"multipleCount":4,"focus":"..."}}。
当前计划：%s
对话：%s`, currentPlan, strings.Join(history, "\n"))
	content, err := callJSON(prompt)
	if err != nil {
		return PlanResult{}, err
	}
	var result PlanResult
	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return result, fmt.Errorf("Agent 返回无效 JSON：%w", err)
	}
	if result.Ready {
		p := result.Plan
		if p.Title == "" || p.Focus == "" || p.QuestionCount < 1 || p.QuestionCount > 50 || p.DurationMinutes < 1 || p.DurationMinutes > 180 || p.SingleCount < 0 || p.MultipleCount < 0 || p.SingleCount+p.MultipleCount != p.QuestionCount {
			return result, fmt.Errorf("Agent 返回的出卷计划不完整")
		}
	}
	return result, nil
}

func GenerateExamQuestions(context, focus, difficulty string, singleCount, multipleCount int) ([]GeneratedQuestion, error) {
	if config.AIApiKey == "" {
		return nil, fmt.Errorf("未配置 AI_API_KEY")
	}
	prompt := fmt.Sprintf(`你是严谨的中文教育测验命题专家。只能根据下方讲义生成题目，不得编造讲义外事实。生成 %d 道单选题和 %d 道多选题，难度：%s，重点：%s。
只返回 JSON 对象，格式为 {"questions":[{"type":"single|multiple","content":"题干","options":[{"id":"A","content":"..."}],"answer":["A"],"explanation":"解析","citation":"讲义原文短引"}]}。
单选必须 4 个选项且 answer 只有 1 项；多选必须 4 个选项且 answer 有 2-3 项；引用必须来自讲义。
讲义：

%s`, singleCount, multipleCount, difficulty, focus, context)
	questions, err := generateAndValidate(prompt, context, singleCount+multipleCount)
	if err == nil {
		return questions, nil
	}
	repairPrompt := prompt + "\n\n上一次输出校验失败：" + err.Error() + "。请修正所有问题并重新返回完整 JSON。"
	questions, repairErr := generateAndValidate(repairPrompt, context, singleCount+multipleCount)
	if repairErr != nil {
		return nil, fmt.Errorf("AI 题目质量校验失败，自动修复后仍不合格：%w", repairErr)
	}
	return questions, nil
}

func generateAndValidate(prompt, context string, expected int) ([]GeneratedQuestion, error) {
	content, err := callJSON(prompt)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Questions []GeneratedQuestion `json:"questions"`
	}
	if err = json.Unmarshal([]byte(content), &payload); err != nil {
		return nil, fmt.Errorf("AI 返回的题目不是有效 JSON：%w", err)
	}
	if len(payload.Questions) != expected {
		return nil, fmt.Errorf("AI 返回题量不正确：期望 %d，实际 %d", expected, len(payload.Questions))
	}
	seen := map[string]bool{}
	for _, q := range payload.Questions {
		if err := ValidateQuestion(q); err != nil {
			return nil, err
		}
		key := strings.TrimSpace(strings.ToLower(q.Content))
		if seen[key] {
			return nil, fmt.Errorf("存在重复题目")
		}
		seen[key] = true
		if !strings.Contains(context, strings.TrimSpace(q.Citation)) {
			return nil, fmt.Errorf("讲义引用不在原文中")
		}
	}
	return payload.Questions, nil
}

func callJSON(prompt string) (string, error) {
	if config.AIApiKey == "" {
		return "", fmt.Errorf("未配置 AI_API_KEY")
	}
	body, _ := json.Marshal(chatRequest{Model: config.AIModel, Messages: []chatMessage{{Role: "user", Content: prompt}}, Temperature: 0.2, ResponseFormat: map[string]string{"type": "json_object"}})
	base := strings.TrimRight(config.AIBaseURL, "/")
	if !strings.HasSuffix(base, "/chat/completions") {
		if !strings.HasSuffix(base, "/v1") {
			base += "/v1"
		}
		base += "/chat/completions"
	}
	req, err := http.NewRequest(http.MethodPost, base, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+config.AIApiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 90 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("AI 请求失败：%w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI 接口返回 %d：%s", resp.StatusCode, string(raw))
	}
	var result chatResponse
	if err = json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("AI 未返回内容")
	}
	return result.Choices[0].Message.Content, nil
}
func ValidateQuestion(q GeneratedQuestion) error {
	if q.Type != "single" && q.Type != "multiple" {
		return fmt.Errorf("题型错误")
	}
	if strings.TrimSpace(q.Content) == "" || strings.TrimSpace(q.Explanation) == "" || strings.TrimSpace(q.Citation) == "" {
		return fmt.Errorf("题目缺少题干、解析或讲义引用")
	}
	if len(q.Options) != 4 {
		return fmt.Errorf("每题必须有 4 个选项")
	}
	if q.Type == "single" && len(q.Answer) != 1 {
		return fmt.Errorf("单选题必须只有一个答案")
	}
	if q.Type == "multiple" && (len(q.Answer) < 2 || len(q.Answer) > 3) {
		return fmt.Errorf("多选题答案数量必须为 2-3")
	}
	ids := map[string]bool{}
	for _, o := range q.Options {
		if o.ID == "" || o.Content == "" {
			return fmt.Errorf("选项内容不完整")
		}
		if ids[o.ID] {
			return fmt.Errorf("选项 ID 重复")
		}
		ids[o.ID] = true
	}
	seenAnswers := map[string]bool{}
	for _, a := range q.Answer {
		if !ids[a] || seenAnswers[a] {
			return fmt.Errorf("答案不在选项中")
		}
		seenAnswers[a] = true
	}
	return nil
}
