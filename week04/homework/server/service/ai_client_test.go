package service

import "testing"

func validQuestion() GeneratedQuestion {
	return GeneratedQuestion{
		Type: "single", Content: "题干", Explanation: "解析", Citation: "讲义原文",
		Options: []Option{{ID: "A", Content: "甲"}, {ID: "B", Content: "乙"}, {ID: "C", Content: "丙"}, {ID: "D", Content: "丁"}},
		Answer:  []string{"A"},
	}
}

func TestValidateQuestion(t *testing.T) {
	if err := ValidateQuestion(validQuestion()); err != nil {
		t.Fatalf("valid question rejected: %v", err)
	}
	duplicateOption := validQuestion()
	duplicateOption.Options[1].ID = "A"
	if err := ValidateQuestion(duplicateOption); err == nil {
		t.Fatal("duplicate option ID should be rejected")
	}
	multiple := validQuestion()
	multiple.Type = "multiple"
	multiple.Answer = []string{"A"}
	if err := ValidateQuestion(multiple); err == nil {
		t.Fatal("multiple choice with one answer should be rejected")
	}
	duplicateAnswer := validQuestion()
	duplicateAnswer.Type = "multiple"
	duplicateAnswer.Answer = []string{"A", "A"}
	if err := ValidateQuestion(duplicateAnswer); err == nil {
		t.Fatal("duplicate answers should be rejected")
	}
}
