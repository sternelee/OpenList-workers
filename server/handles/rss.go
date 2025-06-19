package handles

import (
	"strconv"

	"github.com/sternelee/OpenList-workers/v3/internal/db"
	"github.com/sternelee/OpenList-workers/v3/internal/model"
	"github.com/sternelee/OpenList-workers/v3/internal/rss"
	"github.com/sternelee/OpenList-workers/v3/server/common"
	"github.com/gin-gonic/gin"
)

var RSSService *rss.Service

// RSS 文件夹管理
func ListRSSFolders(c *gin.Context) {
	var folders []model.RSSFolder
	if err := db.db.Preload("Children").Preload("Feeds").Where("parent_id IS NULL").Find(&folders).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, folders)
}

func CreateRSSFolder(c *gin.Context) {
	var req model.AddRSSFolderReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	folder, err := RSSService.AddFolder(req.Name, req.ParentPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, folder)
}

func DeleteRSSFolder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	// 检查文件夹是否为空
	var count int64
	db.db.Model(&model.RSSFeed{}).Where("folder_id = ?", uint(id)).Count(&count)
	if count > 0 {
		common.ErrorStrResp(c, "folder is not empty", 400)
		return
	}

	var subFolderCount int64
	db.db.Model(&model.RSSFolder{}).Where("parent_id = ?", uint(id)).Count(&subFolderCount)
	if subFolderCount > 0 {
		common.ErrorStrResp(c, "folder contains subfolders", 400)
		return
	}

	if err := db.db.Delete(&model.RSSFolder{}, uint(id)).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c)
}

// RSS 订阅管理
func ListRSSFeeds(c *gin.Context) {
	feeds, err := RSSService.GetFeeds()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, feeds)
}

func CreateRSSFeed(c *gin.Context) {
	var req model.AddRSSFeedReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	if req.RefreshInterval <= 0 {
		req.RefreshInterval = 300 // 默认5分钟
	}

	feed, err := RSSService.AddFeed(req.Name, req.URL, req.FolderPath, req.RefreshInterval)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, feed)
}

func GetRSSFeed(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	feed, err := RSSService.GetFeedByID(uint(id))
	if err != nil {
		common.ErrorResp(c, err, 404)
		return
	}

	common.SuccessResp(c, feed)
}

func UpdateRSSFeed(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	var req model.UpdateRSSFeedReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	if err := RSSService.UpdateFeed(uint(id), &req); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c)
}

func DeleteRSSFeed(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	if err := RSSService.DeleteFeed(uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c)
}

func RefreshRSSFeed(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	if err := RSSService.RefreshFeed(uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, gin.H{"message": "feed refresh started"})
}

// RSS 文章管理
func ListRSSArticles(c *gin.Context) {
	feedIDStr := c.Query("feed_id")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "50")
	unreadOnly := c.Query("unread_only") == "true"

	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)

	offset := (pageInt - 1) * perPageInt

	query := db.db.Model(&model.RSSArticle{})

	if feedIDStr != "" {
		feedID, err := strconv.ParseUint(feedIDStr, 10, 32)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		query = query.Where("feed_id = ?", uint(feedID))
	}

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	var total int64
	query.Count(&total)

	var articles []model.RSSArticle
	if err := query.Preload("Feed").
		Order("pub_date DESC").
		Limit(perPageInt).
		Offset(offset).
		Find(&articles).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, gin.H{
		"articles": articles,
		"total":    total,
		"page":     pageInt,
		"per_page": perPageInt,
	})
}

func MarkRSSArticleRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	if err := db.db.Model(&model.RSSArticle{}).Where("id = ?", uint(id)).Update("is_read", true).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c)
}

func MarkAllRSSArticlesRead(c *gin.Context) {
	feedIDStr := c.Query("feed_id")

	query := db.db.Model(&model.RSSArticle{})
	if feedIDStr != "" {
		feedID, err := strconv.ParseUint(feedIDStr, 10, 32)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		query = query.Where("feed_id = ?", uint(feedID))
	}

	if err := query.Update("is_read", true).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c)
}

// RSS 自动下载规则管理
func ListRSSAutoDownloadRules(c *gin.Context) {
	var rules []model.RSSAutoDownloadRule
	if err := db.db.Find(&rules).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, rules)
}

func CreateRSSAutoDownloadRule(c *gin.Context) {
	var req model.AddAutoDownloadRuleReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	rule := &model.RSSAutoDownloadRule{
		Name:               req.Name,
		IsEnabled:          true,
		MustContain:        req.MustContain,
		MustNotContain:     req.MustNotContain,
		UseRegex:           req.UseRegex,
		EpisodeFilter:      req.EpisodeFilter,
		SmartFilter:        req.SmartFilter,
		AffectedFeeds:      req.AffectedFeeds,
		DestinationPath:    req.DestinationPath,
		CategoryAssignment: req.CategoryAssignment,
		AddPaused:          req.AddPaused,
		DownloadTool:       req.DownloadTool,
		DeletePolicy:       req.DeletePolicy,
		TorrentTempPath:    req.TorrentTempPath,
	}

	if err := db.db.Create(rule).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, rule)
}

func UpdateRSSAutoDownloadRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	var req model.AddAutoDownloadRuleReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	updates := map[string]interface{}{
		"name":                req.Name,
		"must_contain":        req.MustContain,
		"must_not_contain":    req.MustNotContain,
		"use_regex":           req.UseRegex,
		"episode_filter":      req.EpisodeFilter,
		"smart_filter":        req.SmartFilter,
		"affected_feeds":      req.AffectedFeeds,
		"destination_path":    req.DestinationPath,
		"category_assignment": req.CategoryAssignment,
		"add_paused":          req.AddPaused,
		"download_tool":       req.DownloadTool,
		"delete_policy":       req.DeletePolicy,
		"torrent_temp_path":   req.TorrentTempPath,
	}

	if err := db.db.Model(&model.RSSAutoDownloadRule{}).Where("id = ?", uint(id)).Updates(updates).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c)
}

func DeleteRSSAutoDownloadRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	if err := db.db.Delete(&model.RSSAutoDownloadRule{}, uint(id)).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c)
}

func ToggleRSSAutoDownloadRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	var rule model.RSSAutoDownloadRule
	if err := db.db.First(&rule, uint(id)).Error; err != nil {
		common.ErrorResp(c, err, 404)
		return
	}

	if err := db.db.Model(&rule).Update("is_enabled", !rule.IsEnabled).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, gin.H{"is_enabled": !rule.IsEnabled})
}

// GetRSSDownloadTools 获取RSS可用的下载工具
func GetRSSDownloadTools(c *gin.Context) {
	tools, err := RSSService.GetAvailableDownloadTools()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, tools)
}
