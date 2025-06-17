package models

import (
	"time"
)

// Storage 存储配置模型
type Storage struct {
	ID              int       `json:"id" db:"id"`
	UserID          int       `json:"user_id" db:"user_id"` // 所属用户ID
	MountPath       string    `json:"mount_path" db:"mount_path" binding:"required"`
	OrderIndex      int       `json:"order" db:"order_index"`
	Driver          string    `json:"driver" db:"driver"`
	CacheExpiration int       `json:"cache_expiration" db:"cache_expiration"`
	Status          string    `json:"status" db:"status"`
	Addition        string    `json:"addition" db:"addition"`
	Remark          string    `json:"remark" db:"remark"`
	Modified        time.Time `json:"modified" db:"modified"`
	Disabled        bool      `json:"disabled" db:"disabled"`
	DisableIndex    bool      `json:"disable_index" db:"disable_index"`
	EnableSign      bool      `json:"enable_sign" db:"enable_sign"`
	
	// Access control fields
	IsPublic     bool `json:"is_public" db:"is_public"`         // 是否公开访问
	AllowGuest   bool `json:"allow_guest" db:"allow_guest"`     // 是否允许访客访问
	RequireAuth  bool `json:"require_auth" db:"require_auth"`   // 是否需要认证
	
	// Sort fields
	OrderBy        string `json:"order_by" db:"order_by"`
	OrderDirection string `json:"order_direction" db:"order_direction"`
	ExtractFolder  string `json:"extract_folder" db:"extract_folder"`
	
	// Proxy fields
	WebProxy     bool   `json:"web_proxy" db:"web_proxy"`
	WebdavPolicy string `json:"webdav_policy" db:"webdav_policy"`
	ProxyRange   bool   `json:"proxy_range" db:"proxy_range"`
	DownProxyUrl string `json:"down_proxy_url" db:"down_proxy_url"`
	
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GetStorage 返回存储配置
func (s *Storage) GetStorage() *Storage {
	return s
}

// SetStorage 设置存储配置
func (s *Storage) SetStorage(storage Storage) {
	*s = storage
}

// SetStatus 设置状态
func (s *Storage) SetStatus(status string) {
	s.Status = status
}

// Webdav302 检查是否使用302重定向
func (s *Storage) Webdav302() bool {
	return s.WebdavPolicy == "302_redirect"
}

// WebdavProxy 检查是否使用代理URL
func (s *Storage) WebdavProxy() bool {
	return s.WebdavPolicy == "use_proxy_url"
}

// WebdavNative 检查是否使用原生方式
func (s *Storage) WebdavNative() bool {
	return !s.Webdav302() && !s.WebdavProxy()
} 