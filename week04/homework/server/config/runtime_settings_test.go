package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateEnvFilePreservesAndUpdatesValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	initial := "# keep comment\nAI_MODEL=old-model\nOTHER=value\n"
	if err := os.WriteFile(path, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}
	if err := updateEnvFile(path, map[string]string{"AI_MODEL": "new-model", "AI_BASE_URL": "https://example.com"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	for _, expected := range []string{"# keep comment", "AI_MODEL=new-model", "OTHER=value", "AI_BASE_URL=https://example.com"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected %q in %q", expected, content)
		}
	}
}

func TestMaskedAPIKey(t *testing.T) {
	original := AIApiKey
	t.Cleanup(func() { AIApiKey = original })
	AIApiKey = "sk-1234567890abcd"
	masked := MaskedAPIKey()
	if strings.Contains(masked, "1234567890") || !strings.HasPrefix(masked, "sk-") || !strings.HasSuffix(masked, "abcd") {
		t.Fatalf("key was not safely masked: %q", masked)
	}
}
