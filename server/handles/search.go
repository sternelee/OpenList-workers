package handles

import (
	"context"
	"path"
	"strconv"
	"strings"

	"github.com/sternelee/OpenList-workers/v3/internal/db"
	"github.com/sternelee/OpenList-workers/v3/internal/errs"
	"github.com/sternelee/OpenList-workers/v3/internal/model"
	"github.com/sternelee/OpenList-workers/v3/internal/offline_download/tool"
	"github.com/sternelee/OpenList-workers/v3/internal/op"
	"github.com/sternelee/OpenList-workers/v3/internal/search"
	"github.com/sternelee/OpenList-workers/v3/pkg/utils"
	"github.com/sternelee/OpenList-workers/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type SearchReq struct {
	model.SearchReq
	Password string `json:"password"`
}

type SearchResp struct {
	model.SearchNode
	Type int `json:"type"`
}

var SearchManager *search.PluginManager

// 搜索插件管理
func ListSearchPlugins(c *gin.Context) {
	plugins := SearchManager.GetPlugins()
	common.SuccessResp(c, plugins)
}

func InstallSearchPlugin(c *gin.Context) {
	var req model.InstallSearchPluginReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	plugin, err := SearchManager.InstallPlugin(req.Name, req.URL)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, plugin)
}

func EnableSearchPlugin(c *gin.Context) {
	name := c.Param("name")
	if err := SearchManager.EnablePlugin(name); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, gin.H{"message": "plugin enabled"})
}

func DisableSearchPlugin(c *gin.Context) {
	name := c.Param("name")
	if err := SearchManager.DisablePlugin(name); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, gin.H{"message": "plugin disabled"})
}

func UninstallSearchPlugin(c *gin.Context) {
	name := c.Param("name")
	if err := SearchManager.UninstallPlugin(name); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, gin.H{"message": "plugin uninstalled"})
}

// 资源搜索
func SearchResources(c *gin.Context) {
	var req model.ResourceSearchReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	// 启动搜索任务
	job, err := SearchManager.Search(context.Background(), &req)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, gin.H{
		"job_id":     job.ID,
		"status":     job.Status,
		"start_time": job.StartTime,
	})
}

func GetSearchJob(c *gin.Context) {
	jobID := c.Param("job_id")

	job, err := SearchManager.GetSearchJob(jobID)
	if err != nil {
		common.ErrorResp(c, err, 404)
		return
	}

	job.Mu.RLock()
	response := gin.H{
		"job_id":       job.ID,
		"query":        job.Query,
		"status":       job.Status,
		"start_time":   job.StartTime,
		"end_time":     job.EndTime,
		"error":        job.Error,
		"result_count": len(job.Results),
	}
	job.Mu.RUnlock()

	common.SuccessResp(c, response)
}

func GetSearchResults(c *gin.Context) {
	searchID := c.Param("search_id")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "50")

	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)

	offset := (pageInt - 1) * perPageInt

	results, err := SearchManager.GetSearchResults(searchID, perPageInt, offset)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	// 获取总数
	var total int64
	db.db.Model(&model.SearchResult{}).Where("search_id = ?", searchID).Count(&total)

	common.SuccessResp(c, gin.H{
		"results":  results,
		"total":    total,
		"page":     pageInt,
		"per_page": perPageInt,
	})
}

// 直接下载搜索结果
func DownloadSearchResult(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if !user.CanAddOfflineDownloadTasks() {
		common.ErrorStrResp(c, "permission denied", 403)
		return
	}

	var req struct {
		ResultID        uint   `json:"result_id" binding:"required"`
		DestinationPath string `json:"destination_path" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	// 获取搜索结果
	var result model.SearchResult
	if err := db.db.First(&result, req.ResultID).Error; err != nil {
		common.ErrorResp(c, err, 404)
		return
	}

	// 确定下载URL
	var downloadURL string
	if result.MagnetLink != "" {
		downloadURL = result.MagnetLink
	} else if result.TorrentURL != "" {
		downloadURL = result.TorrentURL
	} else {
		common.ErrorStrResp(c, "no download URL available", 400)
		return
	}

	// 验证目标路径
	reqPath, err := user.JoinPath(req.DestinationPath)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	// 添加到离线下载队列
	ctx := context.Background()
	task, err := tool.AddURL(ctx, &tool.AddURLArgs{
		URL:          downloadURL,
		DstDirPath:   reqPath,
		Tool:         "aria2", // 默认使用 aria2
		DeletePolicy: tool.DeleteOnUploadSucceed,
	})

	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, gin.H{
		"task_id": task.GetID(),
		"message": "download task created",
	})
}

// 批量下载搜索结果
func BatchDownloadSearchResults(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if !user.CanAddOfflineDownloadTasks() {
		common.ErrorStrResp(c, "permission denied", 403)
		return
	}

	var req struct {
		ResultIDs       []uint `json:"result_ids" binding:"required"`
		DestinationPath string `json:"destination_path" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	// 验证目标路径
	reqPath, err := user.JoinPath(req.DestinationPath)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	// 获取搜索结果
	var results []model.SearchResult
	if err := db.db.Where("id IN ?", req.ResultIDs).Find(&results).Error; err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	var taskIDs []string
	var errors []string

	for _, result := range results {
		// 确定下载URL
		var downloadURL string
		if result.MagnetLink != "" {
			downloadURL = result.MagnetLink
		} else if result.TorrentURL != "" {
			downloadURL = result.TorrentURL
		} else {
			errors = append(errors, "no download URL for: "+result.Title)
			continue
		}

		// 添加到离线下载队列
		ctx := context.Background()
		task, err := tool.AddURL(ctx, &tool.AddURLArgs{
			URL:          downloadURL,
			DstDirPath:   reqPath,
			Tool:         "aria2",
			DeletePolicy: tool.DeleteOnUploadSucceed,
		})

		if err != nil {
			errors = append(errors, "failed to create task for: "+result.Title)
			continue
		}

		taskIDs = append(taskIDs, task.GetID())
	}

	response := gin.H{
		"task_ids":      taskIDs,
		"created_count": len(taskIDs),
		"total_count":   len(results),
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	common.SuccessResp(c, response)
}

func Search(c *gin.Context) {
	var (
		req SearchReq
		err error
	)
	if err = c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	req.Parent, err = user.JoinPath(req.Parent)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := req.Validate(); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	nodes, total, err := search.Search(c, req.SearchReq)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	var filteredNodes []model.SearchNode
	for _, node := range nodes {
		if !strings.HasPrefix(node.Parent, user.BasePath) {
			continue
		}
		meta, err := op.GetNearestMeta(node.Parent)
		if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			continue
		}
		if !common.CanAccess(user, meta, path.Join(node.Parent, node.Name), req.Password) {
			continue
		}
		filteredNodes = append(filteredNodes, node)
	}
	common.SuccessResp(c, common.PageResp{
		Content: utils.MustSliceConvert(filteredNodes, nodeToSearchResp),
		Total:   total,
	})
}

func nodeToSearchResp(node model.SearchNode) SearchResp {
	return SearchResp{
		SearchNode: node,
		Type:       utils.GetObjType(node.Name, node.IsDir),
	}
}
