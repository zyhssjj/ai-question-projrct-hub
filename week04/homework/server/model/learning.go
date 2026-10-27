package model

import "time"

type AnonymousSession struct {
	ID        string    `json:"id" gorm:"primaryKey;size:64"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserProfile struct {
	AnonymousSessionID string    `json:"-" gorm:"primaryKey;size:64"`
	Name               string    `json:"name" gorm:"size:80;not null"`
	Account            string    `json:"account" gorm:"size:80;uniqueIndex;not null"`
	PasswordHash       string    `json:"-" gorm:"size:255"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type Material struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	AnonymousSessionID string    `json:"-" gorm:"index;not null"`
	Name               string    `json:"name"`
	ContentType        string    `json:"contentType"`
	ObjectKey          string    `json:"objectKey"`
	Status             string    `json:"status" gorm:"index"`
	ErrorMessage       string    `json:"errorMessage,omitempty"`
	Text               string    `json:"-" gorm:"type:text"`
	ChapterSummary     string    `json:"chapterSummary" gorm:"type:text"`
	CreatedAt          time.Time `json:"createdAt"`
}

type DocumentChunk struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	MaterialID uint   `json:"materialId" gorm:"index"`
	Ordinal    int    `json:"ordinal"`
	Chapter    string `json:"chapter"`
	Content    string `json:"content" gorm:"type:text"`
	Embedding  string `json:"-" gorm:"type:vector(384)"`
}

type GenerationJob struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	AnonymousSessionID string    `json:"-" gorm:"index;not null"`
	Kind               string    `json:"kind"`
	Status             string    `json:"status" gorm:"index"`
	Progress           int       `json:"progress"`
	ErrorMessage       string    `json:"errorMessage,omitempty"`
	ExamID             uint      `json:"examId,omitempty"`
	MaterialID         uint      `json:"materialId,omitempty"`
	RetryCount         int       `json:"retryCount"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type AgentConversation struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	AnonymousSessionID string    `json:"-" gorm:"index;not null"`
	MaterialIDs        string    `json:"materialIds" gorm:"type:text"`
	PlanJSON           string    `json:"plan" gorm:"type:text"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
}

type AgentMessage struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ConversationID uint      `json:"conversationId" gorm:"index"`
	Role           string    `json:"role"`
	Content        string    `json:"content" gorm:"type:text"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Exam struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	AnonymousSessionID string    `json:"-" gorm:"index;not null"`
	Title              string    `json:"title"`
	Status             string    `json:"status"`
	QuestionCount      int       `json:"questionCount"`
	DurationMinutes    int       `json:"durationMinutes"`
	MaterialIDs        string    `json:"materialIds" gorm:"type:text"`
	SourceExamID       uint      `json:"sourceExamId" gorm:"index"`
	CreatedAt          time.Time `json:"createdAt"`
}

type ExamQuestion struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	ExamID      uint   `json:"examId" gorm:"index"`
	Ordinal     int    `json:"ordinal"`
	Type        string `json:"type"`
	Content     string `json:"content" gorm:"type:text"`
	Options     string `json:"options" gorm:"type:text"`
	Answer      string `json:"-" gorm:"type:text"`
	Explanation string `json:"-" gorm:"type:text"`
	Citation    string `json:"-" gorm:"type:text"`
}

type ExamAttempt struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	AnonymousSessionID string     `json:"-" gorm:"index;not null"`
	ExamID             uint       `json:"examId" gorm:"index"`
	Status             string     `json:"status"`
	StartedAt          time.Time  `json:"startedAt"`
	ExpiresAt          time.Time  `json:"expiresAt"`
	SubmittedAt        *time.Time `json:"submittedAt,omitempty"`
	SubmitReason       string     `json:"submitReason,omitempty"`
	Score              float64    `json:"score"`
}

type AttemptAnswer struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	AttemptID      uint      `json:"attemptId" gorm:"uniqueIndex:attempt_question"`
	ExamQuestionID uint      `json:"examQuestionId" gorm:"uniqueIndex:attempt_question"`
	Answer         string    `json:"answer" gorm:"type:text"`
	IsCorrect      bool      `json:"isCorrect"`
	Marked         bool      `json:"marked"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type WrongQuestion struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	AnonymousSessionID string    `json:"-" gorm:"index;not null"`
	ExamQuestionID     uint      `json:"examQuestionId" gorm:"index"`
	AttemptID          uint      `json:"attemptId"`
	CreatedAt          time.Time `json:"createdAt"`
}
