package service

import (
	"regexp"
	"strings"
	"testing"
)

func TestMaterialObjectKeyDoesNotContainOriginalUnicodeFilename(t *testing.T) {
	key := materialObjectKey("session123", "2026-张雨豪-信息管理与信息系统.pdf", 123456789)

	if key != "session123/123456789.pdf" {
		t.Fatalf("unexpected object key: %q", key)
	}
	if strings.Contains(key, "张雨豪") {
		t.Fatal("object key must not contain the original Unicode filename")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9._/-]+$`).MatchString(key) {
		t.Fatalf("object key is not Storage-safe ASCII: %q", key)
	}
}

func TestMaterialObjectKeyDropsUnsupportedExtension(t *testing.T) {
	key := materialObjectKey("session123", "notes.exe", 123456789)
	if key != "session123/123456789" {
		t.Fatalf("unexpected object key: %q", key)
	}
}
