package service

import (
	"math"
	"testing"
)

func TestLocalEmbedding(t *testing.T) {
	first := LocalEmbedding("计算机网络 TCP 拥塞控制")
	second := LocalEmbedding("计算机网络 TCP 拥塞控制")
	if len(first) != EmbeddingDimensions || len(second) != EmbeddingDimensions {
		t.Fatalf("unexpected dimensions: %d, %d", len(first), len(second))
	}
	var norm float64
	for i := range first {
		if first[i] != second[i] {
			t.Fatal("embedding must be deterministic")
		}
		norm += float64(first[i] * first[i])
	}
	if math.Abs(math.Sqrt(norm)-1) > 0.0001 {
		t.Fatalf("embedding is not normalized: %f", math.Sqrt(norm))
	}
}
