package controller

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"question-bank/config"
	"question-bank/model"
	"question-bank/service"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func currentProfile(c *gin.Context) (*model.UserProfile, error) {
	sid := sessionID(c)
	profile := model.UserProfile{AnonymousSessionID: sid}
	accountSuffix := sid
	if len(accountSuffix) > 8 {
		accountSuffix = accountSuffix[:8]
	}
	account := "rain-" + accountSuffix
	err := config.DB.Where("anonymous_session_id = ?", sid).
		Attrs(model.UserProfile{Name: "张雨豪", Account: account}).FirstOrCreate(&profile).Error
	return &profile, err
}

func GetProfile(c *gin.Context) {
	profile, err := currentProfile(c)
	if err != nil {
		log.Printf("加载个人资料失败: %v", err)
		fail(c, http.StatusServiceUnavailable, "个人资料加载失败")
		return
	}
	ok(c, gin.H{"name": profile.Name, "account": profile.Account, "hasPassword": profile.PasswordHash != "", "createdAt": profile.CreatedAt})
}

func UpdateProfile(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Name) == "" || len([]rune(req.Name)) > 40 {
		fail(c, 400, "姓名不能为空且不能超过 40 个字符")
		return
	}
	profile, err := currentProfile(c)
	if err != nil || config.DB.Model(profile).Update("name", strings.TrimSpace(req.Name)).Error != nil {
		fail(c, http.StatusServiceUnavailable, "个人资料保存失败")
		return
	}
	ok(c, gin.H{"name": strings.TrimSpace(req.Name), "account": profile.Account})
}

func ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if c.ShouldBindJSON(&req) != nil || len(req.NewPassword) < 8 || len(req.NewPassword) > 72 {
		fail(c, 400, "新密码长度必须为 8-72 位")
		return
	}
	profile, err := currentProfile(c)
	if err != nil {
		fail(c, http.StatusServiceUnavailable, "账号信息加载失败")
		return
	}
	if profile.PasswordHash != "" && bcrypt.CompareHashAndPassword([]byte(profile.PasswordHash), []byte(req.CurrentPassword)) != nil {
		fail(c, 400, "当前密码不正确")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil || config.DB.Model(profile).Update("password_hash", string(hash)).Error != nil {
		fail(c, http.StatusInternalServerError, "密码更新失败")
		return
	}
	ok(c, gin.H{"changed": true})
}

func GetAISettings(c *gin.Context) {
	if !localSettingsRequest(c) {
		fail(c, http.StatusForbidden, "智能体配置仅允许在服务器本机管理")
		return
	}
	ok(c, gin.H{"apiKeyMasked": config.MaskedAPIKey(), "configured": config.AIApiKey != "", "baseUrl": config.AIBaseURL, "model": config.AIModel})
}

func UpdateAISettings(c *gin.Context) {
	if !localSettingsRequest(c) {
		fail(c, http.StatusForbidden, "智能体配置仅允许在服务器本机管理")
		return
	}
	var req struct {
		APIKey  string `json:"apiKey"`
		BaseURL string `json:"baseUrl"`
		Model   string `json:"model"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "配置格式错误")
		return
	}
	if err := config.UpdateAISettings(req.APIKey, req.BaseURL, req.Model); err != nil {
		fail(c, 400, err.Error())
		return
	}
	ok(c, gin.H{"apiKeyMasked": config.MaskedAPIKey(), "configured": config.AIApiKey != "", "baseUrl": config.AIBaseURL, "model": config.AIModel})
}

func TestAISettings(c *gin.Context) {
	if !localSettingsRequest(c) {
		fail(c, http.StatusForbidden, "智能体配置仅允许在服务器本机管理")
		return
	}
	if config.AIApiKey == "" {
		fail(c, 400, "请先配置 API Key")
		return
	}
	reply, err := service.TestAIConnection()
	if err != nil {
		fail(c, http.StatusBadGateway, fmt.Sprintf("模型连接失败：%v", err))
		return
	}
	ok(c, gin.H{"connected": true, "reply": truncate(reply, 80)})
}

func localSettingsRequest(c *gin.Context) bool {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		host = c.Request.RemoteAddr
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}
