package bootstrap

import (
	"path/filepath"

	"github.com/sternelee/OpenList-workers/v3/cmd/flags"
	"github.com/sternelee/OpenList-workers/v3/internal/db"
	"github.com/sternelee/OpenList-workers/v3/internal/rss"
	"github.com/sternelee/OpenList-workers/v3/internal/search"
	"github.com/sternelee/OpenList-workers/v3/server/handles"
	log "github.com/sirupsen/logrus"
)

// InitRSSAndSearch 初始化 RSS 和搜索服务
func InitRSSAndSearch() error {
	// 初始化 RSS 服务
	rssService := rss.NewService(db.GetDb())
	if err := rssService.Start(); err != nil {
		return err
	}
	handles.RSSService = rssService
	log.Info("RSS service initialized successfully")

	// 初始化搜索插件管理器
	pluginDir := filepath.Join(flags.DataDir, "search_plugins")
	searchManager := search.NewPluginManager(db.GetDb(), pluginDir)
	if err := searchManager.Start(); err != nil {
		return err
	}
	handles.SearchManager = searchManager
	log.Info("Search plugin manager initialized successfully")

	return nil
}
