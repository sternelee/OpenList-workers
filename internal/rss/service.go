package rss

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sternelee/OpenList-workers/v3/internal/model"
	"github.com/sternelee/OpenList-workers/v3/internal/offline_download/tool"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service RSS 服务管理器
type Service struct {
	db          *gorm.DB
	parser      *FeedParser
	mu          sync.RWMutex
	refreshTicker *time.Ticker
	stopCh      chan struct{}
	maxArticles int
	isRunning   bool
}

// NewService 创建 RSS 服务实例
func NewService(database *gorm.DB) *Service {
	return &Service{
		db:          database,
		parser:      NewFeedParser(),
		maxArticles: 1000, // 每个 feed 最多保存1000篇文章
		stopCh:      make(chan struct{}),
	}
}

// Start 启动 RSS 服务
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return nil
	}

	// 创建数据表
	if err := s.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// 启动定时刷新
	s.refreshTicker = time.NewTicker(5 * time.Minute)
	go s.refreshLoop()

	s.isRunning = true
	log.Info("RSS service started")
	return nil
}

// Stop 停止 RSS 服务
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	close(s.stopCh)
	if s.refreshTicker != nil {
		s.refreshTicker.Stop()
	}

	s.isRunning = false
	log.Info("RSS service stopped")
}

// createTables 创建数据表
func (s *Service) createTables() error {
	return s.db.AutoMigrate(
		&model.RSSFolder{},
		&model.RSSFeed{},
		&model.RSSArticle{},
		&model.RSSAutoDownloadRule{},
	)
}

// AddFolder 添加 RSS 文件夹
func (s *Service) AddFolder(name, parentPath string) (*model.RSSFolder, error) {
	var parentFolder *model.RSSFolder
	var parentID *uint

	if parentPath != "" && parentPath != "/" {
		var err error
		parentFolder, err = s.GetFolderByPath(parentPath)
		if err != nil {
			return nil, fmt.Errorf("parent folder not found: %w", err)
		}
		parentID = &parentFolder.ID
	}

	path := parentPath
	if path == "" || path == "/" {
		path = "/" + name
	} else {
		path = strings.TrimSuffix(path, "/") + "/" + name
	}

	folder := &model.RSSFolder{
		Name:     name,
		Path:     path,
		ParentID: parentID,
	}

	if err := s.db.Create(folder).Error; err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return folder, nil
}

// GetFolderByPath 根据路径获取文件夹
func (s *Service) GetFolderByPath(path string) (*model.RSSFolder, error) {
	var folder model.RSSFolder
	if err := s.db.Where("path = ?", path).First(&folder).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

// AddFeed 添加 RSS 订阅
func (s *Service) AddFeed(name, url, folderPath string, refreshInterval int) (*model.RSSFeed, error) {
	// 检查是否已存在
	var existingFeed model.RSSFeed
	if err := s.db.Where("url = ?", url).First(&existingFeed).Error; err == nil {
		return nil, fmt.Errorf("feed already exists")
	}

	var folderID *uint
	if folderPath != "" && folderPath != "/" {
		folder, err := s.GetFolderByPath(folderPath)
		if err != nil {
			return nil, fmt.Errorf("folder not found: %w", err)
		}
		folderID = &folder.ID
	}

	// 生成唯一ID
	uid := uuid.New().String()

	feed := &model.RSSFeed{
		UID:             uid,
		Name:            name,
		URL:             url,
		FolderID:        folderID,
		RefreshInterval: refreshInterval,
		IsEnabled:       true,
	}

	// 首次获取 feed 内容
	if err := s.refreshFeed(feed); err != nil {
		log.Warnf("Failed to refresh feed on creation: %v", err)
		feed.HasError = true
		feed.ErrorMessage = err.Error()
	}

	if err := s.db.Create(feed).Error; err != nil {
		return nil, fmt.Errorf("failed to create feed: %w", err)
	}

	return feed, nil
}

// GetFeeds 获取所有订阅
func (s *Service) GetFeeds() ([]model.RSSFeed, error) {
	var feeds []model.RSSFeed
	if err := s.db.Preload("Folder").Find(&feeds).Error; err != nil {
		return nil, err
	}
	return feeds, nil
}

// GetFeedByID 根据ID获取订阅
func (s *Service) GetFeedByID(id uint) (*model.RSSFeed, error) {
	var feed model.RSSFeed
	if err := s.db.Preload("Articles").First(&feed, id).Error; err != nil {
		return nil, err
	}
	return &feed, nil
}

// UpdateFeed 更新订阅
func (s *Service) UpdateFeed(id uint, req *model.UpdateRSSFeedReq) error {
	feed, err := s.GetFeedByID(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.RefreshInterval > 0 {
		updates["refresh_interval"] = req.RefreshInterval
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	// 处理文件夹变更
	if req.FolderPath != "" {
		if req.FolderPath == "/" {
			updates["folder_id"] = nil
		} else {
			folder, err := s.GetFolderByPath(req.FolderPath)
			if err != nil {
				return fmt.Errorf("folder not found: %w", err)
			}
			updates["folder_id"] = folder.ID
		}
	}

	return s.db.Model(feed).Updates(updates).Error
}

// DeleteFeed 删除订阅
func (s *Service) DeleteFeed(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除相关文章
		if err := tx.Where("feed_id = ?", id).Delete(&model.RSSArticle{}).Error; err != nil {
			return err
		}
		// 删除订阅
		return tx.Delete(&model.RSSFeed{}, id).Error
	})
}

// RefreshFeed 刷新单个订阅
func (s *Service) RefreshFeed(id uint) error {
	feed, err := s.GetFeedByID(id)
	if err != nil {
		return err
	}

	return s.refreshFeed(feed)
}

// refreshFeed 刷新订阅内容
func (s *Service) refreshFeed(feed *model.RSSFeed) error {
	log.Infof("Refreshing feed: %s (%s)", feed.Name, feed.URL)

	feedInfo, articles, err := s.parser.ParseFeed(feed.URL)
	if err != nil {
		feed.HasError = true
		feed.ErrorMessage = err.Error()
		s.db.Model(feed).Updates(map[string]interface{}{
			"has_error":     true,
			"error_message": err.Error(),
			"last_refresh":  time.Now(),
		})
		return err
	}

	// 更新 feed 信息
	now := time.Now()
	updates := map[string]interface{}{
		"has_error":    false,
		"error_message": "",
		"last_refresh": now,
	}

	if feed.Name == "" || feed.Name == feedInfo.Name {
		updates["name"] = feedInfo.Name
	}

	s.db.Model(feed).Updates(updates)

	// 处理文章
	return s.processArticles(feed, articles)
}

// processArticles 处理文章
func (s *Service) processArticles(feed *model.RSSFeed, articles []model.RSSArticle) error {
	for _, article := range articles {
		article.FeedID = feed.ID

		// 检查文章是否已存在
		var existing model.RSSArticle
		if err := s.db.Where("guid = ? AND feed_id = ?", article.GUID, feed.ID).First(&existing).Error; err == nil {
			continue // 文章已存在，跳过
		}

		// 保存新文章
		if err := s.db.Create(&article).Error; err != nil {
			log.Errorf("Failed to save article %s: %v", article.Title, err)
			continue
		}

		// 检查自动下载规则
		s.checkAutoDownloadRules(&article)
	}

	// 清理旧文章
	s.cleanupOldArticles(feed.ID)

	return nil
}

// cleanupOldArticles 清理旧文章
func (s *Service) cleanupOldArticles(feedID uint) {
	var count int64
	s.db.Model(&model.RSSArticle{}).Where("feed_id = ?", feedID).Count(&count)

	if count > int64(s.maxArticles) {
		// 保留最新的文章
		var articles []model.RSSArticle
		s.db.Where("feed_id = ?", feedID).
			Order("pub_date DESC").
			Offset(s.maxArticles).
			Find(&articles)

		var ids []uint
		for _, article := range articles {
			ids = append(ids, article.ID)
		}

		if len(ids) > 0 {
			s.db.Where("id IN ?", ids).Delete(&model.RSSArticle{})
		}
	}
}

// refreshLoop 定时刷新循环
func (s *Service) refreshLoop() {
	for {
		select {
		case <-s.refreshTicker.C:
			s.refreshAllFeeds()
		case <-s.stopCh:
			return
		}
	}
}

// refreshAllFeeds 刷新所有启用的订阅
func (s *Service) refreshAllFeeds() {
	var feeds []model.RSSFeed
	if err := s.db.Where("is_enabled = ?", true).Find(&feeds).Error; err != nil {
		log.Errorf("Failed to get enabled feeds: %v", err)
		return
	}

	for _, feed := range feeds {
		// 检查是否需要刷新
		if feed.LastRefresh != nil {
			nextRefresh := feed.LastRefresh.Add(time.Duration(feed.RefreshInterval) * time.Second)
			if time.Now().Before(nextRefresh) {
				continue
			}
		}

		go func(f model.RSSFeed) {
			if err := s.refreshFeed(&f); err != nil {
				log.Errorf("Failed to refresh feed %s: %v", f.Name, err)
			}
		}(feed)
	}
}

// checkAutoDownloadRules 检查自动下载规则
func (s *Service) checkAutoDownloadRules(article *model.RSSArticle) {
	var rules []model.RSSAutoDownloadRule
	if err := s.db.Where("is_enabled = ?", true).Find(&rules).Error; err != nil {
		log.Errorf("Failed to get auto download rules: %v", err)
		return
	}

	var feed model.RSSFeed
	if err := s.db.First(&feed, article.FeedID).Error; err != nil {
		return
	}

	for _, rule := range rules {
		if s.matchesRule(&rule, article, &feed) {
			s.executeAutoDownload(&rule, article)
		}
	}
}

// matchesRule 检查文章是否匹配自动下载规则
func (s *Service) matchesRule(rule *model.RSSAutoDownloadRule, article *model.RSSArticle, feed *model.RSSFeed) bool {
	// 检查是否影响此 feed
	if len(rule.AffectedFeeds) > 0 {
		found := false
		for _, feedUID := range rule.AffectedFeeds {
			if feedUID == feed.UID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查必须包含的关键词
	if rule.MustContain != "" {
		if rule.UseRegex {
			if matched, _ := regexp.MatchString(rule.MustContain, article.Title); !matched {
				return false
			}
		} else {
			if !strings.Contains(strings.ToLower(article.Title), strings.ToLower(rule.MustContain)) {
				return false
			}
		}
	}

	// 检查必须不包含的关键词
	if rule.MustNotContain != "" {
		if rule.UseRegex {
			if matched, _ := regexp.MatchString(rule.MustNotContain, article.Title); matched {
				return false
			}
		} else {
			if strings.Contains(strings.ToLower(article.Title), strings.ToLower(rule.MustNotContain)) {
				return false
			}
		}
	}

	return true
}

// executeAutoDownload 执行自动下载
func (s *Service) executeAutoDownload(rule *model.RSSAutoDownloadRule, article *model.RSSArticle) {
	log.Infof("Auto downloading: %s using tool: %s", article.Title, rule.DownloadTool)

	var downloadURL string
	if article.MagnetLink != "" {
		downloadURL = article.MagnetLink
	} else if article.TorrentURL != "" {
		downloadURL = article.TorrentURL
	} else {
		log.Warnf("No download URL found for article: %s", article.Title)
		return
	}

	// 确定下载工具
	downloadTool := rule.DownloadTool
	if downloadTool == "" {
		downloadTool = "aria2" // 默认使用 aria2
	}

	// 确定删除策略
	deletePolicy := tool.DeletePolicy(rule.DeletePolicy)
	if rule.DeletePolicy == "" {
		deletePolicy = tool.DeleteOnUploadSucceed // 默认策略
	}

	// 确定下载目标路径
	dstPath := rule.DestinationPath

	// 对于云盘下载，如果指定了临时路径，使用临时路径
	if s.isCloudDownloadTool(downloadTool) && rule.TorrentTempPath != "" {
		dstPath = rule.TorrentTempPath
	}

	// 使用离线下载工具
	ctx := context.Background()
	task, err := tool.AddURL(ctx, &tool.AddURLArgs{
		URL:          downloadURL,
		DstDirPath:   dstPath,
		Tool:         downloadTool,
		DeletePolicy: deletePolicy,
	})

	if err != nil {
		log.Errorf("Failed to add auto download task: %v", err)
		return
	}

	log.Infof("Successfully created download task: %s", task.GetID())

	// 更新规则的最后匹配时间
	now := time.Now()
	rule.LastMatch = &now
	s.db.Model(rule).Update("last_match", now)
}

// isCloudDownloadTool 判断是否为云盘下载工具
func (s *Service) isCloudDownloadTool(toolName string) bool {
	cloudTools := []string{"115 Cloud", "PikPak", "Thunder"}
	for _, cloudTool := range cloudTools {
		if toolName == cloudTool {
			return true
		}
	}
	return false
}

// GetAvailableDownloadTools 获取可用的下载工具
func (s *Service) GetAvailableDownloadTools() (*model.DownloadToolStatus, error) {
	// 获取所有注册的下载工具
	toolNames := tool.Tools.Names()

	var tools []model.OfflineDownloadTool
	var recommendedTool string

	for _, name := range toolNames {
		// 获取工具实例检查可用性
		_, err := tool.Tools.Get(name)
		isAvailable := err == nil

		// 确定工具类型
		toolType := "local"
		if s.isCloudDownloadTool(name) {
			toolType = "cloud"
		}

		// 获取显示名称和描述
		displayName, description := s.getToolDisplayInfo(name)

		downloadTool := model.OfflineDownloadTool{
			Name:         name,
			DisplayName:  displayName,
			Type:         toolType,
			IsConfigured: isAvailable,
			IsAvailable:  isAvailable,
			Categories:   []string{"all"}, // 大部分工具都支持所有类型
			Description:  description,
		}

		tools = append(tools, downloadTool)

		// 设置推荐工具（优先云盘工具）
		if isAvailable && recommendedTool == "" {
			if toolType == "cloud" {
				recommendedTool = name
			} else if recommendedTool == "" {
				recommendedTool = name
			}
		}
	}

	return &model.DownloadToolStatus{
		Tools:           tools,
		RecommendedTool: recommendedTool,
	}, nil
}

// getToolDisplayInfo 获取工具的显示信息
func (s *Service) getToolDisplayInfo(toolName string) (displayName, description string) {
	switch toolName {
	case "aria2":
		return "Aria2", "高性能多协议下载工具，支持 HTTP、FTP、BitTorrent 等"
	case "qBittorrent":
		return "qBittorrent", "功能强大的 BitTorrent 客户端，支持做种管理"
	case "Transmission":
		return "Transmission", "轻量级跨平台 BitTorrent 客户端"
	case "115 Cloud":
		return "115云盘", "使用115网盘的云端离线下载功能"
	case "PikPak":
		return "PikPak网盘", "使用PikPak网盘的云端离线下载功能"
	case "Thunder":
		return "迅雷网盘", "使用迅雷网盘的云端离线下载功能"
	default:
		return toolName, "下载工具"
	}
}