package model

import (
	"time"
)

// RSSFolder RSS 文件夹
type RSSFolder struct {
	ID       uint        `json:"id" gorm:"primaryKey"`
	Name     string      `json:"name" gorm:"not null"`
	Path     string      `json:"path" gorm:"uniqueIndex;not null"`
	ParentID *uint       `json:"parent_id"`
	Parent   *RSSFolder  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []RSSFolder `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Feeds    []RSSFeed   `json:"feeds,omitempty" gorm:"foreignKey:FolderID"`
	Created  time.Time   `json:"created" gorm:"autoCreateTime"`
	Updated  time.Time   `json:"updated" gorm:"autoUpdateTime"`
}

// RSSFeed RSS 订阅源
type RSSFeed struct {
	ID              uint         `json:"id" gorm:"primaryKey"`
	UID             string       `json:"uid" gorm:"uniqueIndex;not null"` // 唯一标识符
	Name            string       `json:"name" gorm:"not null"`            // 别名
	URL             string       `json:"url" gorm:"not null"`             // RSS URL
	FolderID        *uint        `json:"folder_id"`
	Folder          *RSSFolder   `json:"folder,omitempty" gorm:"foreignKey:FolderID"`
	RefreshInterval int          `json:"refresh_interval" gorm:"default:300"` // 刷新间隔(秒)
	LastRefresh     *time.Time   `json:"last_refresh"`                        // 最后刷新时间
	IsEnabled       bool         `json:"is_enabled" gorm:"default:true"`      // 是否启用
	HasError        bool         `json:"has_error" gorm:"default:false"`      // 是否有错误
	ErrorMessage    string       `json:"error_message"`                       // 错误信息
	IconPath        string       `json:"icon_path"`                           // 图标路径
	Articles        []RSSArticle `json:"articles,omitempty" gorm:"foreignKey:FeedID"`
	Created         time.Time    `json:"created" gorm:"autoCreateTime"`
	Updated         time.Time    `json:"updated" gorm:"autoUpdateTime"`
}

// RSSArticle RSS 文章
type RSSArticle struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FeedID      uint      `json:"feed_id" gorm:"index;not null"`
	Feed        *RSSFeed  `json:"feed,omitempty" gorm:"foreignKey:FeedID"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	URL         string    `json:"url" gorm:"not null"`
	GUID        string    `json:"guid" gorm:"uniqueIndex;not null"` // 全局唯一标识
	Author      string    `json:"author"`
	Category    string    `json:"category"`
	PubDate     time.Time `json:"pub_date"`
	IsRead      bool      `json:"is_read" gorm:"default:false"`
	TorrentURL  string    `json:"torrent_url"`  // 种子链接
	MagnetLink  string    `json:"magnet_link"`  // 磁力链接
	FileSize    int64     `json:"file_size"`    // 文件大小
	Seeders     int       `json:"seeders"`      // 做种数
	Leechers    int       `json:"leechers"`     // 下载数
	Created     time.Time `json:"created" gorm:"autoCreateTime"`
	Updated     time.Time `json:"updated" gorm:"autoUpdateTime"`
}

// RSSAutoDownloadRule RSS 自动下载规则
type RSSAutoDownloadRule struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	Name                 string    `json:"name" gorm:"not null"`
	IsEnabled            bool      `json:"is_enabled" gorm:"default:true"`
	MustContain          string    `json:"must_contain"`          // 必须包含的关键词
	MustNotContain       string    `json:"must_not_contain"`      // 必须不包含的关键词
	UseRegex             bool      `json:"use_regex"`             // 是否使用正则表达式
	EpisodeFilter        string    `json:"episode_filter"`        // 剧集过滤器
	SmartFilter          bool      `json:"smart_filter"`          // 智能过滤
	PreviouslyMatchedEps []string  `json:"previously_matched_eps" gorm:"type:json"` // 之前匹配的剧集
	AffectedFeeds        []string  `json:"affected_feeds" gorm:"type:json"`         // 影响的Feed UID列表
	DestinationPath      string    `json:"destination_path"`                        // 目标路径
	CategoryAssignment   string    `json:"category_assignment"`                     // 分类分配
	AddPaused            bool      `json:"add_paused"`                              // 添加时暂停
	// 新增：下载工具配置
	DownloadTool         string    `json:"download_tool" gorm:"default:aria2"`     // 下载工具: aria2, qBittorrent, Transmission, 115 Cloud, PikPak, Thunder
	DeletePolicy         string    `json:"delete_policy" gorm:"default:delete_on_upload_succeed"` // 删除策略
	TorrentTempPath      string    `json:"torrent_temp_path"`                       // 种子临时路径(用于云盘下载)
	LastMatch            *time.Time `json:"last_match"`                             // 最后匹配时间
	Created              time.Time `json:"created" gorm:"autoCreateTime"`
	Updated              time.Time `json:"updated" gorm:"autoUpdateTime"`
}

// SearchPlugin 搜索插件
type SearchPlugin struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	DisplayName string    `json:"display_name" gorm:"not null"`
	Version     string    `json:"version"`
	URL         string    `json:"url"`         // 插件下载URL
	FilePath    string    `json:"file_path"`   // 插件文件路径
	IsEnabled   bool      `json:"is_enabled" gorm:"default:true"`
	Categories  []string  `json:"categories" gorm:"type:json"` // 支持的分类
	Created     time.Time `json:"created" gorm:"autoCreateTime"`
	Updated     time.Time `json:"updated" gorm:"autoUpdateTime"`
}

// SearchResult 搜索结果
type SearchResult struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SearchID   string    `json:"search_id" gorm:"index;not null"` // 搜索会话ID
	PluginName string    `json:"plugin_name" gorm:"not null"`
	Title      string    `json:"title" gorm:"not null"`
	URL        string    `json:"url" gorm:"not null"`
	TorrentURL string    `json:"torrent_url"`
	MagnetLink string    `json:"magnet_link"`
	Size       string    `json:"size"`
	Seeds      int       `json:"seeds"`
	Leechs     int       `json:"leechs"`
	Category   string    `json:"category"`
	Created    time.Time `json:"created" gorm:"autoCreateTime"`
}

// RSS 相关请求结构体
type AddRSSFeedReq struct {
	Name            string `json:"name" binding:"required"`
	URL             string `json:"url" binding:"required"`
	FolderPath      string `json:"folder_path"`
	RefreshInterval int    `json:"refresh_interval"`
}

type UpdateRSSFeedReq struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	FolderPath      string `json:"folder_path"`
	RefreshInterval int    `json:"refresh_interval"`
	IsEnabled       *bool  `json:"is_enabled"`
}

type AddRSSFolderReq struct {
	Name       string `json:"name" binding:"required"`
	ParentPath string `json:"parent_path"`
}

type AddAutoDownloadRuleReq struct {
	Name                string   `json:"name" binding:"required"`
	MustContain         string   `json:"must_contain"`
	MustNotContain      string   `json:"must_not_contain"`
	UseRegex            bool     `json:"use_regex"`
	EpisodeFilter       string   `json:"episode_filter"`
	SmartFilter         bool     `json:"smart_filter"`
	AffectedFeeds       []string `json:"affected_feeds"`
	DestinationPath     string   `json:"destination_path"`
	CategoryAssignment  string   `json:"category_assignment"`
	AddPaused           bool     `json:"add_paused"`
	// 新增：下载工具配置
	DownloadTool        string   `json:"download_tool"`     // 下载工具选择
	DeletePolicy        string   `json:"delete_policy"`     // 删除策略
	TorrentTempPath     string   `json:"torrent_temp_path"` // 种子临时路径
}

// 搜索相关请求结构体
type ResourceSearchReq struct {
	Query     string   `json:"query" binding:"required"`
	Plugins   []string `json:"plugins"`   // 指定搜索插件
	Category  string   `json:"category"`  // 搜索分类
	MinSeeds  int      `json:"min_seeds"` // 最小种子数
	MaxSize   string   `json:"max_size"`  // 最大文件大小
}

type InstallSearchPluginReq struct {
	Name string `json:"name" binding:"required"`
	URL  string `json:"url" binding:"required"`
}

// OfflineDownloadTool 离线下载工具信息
type OfflineDownloadTool struct {
	Name         string   `json:"name"`         // 工具名称
	DisplayName  string   `json:"display_name"` // 显示名称
	Type         string   `json:"type"`         // 类型: local, cloud
	IsConfigured bool     `json:"is_configured"` // 是否已配置
	IsAvailable  bool     `json:"is_available"`  // 是否可用
	Categories   []string `json:"categories"`    // 支持的分类
	Description  string   `json:"description"`   // 描述
}

// DownloadToolStatus 下载工具状态响应
type DownloadToolStatus struct {
	Tools           []OfflineDownloadTool `json:"tools"`
	RecommendedTool string                `json:"recommended_tool"`
}