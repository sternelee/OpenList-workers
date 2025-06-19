package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/sternelee/OpenList-workers/v3/internal/model"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PluginManager 搜索插件管理器
type PluginManager struct {
	db         *gorm.db
	pluginDir  string
	mu         sync.RWMutex
	plugins    map[string]*model.SearchPlugin
	activeJobs map[string]*SearchJob
	client     *http.Client
}

// SearchJob 搜索任务
type SearchJob struct {
	ID        string
	Query     string
	Plugins   []string
	Results   []model.SearchResult
	Status    string // running, completed, error
	Error     string
	StartTime time.Time
	EndTime   time.Time
	Mu        sync.RWMutex
}

// NewPluginManager 创建插件管理器
func NewPluginManager(db *gorm.db, pluginDir string) *PluginManager {
	return &PluginManager{
		db:         db,
		pluginDir:  pluginDir,
		plugins:    make(map[string]*model.SearchPlugin),
		activeJobs: make(map[string]*SearchJob),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Start 启动插件管理器
func (pm *PluginManager) Start() error {
	// 创建插件目录
	if err := os.MkdirAll(pm.pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// 创建数据表
	if err := pm.db.AutoMigrate(&model.SearchPlugin{}, &model.SearchResult{}); err != nil {
		return fmt.Errorf("failed to migrate search tables: %w", err)
	}

	// 加载已安装的插件
	return pm.loadInstalledPlugins()
}

// loadInstalledPlugins 加载已安装的插件
func (pm *PluginManager) loadInstalledPlugins() error {
	var plugins []model.SearchPlugin
	if err := pm.db.Find(&plugins).Error; err != nil {
		return err
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, plugin := range plugins {
		// 检查插件文件是否存在
		if _, err := os.Stat(plugin.FilePath); err != nil {
			log.Warnf("Plugin file not found: %s", plugin.FilePath)
			continue
		}

		pm.plugins[plugin.Name] = &plugin
	}

	log.Infof("Loaded %d search plugins", len(pm.plugins))
	return nil
}

// InstallPlugin 安装搜索插件
func (pm *PluginManager) InstallPlugin(name, url string) (*model.SearchPlugin, error) {
	// 检查插件是否已存在
	if _, exists := pm.plugins[name]; exists {
		return nil, fmt.Errorf("plugin already exists: %s", name)
	}

	// 下载插件文件
	filePath := filepath.Join(pm.pluginDir, name+".py")
	if err := pm.downloadFile(url, filePath); err != nil {
		return nil, fmt.Errorf("failed to download plugin: %w", err)
	}

	// 解析插件信息
	pluginInfo, err := pm.parsePluginInfo(filePath)
	if err != nil {
		os.Remove(filePath) // 清理失败的下载
		return nil, fmt.Errorf("failed to parse plugin info: %w", err)
	}

	plugin := &model.SearchPlugin{
		Name:        name,
		DisplayName: pluginInfo.DisplayName,
		Version:     pluginInfo.Version,
		URL:         url,
		FilePath:    filePath,
		IsEnabled:   true,
		Categories:  pluginInfo.Categories,
	}

	// 保存到数据库
	if err := pm.db.Create(plugin).Error; err != nil {
		os.Remove(filePath) // 清理失败的下载
		return nil, fmt.Errorf("failed to save plugin to database: %w", err)
	}

	pm.mu.Lock()
	pm.plugins[name] = plugin
	pm.mu.Unlock()

	log.Infof("Successfully installed plugin: %s", name)
	return plugin, nil
}

// downloadFile 下载文件
func (pm *PluginManager) downloadFile(url, filePath string) error {
	resp, err := pm.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// PluginInfo 插件信息
type PluginInfo struct {
	DisplayName string   `json:"display_name"`
	Version     string   `json:"version"`
	Categories  []string `json:"categories"`
}

// parsePluginInfo 解析插件信息
func (pm *PluginManager) parsePluginInfo(filePath string) (*PluginInfo, error) {
	// 执行 Python 脚本获取插件信息
	cmd := exec.Command("python3", filePath, "--info")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var info PluginInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetPlugins 获取所有插件
func (pm *PluginManager) GetPlugins() []model.SearchPlugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]model.SearchPlugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, *plugin)
	}

	return plugins
}

// EnablePlugin 启用插件
func (pm *PluginManager) EnablePlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	plugin.IsEnabled = true
	return pm.db.Model(plugin).Update("is_enabled", true).Error
}

// DisablePlugin 禁用插件
func (pm *PluginManager) DisablePlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	plugin.IsEnabled = false
	return pm.db.Model(plugin).Update("is_enabled", false).Error
}

// UninstallPlugin 卸载插件
func (pm *PluginManager) UninstallPlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// 删除文件
	if err := os.Remove(plugin.FilePath); err != nil {
		log.Warnf("Failed to remove plugin file: %v", err)
	}

	// 从数据库删除
	if err := pm.db.Delete(plugin).Error; err != nil {
		return err
	}

	delete(pm.plugins, name)
	log.Infof("Successfully uninstalled plugin: %s", name)
	return nil
}

// Search 执行搜索
func (pm *PluginManager) Search(ctx context.Context, req *model.ResourceSearchReq) (*SearchJob, error) {
	jobID := uuid.New().String()

	job := &SearchJob{
		ID:        jobID,
		Query:     req.Query,
		Plugins:   req.Plugins,
		Results:   make([]model.SearchResult, 0),
		Status:    "running",
		StartTime: time.Now(),
	}

	pm.mu.Lock()
	pm.activeJobs[jobID] = job
	pm.mu.Unlock()

	// 异步执行搜索
	go pm.executeSearch(ctx, job, req)

	return job, nil
}

// executeSearch 执行搜索任务
func (pm *PluginManager) executeSearch(ctx context.Context, job *SearchJob, req *model.ResourceSearchReq) {
	defer func() {
		job.Mu.Lock()
		job.EndTime = time.Now()
		if job.Status == "running" {
			job.Status = "completed"
		}
		job.Mu.Unlock()

		// 5分钟后清理任务
		time.AfterFunc(5*time.Minute, func() {
			pm.mu.Lock()
			delete(pm.activeJobs, job.ID)
			pm.mu.Unlock()
		})
	}()

	// 确定要使用的插件
	var pluginsToUse []string
	if len(req.Plugins) > 0 {
		pluginsToUse = req.Plugins
	} else {
		// 使用所有启用的插件
		pm.mu.RLock()
		for name, plugin := range pm.plugins {
			if plugin.IsEnabled {
				pluginsToUse = append(pluginsToUse, name)
			}
		}
		pm.mu.RUnlock()
	}

	// 并发搜索多个插件
	resultChan := make(chan []model.SearchResult, len(pluginsToUse))
	var wg sync.WaitGroup

	for _, pluginName := range pluginsToUse {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			pm.mu.RLock()
			plugin, exists := pm.plugins[name]
			pm.mu.RUnlock()

			if !exists || !plugin.IsEnabled {
				return
			}

			results, err := pm.searchWithPlugin(ctx, plugin, req)
			if err != nil {
				log.Errorf("Search failed for plugin %s: %v", name, err)
				return
			}

			resultChan <- results
		}(pluginName)
	}

	// 等待所有搜索完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集搜索结果
	var allResults []model.SearchResult
	for results := range resultChan {
		allResults = append(allResults, results...)
	}

	// 过滤和排序结果
	filteredResults := pm.filterResults(allResults, req)

	job.Mu.Lock()
	job.Results = filteredResults
	job.Mu.Unlock()

	// 保存搜索结果到数据库
	pm.saveSearchResults(job.ID, filteredResults)
}

// searchWithPlugin 使用指定插件搜索
func (pm *PluginManager) searchWithPlugin(ctx context.Context, plugin *model.SearchPlugin, req *model.ResourceSearchReq) ([]model.SearchResult, error) {
	// 构建命令参数
	args := []string{plugin.FilePath, "--search", req.Query}

	if req.Category != "" {
		args = append(args, "--category", req.Category)
	}

	// 执行搜索命令
	cmd := exec.CommandContext(ctx, "python3", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 解析搜索结果
	var rawResults []map[string]interface{}
	if err := json.Unmarshal(output, &rawResults); err != nil {
		return nil, err
	}

	var results []model.SearchResult
	for _, raw := range rawResults {
		result := model.SearchResult{
			PluginName: plugin.Name,
			Title:      getString(raw, "title"),
			URL:        getString(raw, "url"),
			TorrentURL: getString(raw, "torrent_url"),
			MagnetLink: getString(raw, "magnet_link"),
			Size:       getString(raw, "size"),
			Seeds:      getInt(raw, "seeds"),
			Leechs:     getInt(raw, "leechs"),
			Category:   getString(raw, "category"),
		}
		results = append(results, result)
	}

	return results, nil
}

// filterResults 过滤搜索结果
func (pm *PluginManager) filterResults(results []model.SearchResult, req *model.ResourceSearchReq) []model.SearchResult {
	var filtered []model.SearchResult

	for _, result := range results {
		// 最小种子数过滤
		if req.MinSeeds > 0 && result.Seeds < req.MinSeeds {
			continue
		}

		// 最大文件大小过滤（这里需要实现大小比较逻辑）
		if req.MaxSize != "" && !pm.checkSizeLimit(result.Size, req.MaxSize) {
			continue
		}

		filtered = append(filtered, result)
	}

	// 按种子数排序
	return pm.sortResultsBySeeds(filtered)
}

// checkSizeLimit 检查文件大小限制
func (pm *PluginManager) checkSizeLimit(sizeStr, maxSizeStr string) bool {
	// 这里需要实现文件大小解析和比较逻辑
	// 简化实现，实际需要解析 GB, MB 等单位
	return true
}

// sortResultsBySeeds 按种子数排序
func (pm *PluginManager) sortResultsBySeeds(results []model.SearchResult) []model.SearchResult {
	// 简单的冒泡排序，按种子数降序
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-1-i; j++ {
			if results[j].Seeds < results[j+1].Seeds {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
	return results
}

// saveSearchResults 保存搜索结果
func (pm *PluginManager) saveSearchResults(searchID string, results []model.SearchResult) {
	for _, result := range results {
		result.SearchID = searchID
		if err := pm.db.Create(&result).Error; err != nil {
			log.Errorf("Failed to save search result: %v", err)
		}
	}
}

// GetSearchJob 获取搜索任务
func (pm *PluginManager) GetSearchJob(jobID string) (*SearchJob, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	job, exists := pm.activeJobs[jobID]
	if !exists {
		return nil, fmt.Errorf("search job not found: %s", jobID)
	}

	return job, nil
}

// GetSearchResults 获取搜索结果
func (pm *PluginManager) GetSearchResults(searchID string, limit, offset int) ([]model.SearchResult, error) {
	var results []model.SearchResult
	query := pm.db.Where("search_id = ?", searchID)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("seeds DESC").Find(&results).Error
	return results, err
}

// 辅助函数
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}
