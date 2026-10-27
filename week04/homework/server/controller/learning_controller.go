package controller

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"question-bank/config"
	"question-bank/model"
	"question-bank/service"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const sessionCookie = "study_session"

var sessionExpirations sync.Map
var jobWake = make(chan struct{}, 1)
var materialLists sync.Map

type cachedMaterials struct {
	Items     []model.Material
	ExpiresAt time.Time
}

type examPlan struct {
	Title           string `json:"title"`
	QuestionCount   int    `json:"questionCount"`
	DurationMinutes int    `json:"durationMinutes"`
	Difficulty      string `json:"difficulty"`
	SingleCount     int    `json:"singleCount"`
	MultipleCount   int    `json:"multipleCount"`
	Focus           string `json:"focus"`
}

func sessionID(c *gin.Context) string { v, _ := c.Get("sessionID"); s, _ := v.(string); return s }
func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": data})
}
func fail(c *gin.Context, status int, msg string) { c.JSON(status, gin.H{"code": status, "msg": msg}) }

func BootstrapSession(c *gin.Context) {
	id, err := c.Cookie(sessionCookie)
	if err == nil {
		if sessionIsCached(id) {
			ok(c, gin.H{"id": id})
			return
		}
		var s model.AnonymousSession
		result := config.DB.First(&s, "id = ? AND expires_at > ?", id, time.Now())
		if result.Error == nil {
			sessionExpirations.Store(id, s.ExpiresAt)
			ok(c, gin.H{"id": id})
			return
		}
		if result.Error != gorm.ErrRecordNotFound {
			fail(c, http.StatusServiceUnavailable, "云数据库暂时不可用，请稍后重试")
			return
		}
	}
	id = newSessionID()
	expires := time.Now().AddDate(0, 0, 30)
	if err := config.DB.Create(&model.AnonymousSession{ID: id, ExpiresAt: expires}).Error; err != nil {
		fail(c, http.StatusServiceUnavailable, "创建匿名会话失败，请稍后重试")
		return
	}
	sessionExpirations.Store(id, expires)
	if config.CookieSecure {
		c.SetSameSite(http.SameSiteNoneMode)
	}
	c.SetCookie(sessionCookie, id, 30*24*3600, "/", "", config.CookieSecure, true)
	ok(c, gin.H{"id": id})
}

func newSessionID() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := c.Cookie(sessionCookie)
		if err != nil || id == "" {
			fail(c, http.StatusUnauthorized, "请先初始化匿名会话")
			c.Abort()
			return
		}
		if !sessionIsCached(id) {
			var s model.AnonymousSession
			result := config.DB.First(&s, "id = ? AND expires_at > ?", id, time.Now())
			if result.Error != nil {
				status := http.StatusUnauthorized
				msg := "会话已过期"
				if result.Error != gorm.ErrRecordNotFound {
					status = http.StatusServiceUnavailable
					msg = "云数据库暂时不可用，请稍后重试"
				}
				fail(c, status, msg)
				c.Abort()
				return
			}
			sessionExpirations.Store(id, s.ExpiresAt)
		}
		c.Set("sessionID", id)
		c.Next()
	}
}

func sessionIsCached(id string) bool {
	value, ok := sessionExpirations.Load(id)
	if !ok {
		return false
	}
	expiresAt, ok := value.(time.Time)
	if !ok || !expiresAt.After(time.Now()) {
		sessionExpirations.Delete(id)
		return false
	}
	return true
}

func UploadMaterial(c *gin.Context) {
	f, err := c.FormFile("file")
	if err != nil {
		fail(c, 400, "请选择资料文件")
		return
	}
	if f.Size > 50<<20 {
		fail(c, 400, "文件不能超过50MB")
		return
	}
	ext := strings.ToLower(filepath.Ext(f.Filename))
	if ext != ".txt" && ext != ".md" && ext != ".pdf" && ext != ".docx" {
		fail(c, 400, "仅支持 TXT、Markdown、PDF、DOCX")
		return
	}
	file, err := f.Open()
	if err != nil {
		fail(c, 500, "无法读取文件")
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		fail(c, 500, "无法读取文件")
		return
	}
	material := model.Material{AnonymousSessionID: sessionID(c), Name: f.Filename, ContentType: ext, Status: "processing"}
	if err := config.DB.Create(&material).Error; err != nil {
		fail(c, 500, "保存资料记录失败")
		return
	}
	objectKey, storeErr := service.StoreMaterial(sessionID(c), f.Filename, content)
	if storeErr != nil {
		material.Status = "failed"
		material.ErrorMessage = storeErr.Error()
		config.DB.Save(&material)
		fail(c, 502, storeErr.Error())
		return
	}
	material.ObjectKey = objectKey
	config.DB.Save(&material)
	job := model.GenerationJob{AnonymousSessionID: sessionID(c), Kind: "material_parse", Status: "queued", MaterialID: material.ID}
	if err := config.DB.Create(&job).Error; err != nil {
		fail(c, 500, "创建解析任务失败")
		return
	}
	cacheMaterialList(sessionID(c), nil)
	wakeJobWorker()
	ok(c, gin.H{"material": material, "jobId": job.ID})
}
func chapters(text string) string {
	var hs []string
	for _, l := range strings.Split(text, "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "#") {
			hs = append(hs, strings.TrimSpace(strings.TrimLeft(l, "#")))
		}
	}
	return strings.Join(hs, "\n")
}
func makeChunks(id uint, text string) {
	config.DB.Where("material_id = ?", id).Delete(&model.DocumentChunk{})
	rs := []rune(text)
	for i, n := 0, 0; i < len(rs); i, n = i+1200, n+1 {
		end := i + 1200
		if end > len(rs) {
			end = len(rs)
		}
		content := string(rs[i:end])
		embedding := service.VectorLiteral(service.LocalEmbedding(content))
		config.DB.Create(&model.DocumentChunk{MaterialID: id, Ordinal: n, Content: content, Embedding: embedding})
	}
}
func ListMaterials(c *gin.Context) {
	if cached, found := materialLists.Load(sessionID(c)); found {
		entry := cached.(cachedMaterials)
		if entry.ExpiresAt.After(time.Now()) {
			ok(c, entry.Items)
			return
		}
	}
	var ms []model.Material
	if err := config.DB.Where("anonymous_session_id = ?", sessionID(c)).Order("created_at DESC").Find(&ms).Error; err != nil {
		fail(c, http.StatusServiceUnavailable, "资料列表加载失败，请稍后重试")
		return
	}
	cacheMaterialList(sessionID(c), ms)
	ok(c, ms)
}

func cacheMaterialList(sessionID string, items []model.Material) {
	if items == nil {
		materialLists.Delete(sessionID)
		return
	}
	materialLists.Store(sessionID, cachedMaterials{Items: items, ExpiresAt: time.Now().Add(5 * time.Minute)})
}
func GetMaterial(c *gin.Context) {
	var m model.Material
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&m).Error != nil {
		fail(c, 404, "资料不存在")
		return
	}
	ok(c, m)
}
func GetJob(c *gin.Context) {
	var j model.GenerationJob
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&j).Error != nil {
		fail(c, 404, "任务不存在")
		return
	}
	ok(c, j)
}

func CreateConversation(c *gin.Context) {
	var req struct {
		MaterialIDs []uint `json:"materialIds"`
	}
	if c.ShouldBindJSON(&req) != nil || len(req.MaterialIDs) == 0 {
		fail(c, 400, "请选择至少一份资料")
		return
	}
	var count int64
	config.DB.Model(&model.Material{}).Where("id IN ? AND anonymous_session_id = ? AND status = ?", req.MaterialIDs, sessionID(c), "ready").Count(&count)
	if count != int64(len(req.MaterialIDs)) {
		fail(c, 400, "只能选择当前会话中已解析完成的资料")
		return
	}
	b, _ := json.Marshal(req.MaterialIDs)
	conv := model.AgentConversation{AnonymousSessionID: sessionID(c), MaterialIDs: string(b), Status: "collecting"}
	config.DB.Create(&conv)
	config.DB.Create(&model.AgentMessage{ConversationID: conv.ID, Role: "assistant", Content: "请告诉我试卷名称、题量、单选/多选比例、难度、覆盖章节或知识点，以及考试时长。"})
	ok(c, conv)
}
func AddMessage(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "消息格式错误")
		return
	}
	var conv model.AgentConversation
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&conv).Error != nil {
		fail(c, 404, "会话不存在")
		return
	}
	if conv.Status != "collecting" && conv.Status != "ready" {
		fail(c, 409, "当前会话不能继续修改")
		return
	}
	if err := config.DB.Create(&model.AgentMessage{ConversationID: conv.ID, Role: "user", Content: strings.TrimSpace(req.Content)}).Error; err != nil {
		fail(c, 500, "保存消息失败")
		return
	}
	var messages []model.AgentMessage
	config.DB.Where("conversation_id = ?", conv.ID).Order("created_at").Find(&messages)
	history := make([]string, 0, len(messages))
	for _, m := range messages {
		history = append(history, m.Role+": "+m.Content)
	}
	result, err := service.BuildExamPlan(history, conv.PlanJSON)
	if err != nil {
		fail(c, 502, err.Error())
		return
	}
	planJSON, _ := json.Marshal(result.Plan)
	conv.PlanJSON = string(planJSON)
	if result.Ready {
		conv.Status = "ready"
	} else {
		conv.Status = "collecting"
	}
	config.DB.Save(&conv)
	reply := result.Reply
	if reply == "" {
		reply = "请继续补充出卷要求。"
	}
	assistant := model.AgentMessage{ConversationID: conv.ID, Role: "assistant", Content: reply}
	config.DB.Create(&assistant)
	ok(c, gin.H{"plan": result.Plan, "status": conv.Status, "reply": reply, "missing": result.Missing})
}

func GetConversation(c *gin.Context) {
	var conv model.AgentConversation
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&conv).Error != nil {
		fail(c, 404, "会话不存在")
		return
	}
	var messages []model.AgentMessage
	config.DB.Where("conversation_id = ?", conv.ID).Order("created_at").Find(&messages)
	var plan examPlan
	if conv.PlanJSON != "" {
		_ = json.Unmarshal([]byte(conv.PlanJSON), &plan)
	}
	ok(c, gin.H{"conversation": conv, "messages": messages, "plan": plan})
}
func ConfirmConversation(c *gin.Context) {
	var conv model.AgentConversation
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&conv).Error != nil || conv.Status != "ready" {
		fail(c, 400, "会话尚未准备好")
		return
	}
	var p examPlan
	json.Unmarshal([]byte(conv.PlanJSON), &p)
	exam := model.Exam{AnonymousSessionID: sessionID(c), Title: p.Title, Status: "generating", QuestionCount: p.QuestionCount, DurationMinutes: p.DurationMinutes, MaterialIDs: conv.MaterialIDs}
	if err := config.DB.Create(&exam).Error; err != nil {
		fail(c, 500, "创建试卷失败")
		return
	}
	exam.SourceExamID = exam.ID
	config.DB.Model(&exam).Update("source_exam_id", exam.ID)
	job := model.GenerationJob{AnonymousSessionID: sessionID(c), Kind: "exam_generation", Status: "queued", ExamID: exam.ID}
	if err := config.DB.Create(&job).Error; err != nil {
		fail(c, 500, "创建生成任务失败")
		return
	}
	conv.Status = "confirmed"
	config.DB.Save(&conv)
	wakeJobWorker()
	ok(c, gin.H{"jobId": job.ID, "examId": exam.ID})
}
func generateExam(sid string, exam model.Exam, p *examPlan) error {
	var ids []uint
	json.Unmarshal([]byte(exam.MaterialIDs), &ids)
	var chunks []model.DocumentChunk
	queryText := p.Focus
	if strings.TrimSpace(queryText) == "" {
		queryText = p.Title + " " + p.Difficulty
	}
	if config.DatabaseURL != "" {
		vector := service.VectorLiteral(service.LocalEmbedding(queryText))
		config.DB.Raw("SELECT * FROM document_chunks WHERE material_id IN ? ORDER BY embedding <=> ?::vector LIMIT 12", ids, vector).Scan(&chunks)
	} else {
		config.DB.Where("material_id IN ?", ids).Order("material_id, ordinal").Limit(12).Find(&chunks)
	}
	if len(chunks) == 0 {
		return fmt.Errorf("选定资料尚未解析完成")
	}
	var context strings.Builder
	for _, chunk := range chunks {
		context.WriteString("\n---\n")
		context.WriteString(chunk.Content)
	}
	generated, err := service.GenerateExamQuestions(truncate(context.String(), 12000), p.Focus, p.Difficulty, p.SingleCount, p.MultipleCount)
	if err != nil {
		return err
	}
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("exam_id = ?", exam.ID).Delete(&model.ExamQuestion{}).Error; err != nil {
			return err
		}
		for i, item := range generated {
			options, _ := json.Marshal(item.Options)
			answer, _ := json.Marshal(item.Answer)
			q := model.ExamQuestion{ExamID: exam.ID, Ordinal: i + 1, Type: item.Type, Content: item.Content, Options: string(options), Answer: string(answer), Explanation: item.Explanation, Citation: item.Citation}
			if err := tx.Create(&q).Error; err != nil {
				return err
			}
		}
		exam.Status = "ready"
		return tx.Save(&exam).Error
	})
}

func StartWorkers() {
	config.DB.Model(&model.GenerationJob{}).Where("status = ? AND updated_at < ?", "running", time.Now().Add(-10*time.Minute)).Update("status", "queued")
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-jobWake:
			}
			for processNextJob() {
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			submitExpiredAttempts()
		}
	}()
	wakeJobWorker()
}

func wakeJobWorker() {
	select {
	case jobWake <- struct{}{}:
	default:
	}
}

func processNextJob() bool {
	var job model.GenerationJob
	if config.DB.Where("status = ?", "queued").Order("created_at").Limit(1).Find(&job).RowsAffected == 0 {
		return false
	}
	result := config.DB.Model(&model.GenerationJob{}).Where("id = ? AND status = ?", job.ID, "queued").Updates(map[string]interface{}{"status": "running", "progress": 10})
	if result.Error != nil || result.RowsAffected != 1 {
		return true
	}
	var err error
	switch job.Kind {
	case "material_parse":
		err = processMaterialJob(&job)
	case "exam_generation":
		err = processExamJob(&job)
	default:
		err = fmt.Errorf("未知任务类型：%s", job.Kind)
	}
	if err == nil {
		config.DB.Model(&job).Updates(map[string]interface{}{"status": "succeeded", "progress": 100, "error_message": ""})
		return true
	}
	if job.RetryCount < 1 {
		config.DB.Model(&job).Updates(map[string]interface{}{"status": "queued", "progress": 0, "retry_count": job.RetryCount + 1, "error_message": err.Error()})
		return true
	}
	config.DB.Model(&job).Updates(map[string]interface{}{"status": "failed", "progress": 100, "error_message": err.Error()})
	if job.MaterialID != 0 {
		config.DB.Model(&model.Material{}).Where("id = ?", job.MaterialID).Updates(map[string]interface{}{"status": "failed", "error_message": err.Error()})
		cacheMaterialList(job.AnonymousSessionID, nil)
	}
	if job.ExamID != 0 {
		config.DB.Model(&model.Exam{}).Where("id = ?", job.ExamID).Update("status", "failed")
	}
	return true
}

func processMaterialJob(job *model.GenerationJob) error {
	var material model.Material
	if err := config.DB.First(&material, job.MaterialID).Error; err != nil {
		return err
	}
	config.DB.Model(job).Update("progress", 30)
	content, err := service.DownloadMaterial(material.ObjectKey)
	if err != nil {
		return err
	}
	config.DB.Model(job).Update("progress", 55)
	extracted, err := service.ExtractText(material.Name, content)
	if err != nil {
		return err
	}
	if len([]rune(strings.TrimSpace(extracted))) < 20 {
		return fmt.Errorf("资料正文过短，无法出题")
	}
	config.DB.Model(job).Update("progress", 80)
	makeChunks(material.ID, extracted)
	err = config.DB.Model(&material).Updates(map[string]interface{}{"text": extracted, "chapter_summary": chapters(extracted), "status": "ready", "error_message": ""}).Error
	cacheMaterialList(job.AnonymousSessionID, nil)
	return err
}

func processExamJob(job *model.GenerationJob) error {
	var exam model.Exam
	if err := config.DB.Where("id = ? AND anonymous_session_id = ?", job.ExamID, job.AnonymousSessionID).First(&exam).Error; err != nil {
		return err
	}
	var conv model.AgentConversation
	if err := config.DB.Where("anonymous_session_id = ? AND material_ids = ? AND status = ?", job.AnonymousSessionID, exam.MaterialIDs, "confirmed").Order("created_at DESC").First(&conv).Error; err != nil {
		return err
	}
	var p examPlan
	if err := json.Unmarshal([]byte(conv.PlanJSON), &p); err != nil {
		return err
	}
	config.DB.Model(job).Update("progress", 35)
	if err := generateExam(job.AnonymousSessionID, exam, &p); err != nil {
		return err
	}
	config.DB.Model(job).Update("progress", 90)
	return nil
}

func submitExpiredAttempts() {
	var attempts []model.ExamAttempt
	config.DB.Where("status = ? AND expires_at <= ?", "in_progress", time.Now()).Limit(100).Find(&attempts)
	for i := range attempts {
		gradeAttempt(&attempts[i], "timeout")
	}
}
func truncate(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) > n {
		return string(r[:n]) + "..."
	}
	return string(r)
}

func ListExams(c *gin.Context) {
	var xs []model.Exam
	if err := config.DB.Where("anonymous_session_id = ?", sessionID(c)).Order("created_at DESC").Find(&xs).Error; err != nil {
		fail(c, http.StatusServiceUnavailable, "试卷列表加载失败，请稍后重试")
		return
	}
	type examItem struct {
		model.Exam
		AttemptCount int64              `json:"attemptCount"`
		Latest       *model.ExamAttempt `json:"latestAttempt,omitempty"`
	}
	var attempts []model.ExamAttempt
	if err := config.DB.Where("anonymous_session_id = ?", sessionID(c)).Order("started_at DESC").Find(&attempts).Error; err != nil {
		fail(c, http.StatusServiceUnavailable, "作答记录加载失败，请稍后重试")
		return
	}
	counts := make(map[uint]int64, len(xs))
	latest := make(map[uint]*model.ExamAttempt, len(xs))
	for i := range attempts {
		attempt := &attempts[i]
		counts[attempt.ExamID]++
		if latest[attempt.ExamID] == nil {
			latest[attempt.ExamID] = attempt
		}
	}
	items := make([]examItem, 0, len(xs))
	for _, exam := range xs {
		items = append(items, examItem{Exam: exam, AttemptCount: counts[exam.ID], Latest: latest[exam.ID]})
	}
	ok(c, items)
}
func GetExam(c *gin.Context) {
	var x model.Exam
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&x).Error != nil {
		fail(c, 404, "试卷不存在")
		return
	}
	var attempts []model.ExamAttempt
	config.DB.Where("exam_id = ? AND anonymous_session_id = ?", x.ID, sessionID(c)).Order("started_at DESC").Find(&attempts)
	ok(c, gin.H{"exam": x, "attempts": attempts})
}
func StartExam(c *gin.Context) {
	var x model.Exam
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&x).Error != nil || x.Status != "ready" {
		fail(c, 400, "试卷尚不可开始")
		return
	}
	var active model.ExamAttempt
	if config.DB.Where("exam_id = ? AND anonymous_session_id = ? AND status = ? AND expires_at > ?", x.ID, sessionID(c), "in_progress", time.Now()).Order("started_at DESC").First(&active).Error == nil {
		ok(c, active)
		return
	}
	now := time.Now()
	a := model.ExamAttempt{AnonymousSessionID: sessionID(c), ExamID: x.ID, Status: "in_progress", StartedAt: now, ExpiresAt: now.Add(time.Duration(x.DurationMinutes) * time.Minute)}
	config.DB.Create(&a)
	ok(c, a)
}
func GetAttemptQuestions(c *gin.Context) {
	var a model.ExamAttempt
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&a).Error != nil {
		fail(c, 404, "答卷不存在")
		return
	}
	if a.Status != "in_progress" || time.Now().After(a.ExpiresAt) {
		if a.Status == "in_progress" {
			gradeAttempt(&a, "timeout")
		}
		fail(c, 409, "答卷已结束，请查看结果")
		return
	}
	var qs []model.ExamQuestion
	config.DB.Where("exam_id = ?", a.ExamID).Order("ordinal").Find(&qs)
	var answers []model.AttemptAnswer
	config.DB.Where("attempt_id = ?", a.ID).Find(&answers)
	saved := map[uint]string{}
	for _, v := range answers {
		saved[v.ExamQuestionID] = v.Answer
	}
	out := make([]gin.H, 0, len(qs))
	for _, q := range qs {
		answer := saved[q.ID]
		if answer == "" {
			answer = "[]"
		}
		marked := false
		for _, item := range answers {
			if item.ExamQuestionID == q.ID {
				marked = item.Marked
				break
			}
		}
		out = append(out, gin.H{"id": q.ID, "ordinal": q.Ordinal, "type": q.Type, "content": q.Content, "options": json.RawMessage(q.Options), "savedAnswer": json.RawMessage(answer), "marked": marked})
	}
	ok(c, gin.H{"expiresAt": a.ExpiresAt, "questions": out})
}
func SaveAnswer(c *gin.Context) {
	type answerContext struct {
		AttemptID  uint
		ExpiresAt  time.Time
		Status     string
		QuestionID uint
		Type       string
		Options    string
	}
	var ctx answerContext
	err := config.DB.Table("exam_attempts AS a").
		Select("a.id AS attempt_id, a.expires_at, a.status, q.id AS question_id, q.type, q.options").
		Joins("JOIN exam_questions AS q ON q.exam_id = a.exam_id AND q.id = ?", c.Param("questionId")).
		Where("a.id = ? AND a.anonymous_session_id = ?", c.Param("id"), sessionID(c)).
		Scan(&ctx).Error
	if err != nil || ctx.AttemptID == 0 || ctx.Status != "in_progress" {
		fail(c, 409, "答卷已结束")
		return
	}
	if time.Now().After(ctx.ExpiresAt) {
		var attempt model.ExamAttempt
		config.DB.First(&attempt, ctx.AttemptID)
		gradeAttempt(&attempt, "timeout")
		fail(c, 409, "已超时")
		return
	}
	var req struct {
		Answer []string `json:"answer"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "答案格式错误")
		return
	}
	var options []service.Option
	if err := json.Unmarshal([]byte(ctx.Options), &options); err != nil {
		fail(c, 500, "题目选项数据损坏")
		return
	}
	valid := make(map[string]bool, len(options))
	for _, option := range options {
		valid[option.ID] = true
	}
	seen := map[string]bool{}
	for _, choice := range req.Answer {
		if !valid[choice] || seen[choice] {
			fail(c, 400, "答案包含无效或重复的选项")
			return
		}
		seen[choice] = true
	}
	if (ctx.Type == "single" && len(req.Answer) > 1) || (ctx.Type == "multiple" && len(req.Answer) > 3) {
		fail(c, 400, "答案数量不符合题型要求")
		return
	}
	b, _ := json.Marshal(req.Answer)
	ans := model.AttemptAnswer{AttemptID: ctx.AttemptID, ExamQuestionID: ctx.QuestionID, Answer: string(b)}
	if err := config.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "attempt_id"}, {Name: "exam_question_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"answer": string(b), "updated_at": time.Now()}),
	}).Create(&ans).Error; err != nil {
		fail(c, http.StatusServiceUnavailable, "答案保存失败，请重试")
		return
	}
	ok(c, gin.H{"saved": true})
}

func MarkQuestion(c *gin.Context) {
	type markContext struct {
		AttemptID  uint
		QuestionID uint
		ExpiresAt  time.Time
	}
	var ctx markContext
	err := config.DB.Table("exam_attempts AS a").
		Select("a.id AS attempt_id, a.expires_at, q.id AS question_id").
		Joins("JOIN exam_questions AS q ON q.exam_id = a.exam_id AND q.id = ?", c.Param("questionId")).
		Where("a.id = ? AND a.anonymous_session_id = ? AND a.status = ?", c.Param("id"), sessionID(c), "in_progress").
		Scan(&ctx).Error
	if err != nil || ctx.AttemptID == 0 || time.Now().After(ctx.ExpiresAt) {
		fail(c, 409, "答卷已结束")
		return
	}
	var req struct {
		Marked bool `json:"marked"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "标记格式错误")
		return
	}
	answer := model.AttemptAnswer{AttemptID: ctx.AttemptID, ExamQuestionID: ctx.QuestionID, Answer: "[]", Marked: req.Marked}
	if err := config.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "attempt_id"}, {Name: "exam_question_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"marked": req.Marked, "updated_at": time.Now()}),
	}).Create(&answer).Error; err != nil {
		fail(c, http.StatusServiceUnavailable, "标记保存失败，请重试")
		return
	}
	ok(c, gin.H{"marked": req.Marked})
}
func SubmitAttempt(c *gin.Context) {
	var a model.ExamAttempt
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&a).Error != nil {
		fail(c, 404, "答卷不存在")
		return
	}
	if a.Status == "in_progress" {
		reason := "manual"
		if time.Now().After(a.ExpiresAt) {
			reason = "timeout"
		}
		gradeAttempt(&a, reason)
	}
	ok(c, gin.H{"attemptId": a.ID, "status": a.Status})
}
func gradeAttempt(a *model.ExamAttempt, reason string) {
	if a.Status != "in_progress" {
		return
	}
	var qs []model.ExamQuestion
	config.DB.Where("exam_id = ?", a.ExamID).Find(&qs)
	var answers []model.AttemptAnswer
	config.DB.Where("attempt_id = ?", a.ID).Find(&answers)
	m := map[uint]*model.AttemptAnswer{}
	for i := range answers {
		m[answers[i].ExamQuestionID] = &answers[i]
	}
	right := 0
	correctAnswerIDs := make([]uint, 0, len(qs))
	wrongQuestions := make([]model.WrongQuestion, 0, len(qs))
	for _, q := range qs {
		answer := m[q.ID]
		correct := answer != nil && sameAnswers(answer.Answer, q.Answer)
		if correct {
			right++
			correctAnswerIDs = append(correctAnswerIDs, answer.ID)
		} else {
			wrongQuestions = append(wrongQuestions, model.WrongQuestion{AnonymousSessionID: a.AnonymousSessionID, ExamQuestionID: q.ID, AttemptID: a.ID})
		}
	}
	now := time.Now()
	a.Status = "submitted"
	a.SubmittedAt = &now
	a.SubmitReason = reason
	if len(qs) > 0 {
		a.Score = float64(right) * 100 / float64(len(qs))
	}
	config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AttemptAnswer{}).Where("attempt_id = ?", a.ID).Update("is_correct", false).Error; err != nil {
			return err
		}
		if len(correctAnswerIDs) > 0 {
			if err := tx.Model(&model.AttemptAnswer{}).Where("id IN ?", correctAnswerIDs).Update("is_correct", true).Error; err != nil {
				return err
			}
		}
		if len(wrongQuestions) > 0 {
			if err := tx.Create(&wrongQuestions).Error; err != nil {
				return err
			}
		}
		return tx.Save(a).Error
	})
}
func sameAnswers(got, want string) bool {
	var a, b []string
	json.Unmarshal([]byte(got), &a)
	json.Unmarshal([]byte(want), &b)
	if len(a) != len(b) {
		return false
	}
	ma := map[string]bool{}
	for _, x := range a {
		ma[x] = true
	}
	for _, x := range b {
		if !ma[x] {
			return false
		}
	}
	return true
}
func AttemptResult(c *gin.Context) {
	var a model.ExamAttempt
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&a).Error != nil {
		fail(c, 404, "答卷不存在")
		return
	}
	if a.Status == "in_progress" {
		if time.Now().After(a.ExpiresAt) {
			gradeAttempt(&a, "timeout")
		} else {
			fail(c, 409, "请先交卷")
			return
		}
	}
	var qs []model.ExamQuestion
	config.DB.Where("exam_id = ?", a.ExamID).Order("ordinal").Find(&qs)
	var as []model.AttemptAnswer
	config.DB.Where("attempt_id = ?", a.ID).Find(&as)
	m := map[uint]model.AttemptAnswer{}
	for _, v := range as {
		m[v.ExamQuestionID] = v
	}
	items := []gin.H{}
	for _, q := range qs {
		v := m[q.ID]
		userAnswer := v.Answer
		if userAnswer == "" {
			userAnswer = "[]"
		}
		items = append(items, gin.H{"id": q.ID, "ordinal": q.Ordinal, "type": q.Type, "content": q.Content, "options": json.RawMessage(q.Options), "userAnswer": json.RawMessage(userAnswer), "correctAnswer": json.RawMessage(q.Answer), "isCorrect": v.IsCorrect, "explanation": q.Explanation, "citation": q.Citation})
	}
	ok(c, gin.H{"attempt": a, "questions": items})
}
func RetryWrong(c *gin.Context) {
	var old model.ExamAttempt
	if config.DB.Where("id = ? AND anonymous_session_id = ?", c.Param("id"), sessionID(c)).First(&old).Error != nil || old.Status != "submitted" {
		fail(c, 400, "请先完成原试卷")
		return
	}
	var wrong []model.WrongQuestion
	config.DB.Where("attempt_id = ? AND anonymous_session_id = ?", old.ID, sessionID(c)).Find(&wrong)
	if len(wrong) == 0 {
		fail(c, 400, "没有错题")
		return
	}
	var original model.Exam
	config.DB.First(&original, old.ExamID)
	rootID := original.SourceExamID
	if rootID == 0 {
		rootID = original.ID
	}
	exam := model.Exam{AnonymousSessionID: sessionID(c), Title: original.Title + " - 错题重练", Status: "ready", QuestionCount: len(wrong), DurationMinutes: original.DurationMinutes, SourceExamID: rootID}
	config.DB.Create(&exam)
	for i, w := range wrong {
		var q model.ExamQuestion
		config.DB.First(&q, w.ExamQuestionID)
		q.ID = 0
		q.ExamID = exam.ID
		q.Ordinal = i + 1
		config.DB.Create(&q)
	}
	ok(c, exam)
}

func ListWrongQuestions(c *gin.Context) {
	type wrongItem struct {
		ID          uint      `json:"id"`
		AttemptID   uint      `json:"attemptId"`
		Content     string    `json:"content"`
		Type        string    `json:"type"`
		Options     string    `json:"-"`
		Answer      string    `json:"-"`
		Explanation string    `json:"explanation"`
		Citation    string    `json:"citation"`
		CreatedAt   time.Time `json:"createdAt"`
	}
	var rows []wrongItem
	config.DB.Table("wrong_questions AS w").
		Select("w.id, w.attempt_id, w.created_at, q.content, q.type, q.options, q.answer, q.explanation, q.citation").
		Joins("JOIN exam_questions AS q ON q.id = w.exam_question_id").
		Where("w.anonymous_session_id = ?", sessionID(c)).Order("w.created_at DESC").Scan(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{"id": row.ID, "attemptId": row.AttemptID, "content": row.Content, "type": row.Type, "options": json.RawMessage(row.Options), "correctAnswer": json.RawMessage(row.Answer), "explanation": row.Explanation, "citation": row.Citation, "createdAt": row.CreatedAt})
	}
	ok(c, items)
}

func ListWrongBooks(c *gin.Context) {
	type attemptSummary struct {
		ID          uint       `json:"id"`
		ExamID      uint       `json:"examId"`
		Score       float64    `json:"score"`
		SubmittedAt *time.Time `json:"submittedAt"`
		WrongCount  int64      `json:"wrongCount"`
	}
	type book struct {
		ExamID          uint             `json:"examId"`
		Title           string           `json:"title"`
		LatestAttemptID uint             `json:"latestAttemptId"`
		LastPracticedAt *time.Time       `json:"lastPracticedAt"`
		Attempts        []attemptSummary `json:"attempts"`
		TotalWrong      int64            `json:"totalWrong"`
	}
	type wrongBookRow struct {
		AttemptID   uint       `gorm:"column:attempt_id"`
		ExamID      uint       `gorm:"column:exam_id"`
		RootExamID  uint       `gorm:"column:root_exam_id"`
		RootTitle   string     `gorm:"column:root_title"`
		Score       float64    `gorm:"column:score"`
		SubmittedAt *time.Time `gorm:"column:submitted_at"`
		WrongCount  int64      `gorm:"column:wrong_count"`
	}
	var rows []wrongBookRow
	err := config.DB.Raw(`
		SELECT a.id AS attempt_id, a.exam_id, root.id AS root_exam_id,
		       root.title AS root_title, a.score, a.submitted_at,
		       COALESCE(w.wrong_count, 0) AS wrong_count
		FROM exam_attempts AS a
		JOIN exams AS e ON e.id = a.exam_id AND e.anonymous_session_id = a.anonymous_session_id
		JOIN exams AS root ON root.id = CASE WHEN e.source_exam_id = 0 THEN e.id ELSE e.source_exam_id END
		                   AND root.anonymous_session_id = a.anonymous_session_id
		LEFT JOIN (
			SELECT attempt_id, COUNT(*) AS wrong_count
			FROM wrong_questions
			WHERE anonymous_session_id = ?
			GROUP BY attempt_id
		) AS w ON w.attempt_id = a.id
		WHERE a.anonymous_session_id = ? AND a.status = ?
		ORDER BY a.submitted_at DESC`, sessionID(c), sessionID(c), "submitted").Scan(&rows).Error
	if err != nil {
		fail(c, http.StatusServiceUnavailable, "错题本加载失败，请稍后重试")
		return
	}
	groups := map[uint]*book{}
	for _, row := range rows {
		group := groups[row.RootExamID]
		if group == nil {
			group = &book{ExamID: row.RootExamID, Title: row.RootTitle, Attempts: []attemptSummary{}}
			groups[row.RootExamID] = group
		}
		group.Attempts = append(group.Attempts, attemptSummary{ID: row.AttemptID, ExamID: row.ExamID, Score: row.Score, SubmittedAt: row.SubmittedAt, WrongCount: row.WrongCount})
		group.TotalWrong += row.WrongCount
		if group.LatestAttemptID == 0 {
			group.LatestAttemptID = row.AttemptID
			group.LastPracticedAt = row.SubmittedAt
		}
	}
	result := make([]book, 0, len(groups))
	for _, group := range groups {
		if group.TotalWrong > 0 {
			result = append(result, *group)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].LastPracticedAt == nil {
			return false
		}
		if result[j].LastPracticedAt == nil {
			return true
		}
		return result[i].LastPracticedAt.After(*result[j].LastPracticedAt)
	})
	ok(c, result)
}
