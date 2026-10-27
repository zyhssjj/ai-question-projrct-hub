package service

import (
	"fmt"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

const EmbeddingDimensions = 384

// LocalEmbedding uses normalized feature hashing. It adds no external model
// dependency and gives deterministic keyword-aware retrieval for Chinese text.
func LocalEmbedding(text string) []float32 {
	v := make([]float32, EmbeddingDimensions)
	tokens := tokenize(text)
	for _, token := range tokens {
		h := fnv.New32a()
		_, _ = h.Write([]byte(token))
		n := h.Sum32()
		idx := int(n % EmbeddingDimensions)
		sign := float32(1)
		if n&1 == 1 {
			sign = -1
		}
		v[idx] += sign
	}
	var sum float64
	for _, x := range v {
		sum += float64(x * x)
	}
	if sum > 0 {
		norm := float32(math.Sqrt(sum))
		for i := range v {
			v[i] /= norm
		}
	}
	return v
}

func VectorLiteral(v []float32) string {
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = fmt.Sprintf("%.7f", x)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func tokenize(text string) []string {
	r := []rune(strings.ToLower(text))
	result := make([]string, 0, len(r))
	var word []rune
	flush := func() {
		if len(word) > 0 {
			result = append(result, string(word))
			word = nil
		}
	}
	for i, ch := range r {
		if unicode.Is(unicode.Han, ch) {
			flush()
			result = append(result, string(ch))
			if i+1 < len(r) && unicode.Is(unicode.Han, r[i+1]) {
				result = append(result, string([]rune{ch, r[i+1]}))
			}
		} else if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			word = append(word, ch)
		} else {
			flush()
		}
	}
	flush()
	return result
}
