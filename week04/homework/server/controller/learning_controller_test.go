package controller

import "testing"

func TestSameAnswers(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
		ok   bool
	}{
		{"same single", `["A"]`, `["A"]`, true},
		{"unordered multiple", `["C","A"]`, `["A","C"]`, true},
		{"partial multiple", `["A"]`, `["A","C"]`, false},
		{"extra answer", `["A","C","D"]`, `["A","C"]`, false},
		{"invalid json", ``, `["A"]`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameAnswers(tt.got, tt.want); got != tt.ok {
				t.Fatalf("sameAnswers(%s, %s) = %v, want %v", tt.got, tt.want, got, tt.ok)
			}
		})
	}
}
