package models

import "time"

// 系统设置组常量
const (
	SINGLE = iota
	SITE
	STYLE
	PREVIEW
	GLOBAL
	OFFLINE_DOWNLOAD
	INDEX
	SSO
	LDAP
	S3
	FTP
	TRAFFIC
)

// 设置标志常量
const (
	PUBLIC = iota
	PRIVATE
	READONLY
	DEPRECATED
)

// SettingItem 系统设置项
type SettingItem struct {
	Key         string    `json:"key" db:"key" binding:"required"`
	Value       string    `json:"value" db:"value"`
	Help        string    `json:"help" db:"help"`
	Type        string    `json:"type" db:"type"`       // string, number, bool, select
	Options     string    `json:"options" db:"options"` // values for select
	GroupID     int       `json:"group" db:"group_id"`  // use to group setting in frontend
	Flag        int       `json:"flag" db:"flag"`       // 0 = public, 1 = private, 2 = readonly, 3 = deprecated
	IndexOrder  int       `json:"index" db:"index_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// IsDeprecated 检查设置是否已弃用
func (s SettingItem) IsDeprecated() bool {
	return s.Flag == DEPRECATED
} 