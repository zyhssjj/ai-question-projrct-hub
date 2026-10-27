package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var settingsMu sync.Mutex

func UpdateAISettings(apiKey, baseURL, model string) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	model = strings.TrimSpace(model)
	apiKey = strings.TrimSpace(apiKey)
	if baseURL == "" || model == "" {
		return fmt.Errorf("Base URL 和模型名称不能为空")
	}
	if strings.ContainsAny(apiKey+baseURL+model, "\r\n") {
		return fmt.Errorf("配置内容不能包含换行符")
	}
	if !strings.HasPrefix(baseURL, "https://") && !strings.HasPrefix(baseURL, "http://") {
		return fmt.Errorf("Base URL 必须以 http:// 或 https:// 开头")
	}
	values := map[string]string{"AI_BASE_URL": baseURL, "AI_MODEL": model}
	if apiKey != "" {
		values["AI_API_KEY"] = apiKey
	}
	if err := updateEnvFile(".env", values); err != nil {
		return err
	}
	for key, value := range values {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	AIBaseURL = baseURL
	AIModel = model
	if value, ok := values["AI_API_KEY"]; ok {
		AIApiKey = value
	}
	return nil
}

func MaskedAPIKey() string {
	key := strings.TrimSpace(AIApiKey)
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:3] + strings.Repeat("*", len(key)-7) + key[len(key)-4:]
}

func updateEnvFile(path string, values map[string]string) error {
	lines := []string{}
	if file, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		file.Close()
		if err := scanner.Err(); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	seen := make(map[string]bool, len(values))
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.Contains(line, "=") {
			continue
		}
		key := strings.TrimSpace(strings.SplitN(line, "=", 2)[0])
		if value, ok := values[key]; ok {
			lines[i] = key + "=" + value
			seen[key] = true
		}
	}
	for key, value := range values {
		if !seen[key] {
			lines = append(lines, key+"="+value)
		}
	}
	temporary := path + ".tmp"
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(temporary, []byte(content), 0600); err != nil {
		return err
	}
	return os.Rename(temporary, filepath.Clean(path))
}
