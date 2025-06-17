package models

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// SSHPublicKey SSH公钥
type SSHPublicKey struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"-" db:"user_id"`
	Title        string    `json:"title" db:"title"`
	Fingerprint  string    `json:"fingerprint" db:"fingerprint"`
	KeyStr       string    `json:"-" db:"key_str"`
	AddedTime    time.Time `json:"added_time" db:"added_time"`
	LastUsedTime time.Time `json:"last_used_time" db:"last_used_time"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// GenerateFingerprint 生成SSH密钥指纹
func (k *SSHPublicKey) GenerateFingerprint() error {
	// 简化的指纹生成逻辑，实际应该解析SSH密钥格式
	parts := strings.Fields(k.KeyStr)
	if len(parts) < 2 {
		return fmt.Errorf("invalid SSH key format")
	}
	
	keyData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode SSH key: %v", err)
	}
	
	hash := sha256.Sum256(keyData)
	k.Fingerprint = fmt.Sprintf("SHA256:%s", base64.StdEncoding.EncodeToString(hash[:]))
	return nil
}

// UpdateLastUsedTime 更新最后使用时间
func (k *SSHPublicKey) UpdateLastUsedTime() {
	k.LastUsedTime = time.Now()
} 