package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"question-bank/config"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// StoreMaterial uploads source files to the private Supabase bucket.
func StoreMaterial(sessionID, filename string, content []byte) (string, error) {
	if config.SupabaseURL == "" || config.SupabaseServiceRoleKey == "" || strings.HasPrefix(config.SupabaseServiceRoleKey, "REPLACE_") {
		return "", fmt.Errorf("Supabase Storage 未配置：请填写 SUPABASE_SERVICE_ROLE_KEY")
	}
	bucket := config.SupabaseBucket
	if bucket == "" {
		bucket = "materials"
	}
	key := materialObjectKey(sessionID, filename, time.Now().UnixNano())
	objectURL := strings.TrimRight(config.SupabaseURL, "/") + "/storage/v1/object/" + url.PathEscape(bucket) + "/" + escapeObjectKey(key)
	req, err := http.NewRequest(http.MethodPost, objectURL, bytes.NewReader(content))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+config.SupabaseServiceRoleKey)
	req.Header.Set("apikey", config.SupabaseServiceRoleKey)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-upsert", "false")
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("上传资料到 Storage 失败：%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Storage 上传失败：%s", string(body))
	}
	return key, nil
}

func DownloadMaterial(objectKey string) ([]byte, error) {
	bucket := config.SupabaseBucket
	if bucket == "" {
		bucket = "materials"
	}
	objectURL := strings.TrimRight(config.SupabaseURL, "/") + "/storage/v1/object/authenticated/" + url.PathEscape(bucket) + "/" + escapeObjectKey(objectKey)
	req, err := http.NewRequest(http.MethodGet, objectURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+config.SupabaseServiceRoleKey)
	req.Header.Set("apikey", config.SupabaseServiceRoleKey)
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载资料失败：%w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Storage 下载失败（HTTP %d）", resp.StatusCode)
	}
	return body, nil
}

func ExtractText(filename string, content []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt", ".md":
		return string(content), nil
	case ".docx":
		return extractDOCX(content)
	case ".pdf":
		return extractPDF(content)
	default:
		return "", fmt.Errorf("不支持的资料格式")
	}
}

func extractDOCX(content []byte) (string, error) {
	z, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("DOCX 文件无效：%w", err)
	}
	for _, f := range z.File {
		if f.Name == "word/document.xml" {
			r, err := f.Open()
			if err != nil {
				return "", err
			}
			defer r.Close()
			dec := xml.NewDecoder(r)
			var b strings.Builder
			for {
				tok, e := dec.Token()
				if e == io.EOF {
					break
				}
				if e != nil {
					return "", e
				}
				switch t := tok.(type) {
				case xml.CharData:
					b.Write([]byte(t))
				case xml.EndElement:
					if t.Name.Local == "p" {
						b.WriteByte('\n')
					}
				}
			}
			return strings.TrimSpace(b.String()), nil
		}
	}
	return "", fmt.Errorf("DOCX 中未找到正文")
}

func extractPDF(content []byte) (string, error) {
	dir, err := os.MkdirTemp("", "study-agent-pdf-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	pdf := filepath.Join(dir, "source.pdf")
	if err = os.WriteFile(pdf, content, 0600); err != nil {
		return "", err
	}
	textFile := filepath.Join(dir, "output.txt")
	pdftotext := popplerCommand("pdftotext")
	pdftoppm := popplerCommand("pdftoppm")
	pdfinfo := popplerCommand("pdfinfo")
	if output, infoErr := exec.Command(pdfinfo, pdf).CombinedOutput(); infoErr == nil {
		matches := regexp.MustCompile(`(?m)^Pages:\s+(\d+)`).FindStringSubmatch(string(output))
		if len(matches) == 2 {
			pages, _ := strconv.Atoi(matches[1])
			if pages > 200 {
				return "", fmt.Errorf("PDF 不能超过 200 页（当前 %d 页）", pages)
			}
		}
	}
	if out, e := exec.Command(pdftotext, pdf, textFile).CombinedOutput(); e == nil {
		if text, e := os.ReadFile(textFile); e == nil && strings.TrimSpace(string(text)) != "" {
			return string(text), nil
		}
	} else {
		_ = out
	}
	if _, e := os.Stat(pdftoppm); e != nil && filepath.Base(pdftoppm) == pdftoppm {
		return "", fmt.Errorf("PDF 没有可提取文字，且未安装 Poppler（pdftoppm）用于 OCR")
	}
	if config.TesseractPath == "" {
		return "", fmt.Errorf("扫描 PDF 需要配置 TESSERACT_PATH")
	}
	prefix := filepath.Join(dir, "page")
	if out, e := exec.Command(pdftoppm, "-png", "-r", "200", pdf, prefix).CombinedOutput(); e != nil {
		return "", fmt.Errorf("PDF 转图片失败：%s", string(out))
	}
	images, _ := filepath.Glob(prefix + "-*.png")
	var b strings.Builder
	for _, image := range images {
		out, e := exec.Command(config.TesseractPath, image, "stdout", "-l", "chi_sim+eng").CombinedOutput()
		if e != nil {
			return "", fmt.Errorf("OCR 失败：%s", string(out))
		}
		b.Write(out)
		b.WriteByte('\n')
	}
	if strings.TrimSpace(b.String()) == "" {
		return "", fmt.Errorf("未从 PDF 提取到文字")
	}
	return b.String(), nil
}

func popplerCommand(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	if filepath.Separator == '\\' {
		roots := []string{config.PopplerBinPath, filepath.Join("D:\\Poppler", "Library", "bin"), filepath.Join("D:\\Release-25.07.0-0", "poppler-25.07.0", "Library", "bin")}
		for _, root := range roots {
			if root == "" {
				continue
			}
			candidate := filepath.Join(root, name+".exe")
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	return name
}

func materialObjectKey(sessionID, filename string, timestamp int64) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt", ".md", ".pdf", ".docx":
	default:
		ext = ""
	}
	return fmt.Sprintf("%s/%d%s", sessionID, timestamp, ext)
}

func escapeObjectKey(key string) string {
	parts := strings.Split(key, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}
