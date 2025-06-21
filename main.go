package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/OpenListTeam/OpenList/drivers"
	"github.com/OpenListTeam/OpenList/internal/driver"
	"github.com/OpenListTeam/OpenList/internal/model"
	"github.com/OpenListTeam/OpenList/internal/op"
	"github.com/OpenListTeam/OpenList/internal/stream"
	"github.com/syumai/workers"
)

// JWT 配置
const (
	JWT_SECRET     = "openlist-workers-secret-key-2024"
	JWT_EXPIRATION = 24 * time.Hour // 24小时过期
)

// JWT Claims 结构
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// 认证响应结构
type AuthResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// 用户注册请求结构
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	BasePath string `json:"base_path,omitempty"`
}

// 用户登录请求结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// API 响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 分页响应结构
type PageResponse struct {
	Content interface{} `json:"content"`
	Total   int64       `json:"total"`
}

// 分页请求结构
type PageRequest struct {
	Page    int `json:"page" form:"page"`
	PerPage int `json:"per_page" form:"per_page"`
}

func (p *PageRequest) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
}

// 离线下载配置结构
type OfflineDownloadConfig struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`       // 关联用户ID
	ToolName    string `json:"tool_name"`     // 工具名称：aria2, qbittorrent, transmission, 115, pikpak, thunder
	Config      string `json:"config"`        // JSON 格式的配置信息
	TempDirPath string `json:"temp_dir_path"` // 临时目录路径（对于云盘类工具）
	Enabled     bool   `json:"enabled"`       // 是否启用
	Created     string `json:"created"`       // 创建时间
	Modified    string `json:"modified"`      // 修改时间
}

// 离线下载任务结构
type OfflineDownloadTask struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`       // 关联用户ID
	ConfigID     uint   `json:"config_id"`     // 驱动配置ID
	URLs         string `json:"urls"`          // JSON 格式的URL列表
	DstPath      string `json:"dst_path"`      // 目标路径
	Tool         string `json:"tool"`          // 使用的工具
	Status       string `json:"status"`        // 任务状态：pending, running, completed, failed
	Progress     int    `json:"progress"`      // 进度百分比
	DeletePolicy string `json:"delete_policy"` // 删除策略
	Error        string `json:"error"`         // 错误信息
	Created      string `json:"created"`       // 创建时间
	Updated      string `json:"updated"`       // 更新时间
}

// 全局变量
var (
	dbManager *D1DatabaseManager // D1 数据库管理器
	// 内存中的用户映射，用于缓存
	usersMap = make(map[uint]*model.User)
	// 内存中的驱动配置映射，用于缓存
	driversMap = make(map[string]*DriverConfig)
	// 用户驱动实例映射，用于缓存已初始化的驱动
	userDriverInstances = make(map[string]driver.Driver)
	// 离线下载配置映射，用于缓存
	offlineDownloadConfigs = make(map[string]*OfflineDownloadConfig)
	// 离线下载任务映射，用于缓存
	offlineDownloadTasks = make(map[uint]*OfflineDownloadTask)
)

// 驱动配置结构（基于用户）
type DriverConfig struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"` // 关联用户ID
	Name        string `json:"name"`    // 驱动名称
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Config      string `json:"config"`   // JSON 格式的配置信息
	Icon        string `json:"icon"`     // 驱动图标
	Enabled     bool   `json:"enabled"`  // 是否启用
	Order       int    `json:"order"`    // 排序
	Created     string `json:"created"`  // 创建时间
	Modified    string `json:"modified"` // 修改时间
}

// 用户驱动实例键生成
func makeUserDriverKey(userID, configID uint) string {
	return fmt.Sprintf("user_%d_config_%d", userID, configID)
}

// 获取或创建用户驱动实例
func getUserDriverInstance(ctx context.Context, userID, configID uint) (driver.Driver, error) {
	key := makeUserDriverKey(userID, configID)

	// 检查缓存
	if instance, exists := userDriverInstances[key]; exists {
		return instance, nil
	}

	// 获取驱动配置
	config, err := getDriverConfigById(configID)
	if err != nil {
		return nil, fmt.Errorf("driver config not found: %w", err)
	}

	// 验证配置属于指定用户
	if config.UserID != userID {
		return nil, fmt.Errorf("driver config %d does not belong to user %d", configID, userID)
	}

	// 检查驱动是否启用
	if !config.Enabled {
		return nil, fmt.Errorf("driver config %d is disabled", configID)
	}

	// 获取驱动构造函数
	driverConstructor, err := op.GetDriver(config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver %s: %w", config.Name, err)
	}

	// 创建驱动实例
	driverInstance := driverConstructor()

	// 创建存储模型（用于兼容现有驱动接口）
	storage := model.Storage{
		ID:       config.ID,
		Driver:   config.Name,
		Addition: config.Config,
		Disabled: !config.Enabled,
	}

	driverInstance.SetStorage(storage)

	// 解析配置
	if config.Config != "" {
		if err := json.Unmarshal([]byte(config.Config), driverInstance.GetAddition()); err != nil {
			return nil, fmt.Errorf("failed to unmarshal driver config: %w", err)
		}
	}

	// 初始化驱动
	if err := driverInstance.Init(ctx); err != nil {
		return nil, fmt.Errorf("failed to init driver: %w", err)
	}

	// 缓存实例
	userDriverInstances[key] = driverInstance

	return driverInstance, nil
}

// 文件系统操作结构
type FileSystemRequest struct {
	UserID   uint   `json:"user_id" form:"user_id"`
	ConfigID uint   `json:"config_id" form:"config_id"`
	Path     string `json:"path" form:"path"`
}

// 文件系统响应结构
type FileSystemResponse struct {
	APIResponse
	Files []model.Obj `json:"files,omitempty"`
	File  *model.Obj  `json:"file,omitempty"`
}

// 获取当前用户和配置ID
func parseFileSystemRequest(r *http.Request) (*FileSystemRequest, error) {
	req := &FileSystemRequest{}

	// 解析用户ID
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = r.FormValue("user_id")
	}
	if userIDStr == "" {
		req.UserID = getCurrentUserID(r) // 使用默认方法
	} else {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id: %w", err)
		}
		req.UserID = uint(userID)
	}

	// 解析配置ID
	configIDStr := r.URL.Query().Get("config_id")
	if configIDStr == "" {
		configIDStr = r.FormValue("config_id")
	}
	if configIDStr == "" {
		return nil, fmt.Errorf("config_id is required")
	}

	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid config_id: %w", err)
	}
	req.ConfigID = uint(configID)

	// 解析路径
	req.Path = r.URL.Query().Get("path")
	if req.Path == "" {
		req.Path = r.FormValue("path")
	}
	if req.Path == "" {
		req.Path = "/"
	}

	return req, nil
}

// 文件系统列表处理器
func handleFileSystemList(w http.ResponseWriter, r *http.Request) {
	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 创建目录对象
	dir := &model.Object{
		Path: req.Path,
	}

	// 列出文件
	files, err := driverInstance.List(r.Context(), dir, model.ListArgs{})
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to list files: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code: 200,
		Data: map[string]interface{}{
			"files":     files,
			"path":      req.Path,
			"user_id":   req.UserID,
			"config_id": req.ConfigID,
		},
	})
}

// 文件系统获取文件处理器
func handleFileSystemGet(w http.ResponseWriter, r *http.Request) {
	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 检查驱动是否支持 Get 方法
	getter, ok := driverInstance.(driver.Getter)
	if !ok {
		respondJSON(w, APIResponse{
			Code:    501,
			Message: "Driver does not support get operation",
		})
		return
	}

	// 获取文件信息
	fileInfo, err := getter.Get(r.Context(), req.Path)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to get file: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code: 200,
		Data: map[string]interface{}{
			"file":      fileInfo,
			"path":      req.Path,
			"user_id":   req.UserID,
			"config_id": req.ConfigID,
		},
	})
}

// 文件下载处理器
func handleFileSystemDownload(w http.ResponseWriter, r *http.Request) {
	req, err := parseFileSystemRequest(r)
	if err != nil {
		http.Error(w, "Invalid request: "+err.Error(), 400)
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		http.Error(w, "Failed to get driver instance: "+err.Error(), 404)
		return
	}

	// 创建文件对象
	file := &model.Object{
		Path: req.Path,
	}

	// 获取文件链接
	link, err := driverInstance.Link(r.Context(), file, model.LinkArgs{})
	if err != nil {
		http.Error(w, "Failed to get file link: "+err.Error(), 500)
		return
	}

	// 重定向到实际文件链接
	http.Redirect(w, r, link.URL, http.StatusFound)
}

// 创建目录处理器
func handleFileSystemMkdir(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 获取目录名
	dirName := r.FormValue("dir_name")
	if dirName == "" {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "dir_name is required",
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 检查驱动是否支持 MakeDir 方法
	mkdir, ok := driverInstance.(driver.Mkdir)
	if !ok {
		respondJSON(w, APIResponse{
			Code:    501,
			Message: "Driver does not support mkdir operation",
		})
		return
	}

	// 创建父目录对象
	parentDir := &model.Object{
		Path: req.Path,
	}

	// 创建目录
	err = mkdir.MakeDir(r.Context(), parentDir, dirName)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to create directory: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Directory created successfully",
		Data: map[string]interface{}{
			"parent_path": req.Path,
			"dir_name":    dirName,
			"user_id":     req.UserID,
			"config_id":   req.ConfigID,
		},
	})
}

// 删除文件/目录处理器
func handleFileSystemRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 检查驱动是否支持 Remove 方法
	remove, ok := driverInstance.(driver.Remove)
	if !ok {
		respondJSON(w, APIResponse{
			Code:    501,
			Message: "Driver does not support remove operation",
		})
		return
	}

	// 创建文件对象
	file := &model.Object{
		Path: req.Path,
	}

	// 删除文件
	err = remove.Remove(r.Context(), file)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to remove file: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "File removed successfully",
		Data: map[string]interface{}{
			"path":      req.Path,
			"user_id":   req.UserID,
			"config_id": req.ConfigID,
		},
	})
}

// 重命名文件/目录处理器
func handleFileSystemRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 解析新名称
	newName := r.FormValue("new_name")
	if newName == "" {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "new_name is required",
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 检查驱动是否支持 Rename 方法
	rename, ok := driverInstance.(driver.Rename)
	if !ok {
		respondJSON(w, APIResponse{
			Code:    501,
			Message: "Driver does not support rename operation",
		})
		return
	}

	// 创建源文件对象
	srcFile := &model.Object{
		Path: req.Path,
	}

	// 重命名文件
	err = rename.Rename(r.Context(), srcFile, newName)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to rename file: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "File renamed successfully",
		Data: map[string]interface{}{
			"old_path":  req.Path,
			"new_name":  newName,
			"user_id":   req.UserID,
			"config_id": req.ConfigID,
		},
	})
}

// 移动文件/目录处理器
func handleFileSystemMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 解析目标路径
	dstPath := r.FormValue("dst_path")
	if dstPath == "" {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "dst_path is required",
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 检查驱动是否支持 Move 方法
	move, ok := driverInstance.(driver.Move)
	if !ok {
		respondJSON(w, APIResponse{
			Code:    501,
			Message: "Driver does not support move operation",
		})
		return
	}

	// 创建源和目标文件对象
	srcFile := &model.Object{Path: req.Path}
	dstDir := &model.Object{Path: dstPath}

	// 移动文件
	err = move.Move(r.Context(), srcFile, dstDir)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to move file: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "File moved successfully",
		Data: map[string]interface{}{
			"src_path":  req.Path,
			"dst_path":  dstPath,
			"user_id":   req.UserID,
			"config_id": req.ConfigID,
		},
	})
}

// 复制文件/目录处理器
func handleFileSystemCopy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 解析目标路径
	dstPath := r.FormValue("dst_path")
	if dstPath == "" {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "dst_path is required",
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 检查驱动是否支持 Copy 方法
	copy, ok := driverInstance.(driver.Copy)
	if !ok {
		respondJSON(w, APIResponse{
			Code:    501,
			Message: "Driver does not support copy operation",
		})
		return
	}

	// 创建源和目标文件对象
	srcFile := &model.Object{Path: req.Path}
	dstDir := &model.Object{Path: dstPath}

	// 复制文件
	err = copy.Copy(r.Context(), srcFile, dstDir)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to copy file: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "File copied successfully",
		Data: map[string]interface{}{
			"src_path":  req.Path,
			"dst_path":  dstPath,
			"user_id":   req.UserID,
			"config_id": req.ConfigID,
		},
	})
}

// 文件上传处理器
func handleFileSystemUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 获取文件名
	fileName := r.URL.Query().Get("filename")
	if fileName == "" {
		fileName = r.FormValue("filename")
	}
	if fileName == "" {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "filename is required",
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 检查驱动是否支持 Put 方法
	put, ok := driverInstance.(driver.Put)
	if !ok {
		respondJSON(w, APIResponse{
			Code:    501,
			Message: "Driver does not support upload operation",
		})
		return
	}

	// 创建目标目录对象
	dstDir := &model.Object{
		Path: req.Path,
	}

	// 创建文件流对象
	fileStreamer := &stream.FileStream{
		Obj: &model.Object{
			Name: fileName,
			Size: r.ContentLength,
		},
		Reader: r.Body,
	}

	// 上传文件，使用简单的进度更新函数
	progressFn := func(percentage float64) {}
	err = put.Put(r.Context(), dstDir, fileStreamer, progressFn)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to upload file: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "File uploaded successfully",
		Data: map[string]interface{}{
			"path":      req.Path,
			"filename":  fileName,
			"size":      r.ContentLength,
			"user_id":   req.UserID,
			"config_id": req.ConfigID,
		},
	})
}

// 获取文件目录列表处理器
func handleFileSystemDirs(w http.ResponseWriter, r *http.Request) {
	req, err := parseFileSystemRequest(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 获取驱动实例
	driverInstance, err := getUserDriverInstance(r.Context(), req.UserID, req.ConfigID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "Failed to get driver instance: " + err.Error(),
		})
		return
	}

	// 创建目录对象
	dir := &model.Object{
		Path: req.Path,
	}

	// 列出目录
	files, err := driverInstance.List(r.Context(), dir, model.ListArgs{})
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to list directories: " + err.Error(),
		})
		return
	}

	// 过滤出目录
	var dirs []model.Obj
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file)
		}
	}

	respondJSON(w, APIResponse{
		Code: 200,
		Data: map[string]interface{}{
			"dirs":      dirs,
			"path":      req.Path,
			"user_id":   req.UserID,
			"config_id": req.ConfigID,
		},
	})
}

// 初始化 D1 数据库
func initD1DatabaseWithManager(dbName string) error {
	var err error
	dbManager, err = NewD1DatabaseManager(dbName)
	if err != nil {
		return fmt.Errorf("failed to create D1 database manager: %w", err)
	}

	// 创建表
	ctx := context.Background()
	if err := dbManager.CreateTables(ctx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// 初始化默认数据（用户和驱动配置）
	if err := initDefaultData(); err != nil {
		return fmt.Errorf("failed to initialize default data: %w", err)
	}

	return nil
}

// 初始化默认数据（用户和驱动配置）
func initDefaultData() error {
	// 创建默认管理员用户
	adminUser := model.User{
		ID:         1,
		Username:   "admin",
		BasePath:   "/",
		Role:       model.ADMIN,
		Disabled:   false,
		Permission: 0x30FF,
		Authn:      "[]",
	}
	adminUser.SetPassword("admin123") // 默认密码
	usersMap[adminUser.ID] = &adminUser

	// 为管理员创建默认驱动配置
	defaultDrivers := []DriverConfig{
		{
			UserID:      1,
			Name:        "Local",
			DisplayName: "本地存储",
			Description: "本地文件系统存储",
			Config:      `{"root_folder_path": "/data"}`,
			Icon:        "folder",
			Enabled:     true,
			Order:       1,
		},
		{
			UserID:      1,
			Name:        "S3",
			DisplayName: "Amazon S3",
			Description: "Amazon S3 对象存储",
			Config:      `{"bucket": "", "region": "us-east-1", "access_key_id": "", "secret_access_key": ""}`,
			Icon:        "cloud",
			Enabled:     true,
			Order:       2,
		},
		{
			UserID:      1,
			Name:        "Aliyundrive",
			DisplayName: "阿里云盘",
			Description: "阿里云盘存储",
			Config:      `{"refresh_token": "", "root_folder_id": "root"}`,
			Icon:        "cloud",
			Enabled:     true,
			Order:       3,
		},
		{
			UserID:      1,
			Name:        "OneDrive",
			DisplayName: "OneDrive",
			Description: "Microsoft OneDrive 存储",
			Config:      `{"client_id": "", "client_secret": "", "redirect_uri": ""}`,
			Icon:        "cloud",
			Enabled:     true,
			Order:       4,
		},
		{
			UserID:      1,
			Name:        "GoogleDrive",
			DisplayName: "Google Drive",
			Description: "Google Drive 存储",
			Config:      `{"client_id": "", "client_secret": "", "redirect_uri": ""}`,
			Icon:        "cloud",
			Enabled:     true,
			Order:       5,
		},
	}

	// 初始化驱动配置
	for i, driver := range defaultDrivers {
		driver.ID = uint(i + 1)
		driver.Created = time.Now().Format(time.RFC3339)
		driver.Modified = time.Now().Format(time.RFC3339)
		driversMap[fmt.Sprintf("%d_%s", driver.UserID, driver.Name)] = &driver

		// 如果有数据库管理器，同步创建到数据库
		if dbManager != nil {
			if err := dbManager.CreateDriverConfig(context.Background(), driver); err != nil {
				fmt.Printf("Failed to create driver config %s: %v\n", driver.Name, err)
			}
		}

		fmt.Printf("Initialized driver config: %s for user %d\n", driver.Name, driver.UserID)
	}

	return nil
}

// 驱动配置相关函数

// 获取用户的驱动配置列表
func getUserDriverConfigs(userID uint, page, perPage int) ([]DriverConfig, int64, error) {
	if dbManager != nil {
		return dbManager.GetUserDriverConfigs(context.Background(), userID, page, perPage)
	}

	// 回退到内存操作
	var drivers []DriverConfig
	for _, driver := range driversMap {
		if driver.UserID == userID {
			drivers = append(drivers, *driver)
		}
	}

	// 简单排序
	for i := 0; i < len(drivers)-1; i++ {
		for j := i + 1; j < len(drivers); j++ {
			if drivers[i].Order > drivers[j].Order {
				drivers[i], drivers[j] = drivers[j], drivers[i]
			}
		}
	}

	start := (page - 1) * perPage
	end := start + perPage
	total := int64(len(drivers))

	if start > len(drivers) {
		return []DriverConfig{}, total, nil
	}
	if end > len(drivers) {
		end = len(drivers)
	}

	return drivers[start:end], total, nil
}

// 根据用户ID和名称获取驱动配置
func getUserDriverConfigByName(userID uint, name string) (*DriverConfig, error) {
	key := fmt.Sprintf("%d_%s", userID, name)
	if driver, exists := driversMap[key]; exists {
		return driver, nil
	}
	return nil, fmt.Errorf("driver config not found: %s for user %d", name, userID)
}

// 根据ID获取驱动配置
func getDriverConfigById(id uint) (*DriverConfig, error) {
	for _, driver := range driversMap {
		if driver.ID == id {
			return driver, nil
		}
	}
	return nil, fmt.Errorf("driver config not found with id: %d", id)
}

// 创建用户驱动配置
func createUserDriverConfig(ctx context.Context, userID uint, driver DriverConfig) error {
	driver.UserID = userID

	if dbManager != nil {
		return dbManager.CreateDriverConfig(ctx, driver)
	}

	// 回退到内存操作
	driver.ID = uint(time.Now().Unix())
	driver.Created = time.Now().Format(time.RFC3339)
	driver.Modified = time.Now().Format(time.RFC3339)

	key := fmt.Sprintf("%d_%s", userID, driver.Name)
	driversMap[key] = &driver
	fmt.Printf("Created driver config: %s for user %d (memory mode)\n", driver.Name, userID)
	return nil
}

// 更新用户驱动配置
func updateUserDriverConfig(ctx context.Context, userID uint, driver DriverConfig) error {
	driver.UserID = userID

	if dbManager != nil {
		return dbManager.UpdateDriverConfig(ctx, driver)
	}

	// 回退到内存操作
	key := fmt.Sprintf("%d_%s", userID, driver.Name)
	if existing, exists := driversMap[key]; exists {
		driver.ID = existing.ID
		driver.Created = existing.Created
		driver.Modified = time.Now().Format(time.RFC3339)
		driversMap[key] = &driver
		fmt.Printf("Updated driver config: %s for user %d (memory mode)\n", driver.Name, userID)
		return nil
	}
	return fmt.Errorf("driver config not found: %s for user %d", driver.Name, userID)
}

// 删除用户驱动配置
func deleteUserDriverConfigById(ctx context.Context, userID, id uint) error {
	if dbManager != nil {
		return dbManager.DeleteUserDriverConfig(ctx, userID, id)
	}

	// 回退到内存操作
	for key, driver := range driversMap {
		if driver.ID == id && driver.UserID == userID {
			delete(driversMap, key)
			fmt.Printf("Deleted driver config: %s for user %d (memory mode)\n", driver.Name, userID)
			return nil
		}
	}
	return fmt.Errorf("driver config not found with id: %d for user %d", id, userID)
}

// 启用/禁用用户驱动配置
func toggleUserDriverConfig(ctx context.Context, userID, id uint, enabled bool) error {
	for _, driver := range driversMap {
		if driver.ID == id && driver.UserID == userID {
			driver.Enabled = enabled
			driver.Modified = time.Now().Format(time.RFC3339)

			// 如果有数据库管理器，同步更新到数据库
			if dbManager != nil {
				return dbManager.UpdateDriverConfig(ctx, *driver)
			}

			action := "disable"
			if enabled {
				action = "enable"
			}
			fmt.Printf("Would %s driver config: %s for user %d (memory mode)\n", action, driver.Name, userID)
			return nil
		}
	}
	return fmt.Errorf("driver config not found with id: %d for user %d", id, userID)
}

// 用户相关函数

// 获取用户列表
func getUsers(page, perPage int) ([]model.User, int64, error) {
	if dbManager != nil {
		return dbManager.GetUsers(context.Background(), page, perPage)
	}

	// 回退到内存/模拟数据
	users := []model.User{
		{
			ID:         1,
			Username:   "admin",
			BasePath:   "/",
			Role:       model.ADMIN,
			Disabled:   false,
			Permission: 0x30FF,
		},
		{
			ID:         2,
			Username:   "guest",
			BasePath:   "/",
			Role:       model.GUEST,
			Disabled:   true,
			Permission: 0,
		},
	}

	start := (page - 1) * perPage
	end := start + perPage
	if start > len(users) {
		return []model.User{}, int64(len(users)), nil
	}
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], int64(len(users)), nil
}

// 根据ID获取用户
func getUserById(id uint) (*model.User, error) {
	// 先从缓存查找
	if user, exists := usersMap[id]; exists {
		return user, nil
	}

	// 创建模拟用户
	user := &model.User{
		ID:         id,
		Username:   fmt.Sprintf("user%d", id),
		BasePath:   "/",
		Role:       model.GENERAL,
		Disabled:   false,
		Permission: 0,
	}
	usersMap[id] = user
	return user, nil
}

// 创建用户
func createUser(ctx context.Context, user model.User) error {
	if dbManager != nil {
		return dbManager.CreateUser(ctx, user)
	}

	// 回退到内存操作
	user.ID = uint(time.Now().Unix())
	if user.Authn == "" {
		user.Authn = "[]"
	}
	usersMap[user.ID] = &user
	fmt.Printf("Created user: %s (memory mode)\n", user.Username)
	return nil
}

// 更新用户
func updateUser(ctx context.Context, user model.User) error {
	usersMap[user.ID] = &user
	fmt.Printf("Updated user: %s (memory mode)\n", user.Username)
	return nil
}

// 删除用户
func deleteUserById(ctx context.Context, id uint) error {
	delete(usersMap, id)
	fmt.Printf("Deleted user: %d (memory mode)\n", id)
	return nil
}

// 辅助函数
func respondJSON(w http.ResponseWriter, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(response)
}

func parseJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// 获取当前用户ID（更新为使用JWT认证）
func getCurrentUserID(r *http.Request) uint {
	claims, err := getCurrentAuthenticatedUser(r)
	if err != nil {
		// 认证失败，返回0表示未认证
		return 0
	}
	return claims.UserID
}

// 用户认证相关 API 处理器

// 用户注册处理器
func handleUserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	var req RegisterRequest
	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// 验证必需字段
	if req.Username == "" || req.Password == "" {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Username and password are required",
		})
		return
	}

	// 检查用户名长度
	if len(req.Username) < 3 || len(req.Username) > 50 {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Username must be between 3 and 50 characters",
		})
		return
	}

	// 检查密码强度
	if len(req.Password) < 6 {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Password must be at least 6 characters",
		})
		return
	}

	// 检查用户名是否已存在
	for _, user := range usersMap {
		if user.Username == req.Username {
			respondJSON(w, APIResponse{
				Code:    409,
				Message: "Username already exists",
			})
			return
		}
	}

	// 创建新用户
	newUser := model.User{
		Username:   req.Username,
		BasePath:   req.BasePath,
		Role:       model.GENERAL, // 默认为普通用户
		Disabled:   false,
		Permission: 0,
		Authn:      "[]",
	}

	if newUser.BasePath == "" {
		newUser.BasePath = "/"
	}

	// 设置密码
	newUser.SetPassword(req.Password)

	// 保存用户
	if err := createUser(r.Context(), newUser); err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to create user: " + err.Error(),
		})
		return
	}

	// 更新内存中的用户映射
	newUser.ID = uint(time.Now().Unix()) // 生成简单的ID
	usersMap[newUser.ID] = &newUser

	// 生成 JWT Token
	token, err := generateJWTToken(&newUser)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to generate token: " + err.Error(),
		})
		return
	}

	// 清除密码信息
	newUser.Password = ""
	newUser.PwdHash = ""
	newUser.Salt = ""

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "User registered successfully",
		Data: AuthResponse{
			Token: token,
			User:  &newUser,
		},
	})
}

// 用户登录处理器
func handleUserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	var req LoginRequest
	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// 验证必需字段
	if req.Username == "" || req.Password == "" {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Username and password are required",
		})
		return
	}

	// 查找用户
	var user *model.User
	for _, u := range usersMap {
		if u.Username == req.Username {
			user = u
			break
		}
	}

	if user == nil {
		respondJSON(w, APIResponse{
			Code:    401,
			Message: "Invalid username or password",
		})
		return
	}

	// 检查用户是否被禁用
	if user.Disabled {
		respondJSON(w, APIResponse{
			Code:    403,
			Message: "User account is disabled",
		})
		return
	}

	// 验证密码
	if err := user.ValidateRawPassword(req.Password); err != nil {
		respondJSON(w, APIResponse{
			Code:    401,
			Message: "Invalid username or password",
		})
		return
	}

	// 生成 JWT Token
	token, err := generateJWTToken(user)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to generate token: " + err.Error(),
		})
		return
	}

	// 清除敏感信息
	userResponse := *user
	userResponse.Password = ""
	userResponse.PwdHash = ""
	userResponse.Salt = ""
	userResponse.OtpSecret = ""

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Login successful",
		Data: AuthResponse{
			Token: token,
			User:  &userResponse,
		},
	})
}

// 获取当前用户信息处理器
func handleCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, err := getCurrentAuthenticatedUser(r)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    401,
			Message: "Authentication required: " + err.Error(),
		})
		return
	}

	user, err := getUserById(claims.UserID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    404,
			Message: "User not found: " + err.Error(),
		})
		return
	}

	// 清除敏感信息
	userResponse := *user
	userResponse.Password = ""
	userResponse.PwdHash = ""
	userResponse.Salt = ""
	userResponse.OtpSecret = ""

	respondJSON(w, APIResponse{
		Code: 200,
		Data: userResponse,
	})
}

// 用户登出处理器（前端清除token即可）
func handleUserLogout(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Logout successful",
		Data: map[string]interface{}{
			"message": "Please remove the token from client side",
		},
	})
}

// 认证包装器 - 为需要认证的API添加认证检查
func requireAuth(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getCurrentAuthenticatedUser(r)
		if err != nil {
			respondJSON(w, APIResponse{
				Code:    401,
				Message: "Authentication required: " + err.Error(),
			})
			return
		}
		handler(w, r)
	}
}

// 管理员权限包装器
func requireAdmin(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := getCurrentAuthenticatedUser(r)
		if err != nil {
			respondJSON(w, APIResponse{
				Code:    401,
				Message: "Authentication required: " + err.Error(),
			})
			return
		}

		if !checkUserPermission(claims, int(model.ADMIN)) {
			respondJSON(w, APIResponse{
				Code:    403,
				Message: "Admin permission required",
			})
			return
		}

		handler(w, r)
	}
}

// 驱动配置 API 处理器
func handleUserDriversAPI(w http.ResponseWriter, r *http.Request) {
	userID := getCurrentUserID(r)

	switch r.Method {
	case "GET":
		// 获取用户的驱动配置列表
		page := 1
		perPage := 20

		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
			if p, err := strconv.Atoi(perPageStr); err == nil && p > 0 {
				perPage = p
			}
		}

		// 获取启用状态过滤参数
		enabledOnly := r.URL.Query().Get("enabled") == "true"

		drivers, total, err := getUserDriverConfigs(userID, page, perPage)
		if err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to get user driver configs: " + err.Error(),
			})
			return
		}

		// 如果只要启用的驱动，进行过滤
		if enabledOnly {
			var enabledDrivers []DriverConfig
			for _, driver := range drivers {
				if driver.Enabled {
					enabledDrivers = append(enabledDrivers, driver)
				}
			}
			drivers = enabledDrivers
		}

		// 兼容旧 API 格式
		driverNames := make([]string, 0)
		driverInfoMap := make(map[string]interface{})

		for _, driver := range drivers {
			if driver.Enabled {
				driverNames = append(driverNames, driver.Name)
				driverInfoMap[driver.Name] = map[string]interface{}{
					"name":         driver.Name,
					"display_name": driver.DisplayName,
					"description":  driver.Description,
					"icon":         driver.Icon,
					"config":       driver.Config,
					"order":        driver.Order,
				}
			}
		}

		respondJSON(w, APIResponse{
			Code: 200,
			Data: map[string]interface{}{
				"drivers":  driverNames,
				"info":     driverInfoMap,
				"configs":  drivers,
				"total":    total,
				"page":     page,
				"per_page": perPage,
				"user_id":  userID,
			},
		})

	case "POST":
		// 创建用户驱动配置
		var driver DriverConfig
		if err := parseJSON(r, &driver); err != nil {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Invalid JSON: " + err.Error(),
			})
			return
		}

		if driver.Name == "" {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Driver name is required",
			})
			return
		}

		if err := createUserDriverConfig(r.Context(), userID, driver); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to create user driver config: " + err.Error(),
			})
			return
		}

		respondJSON(w, APIResponse{
			Code:    200,
			Message: "User driver config created successfully",
		})

	default:
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
	}
}

// 单个用户驱动配置 API 处理器
func handleUserDriverAPI(w http.ResponseWriter, r *http.Request) {
	userID := getCurrentUserID(r)

	switch r.Method {
	case "GET":
		// 获取单个用户驱动配置
		name := r.URL.Query().Get("name")
		idStr := r.URL.Query().Get("id")

		var driver *DriverConfig
		var err error

		if name != "" {
			driver, err = getUserDriverConfigByName(userID, name)
		} else if idStr != "" {
			id, parseErr := strconv.ParseUint(idStr, 10, 32)
			if parseErr != nil {
				respondJSON(w, APIResponse{
					Code:    400,
					Message: "Invalid driver ID",
				})
				return
			}
			driver, err = getDriverConfigById(uint(id))
			// 验证驱动是否属于当前用户
			if err == nil && driver.UserID != userID {
				err = fmt.Errorf("driver config not accessible for user %d", userID)
			}
		} else {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Driver name or ID is required",
			})
			return
		}

		if err != nil {
			respondJSON(w, APIResponse{
				Code:    404,
				Message: "Driver config not found: " + err.Error(),
			})
			return
		}

		respondJSON(w, APIResponse{
			Code: 200,
			Data: driver,
		})

	case "POST":
		// 更新用户驱动配置
		var driver DriverConfig
		if err := parseJSON(r, &driver); err != nil {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Invalid JSON: " + err.Error(),
			})
			return
		}

		if driver.Name == "" {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Driver name is required",
			})
			return
		}

		if err := updateUserDriverConfig(r.Context(), userID, driver); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to update user driver config: " + err.Error(),
			})
			return
		}

		respondJSON(w, APIResponse{
			Code:    200,
			Message: "User driver config updated successfully",
		})

	default:
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
	}
}

// 删除用户驱动配置 API 处理器
func handleDeleteUserDriverAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid driver ID",
		})
		return
	}

	if err := deleteUserDriverConfigById(r.Context(), userID, uint(id)); err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to delete user driver config: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "User driver config deleted successfully",
	})
}

// 启用用户驱动配置 API 处理器
func handleEnableUserDriverAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid driver ID",
		})
		return
	}

	if err := toggleUserDriverConfig(r.Context(), userID, uint(id), true); err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to enable user driver config: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "User driver config enabled successfully",
	})
}

// 禁用用户驱动配置 API 处理器
func handleDisableUserDriverAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid driver ID",
		})
		return
	}

	if err := toggleUserDriverConfig(r.Context(), userID, uint(id), false); err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to disable user driver config: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "User driver config disabled successfully",
	})
}

// 用户管理 API 处理器
func handleUsersAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// 获取用户列表
		var req PageRequest
		page := 1
		perPage := 20

		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
			if p, err := strconv.Atoi(perPageStr); err == nil && p > 0 {
				perPage = p
			}
		}

		req.Page = page
		req.PerPage = perPage
		req.Validate()

		users, total, err := getUsers(req.Page, req.PerPage)
		if err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to get users: " + err.Error(),
			})
			return
		}

		respondJSON(w, APIResponse{
			Code: 200,
			Data: PageResponse{
				Content: users,
				Total:   total,
			},
		})

	case "POST":
		// 创建用户
		var user model.User
		if err := parseJSON(r, &user); err != nil {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Invalid JSON: " + err.Error(),
			})
			return
		}

		if user.IsAdmin() || user.IsGuest() {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "admin or guest user can not be created",
			})
			return
		}

		user.SetPassword(user.Password)
		user.Password = ""
		user.Authn = "[]"

		if err := createUser(r.Context(), user); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to create user: " + err.Error(),
			})
			return
		}

		respondJSON(w, APIResponse{
			Code:    200,
			Message: "User created successfully",
		})

	default:
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
	}
}

// 单个用户 API 处理器
func handleUserAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// 获取单个用户
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Invalid user ID",
			})
			return
		}

		user, err := getUserById(uint(id))
		if err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to get user: " + err.Error(),
			})
			return
		}

		respondJSON(w, APIResponse{
			Code: 200,
			Data: user,
		})

	case "POST":
		// 更新用户
		var user model.User
		if err := parseJSON(r, &user); err != nil {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Invalid JSON: " + err.Error(),
			})
			return
		}

		oldUser, err := getUserById(user.ID)
		if err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to get user: " + err.Error(),
			})
			return
		}

		if oldUser.Role != user.Role {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "role can not be changed",
			})
			return
		}

		if user.Password == "" {
			user.PwdHash = oldUser.PwdHash
			user.Salt = oldUser.Salt
		} else {
			user.SetPassword(user.Password)
			user.Password = ""
		}

		if user.OtpSecret == "" {
			user.OtpSecret = oldUser.OtpSecret
		}

		if user.Disabled && user.IsAdmin() {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "admin user can not be disabled",
			})
			return
		}

		if err := updateUser(r.Context(), user); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to update user: " + err.Error(),
			})
			return
		}

		respondJSON(w, APIResponse{
			Code:    200,
			Message: "User updated successfully",
		})

	default:
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
	}
}

// 删除用户 API 处理器
func handleDeleteUserAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid user ID",
		})
		return
	}

	if err := deleteUserById(r.Context(), uint(id)); err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to delete user: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "User deleted successfully",
	})
}

// 离线下载配置管理函数

// 获取用户的离线下载配置列表
func getUserOfflineDownloadConfigs(userID uint) ([]*OfflineDownloadConfig, error) {
	if dbManager != nil {
		return dbManager.GetUserOfflineDownloadConfigs(context.Background(), userID)
	}

	// 回退到内存操作
	var configs []*OfflineDownloadConfig
	for _, config := range offlineDownloadConfigs {
		if config.UserID == userID {
			configs = append(configs, config)
		}
	}
	return configs, nil
}

// 根据用户ID和工具名称获取离线下载配置
func getUserOfflineDownloadConfigByTool(userID uint, toolName string) (*OfflineDownloadConfig, error) {
	key := fmt.Sprintf("%d_%s", userID, toolName)
	if config, exists := offlineDownloadConfigs[key]; exists {
		return config, nil
	}
	return nil, fmt.Errorf("offline download config not found: %s for user %d", toolName, userID)
}

// 创建用户离线下载配置
func createUserOfflineDownloadConfig(ctx context.Context, userID uint, config OfflineDownloadConfig) error {
	config.UserID = userID

	if dbManager != nil {
		return dbManager.CreateOfflineDownloadConfig(ctx, config)
	}

	// 回退到内存操作
	config.ID = uint(time.Now().Unix())
	config.Created = time.Now().Format(time.RFC3339)
	config.Modified = time.Now().Format(time.RFC3339)

	key := fmt.Sprintf("%d_%s", userID, config.ToolName)
	offlineDownloadConfigs[key] = &config
	fmt.Printf("Created offline download config: %s for user %d (memory mode)\n", config.ToolName, userID)
	return nil
}

// 更新用户离线下载配置
func updateUserOfflineDownloadConfig(ctx context.Context, userID uint, config OfflineDownloadConfig) error {
	config.UserID = userID

	if dbManager != nil {
		return dbManager.UpdateOfflineDownloadConfig(ctx, config)
	}

	// 回退到内存操作
	key := fmt.Sprintf("%d_%s", userID, config.ToolName)
	if existing, exists := offlineDownloadConfigs[key]; exists {
		config.ID = existing.ID
		config.Created = existing.Created
		config.Modified = time.Now().Format(time.RFC3339)
		offlineDownloadConfigs[key] = &config
		fmt.Printf("Updated offline download config: %s for user %d (memory mode)\n", config.ToolName, userID)
		return nil
	}
	return fmt.Errorf("offline download config not found: %s for user %d", config.ToolName, userID)
}

// 删除用户离线下载配置
func deleteUserOfflineDownloadConfig(ctx context.Context, userID uint, toolName string) error {
	if dbManager != nil {
		return dbManager.DeleteUserOfflineDownloadConfig(ctx, userID, toolName)
	}

	// 回退到内存操作
	key := fmt.Sprintf("%d_%s", userID, toolName)
	if _, exists := offlineDownloadConfigs[key]; exists {
		delete(offlineDownloadConfigs, key)
		fmt.Printf("Deleted offline download config: %s for user %d (memory mode)\n", toolName, userID)
		return nil
	}
	return fmt.Errorf("offline download config not found: %s for user %d", toolName, userID)
}

// 离线下载任务管理函数

// 获取用户的离线下载任务列表
func getUserOfflineDownloadTasks(userID uint, page, perPage int) ([]*OfflineDownloadTask, int64, error) {
	if dbManager != nil {
		return dbManager.GetUserOfflineDownloadTasks(context.Background(), userID, page, perPage)
	}

	// 回退到内存操作
	var tasks []*OfflineDownloadTask
	for _, task := range offlineDownloadTasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}

	// 简单排序（按创建时间倒序）
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].Created < tasks[j].Created {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}

	start := (page - 1) * perPage
	end := start + perPage
	total := int64(len(tasks))

	if start > len(tasks) {
		return []*OfflineDownloadTask{}, total, nil
	}
	if end > len(tasks) {
		end = len(tasks)
	}

	return tasks[start:end], total, nil
}

// 根据ID获取离线下载任务
func getOfflineDownloadTaskById(userID, taskID uint) (*OfflineDownloadTask, error) {
	if task, exists := offlineDownloadTasks[taskID]; exists && task.UserID == userID {
		return task, nil
	}
	return nil, fmt.Errorf("offline download task not found: %d for user %d", taskID, userID)
}

// 创建离线下载任务
func createOfflineDownloadTask(ctx context.Context, userID uint, task OfflineDownloadTask) error {
	task.UserID = userID

	if dbManager != nil {
		return dbManager.CreateOfflineDownloadTask(ctx, task)
	}

	// 回退到内存操作
	task.ID = uint(time.Now().Unix())
	task.Status = "pending"
	task.Progress = 0
	task.Created = time.Now().Format(time.RFC3339)
	task.Updated = time.Now().Format(time.RFC3339)

	offlineDownloadTasks[task.ID] = &task
	fmt.Printf("Created offline download task: %d for user %d (memory mode)\n", task.ID, userID)
	return nil
}

// 更新离线下载任务状态
func updateOfflineDownloadTaskStatus(ctx context.Context, userID, taskID uint, status string, progress int, errorMsg string) error {
	if dbManager != nil {
		return dbManager.UpdateOfflineDownloadTaskStatus(ctx, userID, taskID, status, progress, errorMsg)
	}

	// 回退到内存操作
	if task, exists := offlineDownloadTasks[taskID]; exists && task.UserID == userID {
		task.Status = status
		task.Progress = progress
		task.Error = errorMsg
		task.Updated = time.Now().Format(time.RFC3339)
		fmt.Printf("Updated offline download task: %d status to %s (memory mode)\n", taskID, status)
		return nil
	}
	return fmt.Errorf("offline download task not found: %d for user %d", taskID, userID)
}

// 删除离线下载任务
func deleteOfflineDownloadTask(ctx context.Context, userID, taskID uint) error {
	if dbManager != nil {
		return dbManager.DeleteOfflineDownloadTask(ctx, userID, taskID)
	}

	// 回退到内存操作
	if task, exists := offlineDownloadTasks[taskID]; exists && task.UserID == userID {
		delete(offlineDownloadTasks, taskID)
		fmt.Printf("Deleted offline download task: %d for user %d (memory mode)\n", taskID, userID)
		return nil
	}
	return fmt.Errorf("offline download task not found: %d for user %d", taskID, userID)
}

// 获取支持的离线下载工具列表
func getSupportedOfflineDownloadTools() []string {
	return []string{
		"aria2",        // Aria2 下载器
		"qbittorrent",  // qBittorrent 下载器
		"transmission", // Transmission 下载器
		"115",          // 115 云盘离线下载
		"pikpak",       // PikPak 离线下载
		"thunder",      // 迅雷离线下载
	}
}

// 验证驱动是否支持离线下载
func validateDriverForOfflineDownload(userID, configID uint, toolName string) error {
	// 获取驱动配置
	config, err := getDriverConfigById(configID)
	if err != nil {
		return fmt.Errorf("driver config not found: %w", err)
	}

	// 验证配置属于指定用户
	if config.UserID != userID {
		return fmt.Errorf("driver config %d does not belong to user %d", configID, userID)
	}

	// 验证驱动类型是否支持离线下载
	switch toolName {
	case "115":
		if config.Name != "115" {
			return fmt.Errorf("tool %s requires 115 driver, but got %s", toolName, config.Name)
		}
	case "pikpak":
		if config.Name != "PikPak" {
			return fmt.Errorf("tool %s requires PikPak driver, but got %s", toolName, config.Name)
		}
	case "thunder":
		if config.Name != "Thunder" {
			return fmt.Errorf("tool %s requires Thunder driver, but got %s", toolName, config.Name)
		}
	case "aria2", "qbittorrent", "transmission":
		// 这些工具不依赖特定驱动，只需要目标路径存在即可
		break
	default:
		return fmt.Errorf("unsupported offline download tool: %s", toolName)
	}

	return nil
}

// 离线下载配置 API 处理器

// 获取支持的离线下载工具列表
func handleOfflineDownloadTools(w http.ResponseWriter, r *http.Request) {
	tools := getSupportedOfflineDownloadTools()
	respondJSON(w, APIResponse{
		Code: 200,
		Data: tools,
	})
}

// 获取用户的离线下载配置列表
func handleUserOfflineDownloadConfigs(w http.ResponseWriter, r *http.Request) {
	userID := getCurrentUserID(r)

	configs, err := getUserOfflineDownloadConfigs(userID)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to get offline download configs: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code: 200,
		Data: map[string]interface{}{
			"configs": configs,
			"user_id": userID,
		},
	})
}

// 配置 Aria2 下载器
func handleSetAria2Config(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)

	var req struct {
		URI    string `json:"uri" form:"uri"`
		Secret string `json:"secret" form:"secret"`
	}

	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	config := OfflineDownloadConfig{
		ToolName: "aria2",
		Config:   fmt.Sprintf(`{"uri": "%s", "secret": "%s"}`, req.URI, req.Secret),
		Enabled:  true,
	}

	if err := createUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
		// 如果已存在则更新
		if err := updateUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to configure Aria2: " + err.Error(),
			})
			return
		}
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Aria2 configured successfully",
		Data: map[string]interface{}{
			"tool": "aria2",
			"uri":  req.URI,
		},
	})
}

// 配置 qBittorrent 下载器
func handleSetQbittorrentConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)

	var req struct {
		URL      string `json:"url" form:"url"`
		Seedtime string `json:"seedtime" form:"seedtime"`
	}

	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	config := OfflineDownloadConfig{
		ToolName: "qbittorrent",
		Config:   fmt.Sprintf(`{"url": "%s", "seedtime": "%s"}`, req.URL, req.Seedtime),
		Enabled:  true,
	}

	if err := createUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
		// 如果已存在则更新
		if err := updateUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to configure qBittorrent: " + err.Error(),
			})
			return
		}
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "qBittorrent configured successfully",
	})
}

// 配置 Transmission 下载器
func handleSetTransmissionConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)

	var req struct {
		URI      string `json:"uri" form:"uri"`
		Seedtime string `json:"seedtime" form:"seedtime"`
	}

	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	config := OfflineDownloadConfig{
		ToolName: "transmission",
		Config:   fmt.Sprintf(`{"uri": "%s", "seedtime": "%s"}`, req.URI, req.Seedtime),
		Enabled:  true,
	}

	if err := createUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
		// 如果已存在则更新
		if err := updateUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to configure Transmission: " + err.Error(),
			})
			return
		}
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Transmission configured successfully",
	})
}

// 配置 115 云盘离线下载
func handleSet115Config(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)

	var req struct {
		TempDirPath string `json:"temp_dir_path" form:"temp_dir_path"`
		ConfigID    uint   `json:"config_id" form:"config_id"`
	}

	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// 验证驱动配置
	if req.ConfigID > 0 {
		if err := validateDriverForOfflineDownload(userID, req.ConfigID, "115"); err != nil {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Driver validation failed: " + err.Error(),
			})
			return
		}
	}

	config := OfflineDownloadConfig{
		ToolName:    "115",
		Config:      `{}`,
		TempDirPath: req.TempDirPath,
		Enabled:     true,
	}

	if err := createUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
		// 如果已存在则更新
		if err := updateUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to configure 115: " + err.Error(),
			})
			return
		}
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "115 Cloud configured successfully",
	})
}

// 配置 PikPak 离线下载
func handleSetPikPakConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)

	var req struct {
		TempDirPath string `json:"temp_dir_path" form:"temp_dir_path"`
		ConfigID    uint   `json:"config_id" form:"config_id"`
	}

	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// 验证驱动配置
	if req.ConfigID > 0 {
		if err := validateDriverForOfflineDownload(userID, req.ConfigID, "pikpak"); err != nil {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Driver validation failed: " + err.Error(),
			})
			return
		}
	}

	config := OfflineDownloadConfig{
		ToolName:    "pikpak",
		Config:      `{}`,
		TempDirPath: req.TempDirPath,
		Enabled:     true,
	}

	if err := createUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
		// 如果已存在则更新
		if err := updateUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to configure PikPak: " + err.Error(),
			})
			return
		}
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "PikPak configured successfully",
	})
}

// 配置迅雷离线下载
func handleSetThunderConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)

	var req struct {
		TempDirPath string `json:"temp_dir_path" form:"temp_dir_path"`
		ConfigID    uint   `json:"config_id" form:"config_id"`
	}

	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// 验证驱动配置
	if req.ConfigID > 0 {
		if err := validateDriverForOfflineDownload(userID, req.ConfigID, "thunder"); err != nil {
			respondJSON(w, APIResponse{
				Code:    400,
				Message: "Driver validation failed: " + err.Error(),
			})
			return
		}
	}

	config := OfflineDownloadConfig{
		ToolName:    "thunder",
		Config:      `{}`,
		TempDirPath: req.TempDirPath,
		Enabled:     true,
	}

	if err := createUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
		// 如果已存在则更新
		if err := updateUserOfflineDownloadConfig(r.Context(), userID, config); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to configure Thunder: " + err.Error(),
			})
			return
		}
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Thunder configured successfully",
	})
}

// 离线下载任务 API 处理器

// 获取用户的离线下载任务列表
func handleUserOfflineDownloadTasks(w http.ResponseWriter, r *http.Request) {
	userID := getCurrentUserID(r)

	page := 1
	perPage := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if p, err := strconv.Atoi(perPageStr); err == nil && p > 0 {
			perPage = p
		}
	}

	tasks, total, err := getUserOfflineDownloadTasks(userID, page, perPage)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to get offline download tasks: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code: 200,
		Data: map[string]interface{}{
			"tasks":    tasks,
			"total":    total,
			"page":     page,
			"per_page": perPage,
			"user_id":  userID,
		},
	})
}

// 添加离线下载任务
func handleAddOfflineDownloadTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)

	var req struct {
		URLs         []string `json:"urls"`
		ConfigID     uint     `json:"config_id"`
		DstPath      string   `json:"dst_path"`
		Tool         string   `json:"tool"`
		DeletePolicy string   `json:"delete_policy"`
	}

	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	if len(req.URLs) == 0 {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "URLs are required",
		})
		return
	}

	if req.ConfigID == 0 {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "config_id is required",
		})
		return
	}

	// 验证驱动配置
	if err := validateDriverForOfflineDownload(userID, req.ConfigID, req.Tool); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Driver validation failed: " + err.Error(),
		})
		return
	}

	// 序列化 URLs
	urlsJSON, err := json.Marshal(req.URLs)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to serialize URLs: " + err.Error(),
		})
		return
	}

	task := OfflineDownloadTask{
		ConfigID:     req.ConfigID,
		URLs:         string(urlsJSON),
		DstPath:      req.DstPath,
		Tool:         req.Tool,
		DeletePolicy: req.DeletePolicy,
	}

	if err := createOfflineDownloadTask(r.Context(), userID, task); err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to create offline download task: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Offline download task created successfully",
		Data: map[string]interface{}{
			"urls_count": len(req.URLs),
			"tool":       req.Tool,
			"dst_path":   req.DstPath,
		},
	})
}

// 更新离线下载任务状态
func handleUpdateOfflineDownloadTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)

	var req struct {
		TaskID   uint   `json:"task_id" form:"task_id"`
		Status   string `json:"status" form:"status"`
		Progress int    `json:"progress" form:"progress"`
		Error    string `json:"error" form:"error"`
	}

	if err := parseJSON(r, &req); err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	if req.TaskID == 0 {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "task_id is required",
		})
		return
	}

	if err := updateOfflineDownloadTaskStatus(r.Context(), userID, req.TaskID, req.Status, req.Progress, req.Error); err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to update task status: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Task status updated successfully",
	})
}

// 删除离线下载任务
func handleDeleteOfflineDownloadTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondJSON(w, APIResponse{
			Code:    405,
			Message: "Method not allowed",
		})
		return
	}

	userID := getCurrentUserID(r)
	taskIDStr := r.URL.Query().Get("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		respondJSON(w, APIResponse{
			Code:    400,
			Message: "Invalid task_id",
		})
		return
	}

	if err := deleteOfflineDownloadTask(r.Context(), userID, uint(taskID)); err != nil {
		respondJSON(w, APIResponse{
			Code:    500,
			Message: "Failed to delete task: " + err.Error(),
		})
		return
	}

	respondJSON(w, APIResponse{
		Code:    200,
		Message: "Task deleted successfully",
	})
}

// JWT 辅助函数

// 生成 JWT Token
func generateJWTToken(user *model.User) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Exp:      now.Add(JWT_EXPIRATION).Unix(),
		Iat:      now.Unix(),
	}

	// 创建 JWT header
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	// 编码 header 和 payload
	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(claims)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// 创建签名
	message := headerEncoded + "." + payloadEncoded
	signature := createHMACSignature(message, JWT_SECRET)

	// 组装 JWT
	token := message + "." + signature
	return token, nil
}

// 创建 HMAC 签名
func createHMACSignature(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// 验证 JWT Token
func verifyJWTToken(tokenString string) (*JWTClaims, error) {
	// 分割 token
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded, payloadEncoded, signatureEncoded := parts[0], parts[1], parts[2]

	// 验证签名
	message := headerEncoded + "." + payloadEncoded
	expectedSignature := createHMACSignature(message, JWT_SECRET)

	if expectedSignature != signatureEncoded {
		return nil, fmt.Errorf("invalid token signature")
	}

	// 解码 payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// 检查过期时间
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// 从请求头中提取 JWT Token
func extractTokenFromRequest(r *http.Request) string {
	// 从 Authorization header 中提取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Bearer <token> 格式
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		// 直接是 token
		return authHeader
	}

	// 从查询参数中提取
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	return ""
}

// 认证中间件 - 验证用户登录状态
func authenticateUser(r *http.Request) (*JWTClaims, error) {
	token := extractTokenFromRequest(r)
	if token == "" {
		return nil, fmt.Errorf("no token provided")
	}

	claims, err := verifyJWTToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}

// 检查用户权限
func checkUserPermission(claims *JWTClaims, requiredRole int) bool {
	// ADMIN 角色可以访问所有内容
	if claims.Role == int(model.ADMIN) {
		return true
	}

	// 检查是否满足最低角色要求
	return claims.Role >= requiredRole
}

// 获取当前认证用户信息
func getCurrentAuthenticatedUser(r *http.Request) (*JWTClaims, error) {
	return authenticateUser(r)
}

func main() {
	// 初始化驱动系统
	drivers.All()

	// 设置路由

	// 用户认证相关路由（无需认证）
	http.HandleFunc("/api/auth/register", handleUserRegister)
	http.HandleFunc("/api/auth/login", handleUserLogin)
	http.HandleFunc("/api/auth/logout", handleUserLogout)

	// 获取当前用户信息（需要认证）
	http.HandleFunc("/api/auth/me", requireAuth(handleCurrentUser))

	// 用户驱动配置管理路由（需要认证）
	http.HandleFunc("/api/drivers", requireAuth(handleUserDriversAPI))
	http.HandleFunc("/api/user/driver/list", requireAuth(handleUserDriversAPI))
	http.HandleFunc("/api/user/driver/get", requireAuth(handleUserDriverAPI))
	http.HandleFunc("/api/user/driver/create", requireAuth(handleUserDriversAPI))
	http.HandleFunc("/api/user/driver/update", requireAuth(handleUserDriverAPI))
	http.HandleFunc("/api/user/driver/delete", requireAuth(handleDeleteUserDriverAPI))
	http.HandleFunc("/api/user/driver/enable", requireAuth(handleEnableUserDriverAPI))
	http.HandleFunc("/api/user/driver/disable", requireAuth(handleDisableUserDriverAPI))

	// 用户管理路由（需要管理员权限）
	http.HandleFunc("/api/admin/user/list", requireAdmin(handleUsersAPI))
	http.HandleFunc("/api/admin/user/get", requireAdmin(handleUserAPI))
	http.HandleFunc("/api/admin/user/create", requireAdmin(handleUsersAPI))
	http.HandleFunc("/api/admin/user/update", requireAdmin(handleUserAPI))
	http.HandleFunc("/api/admin/user/delete", requireAdmin(handleDeleteUserAPI))

	// 基于用户驱动配置的文件系统路由（需要认证）
	http.HandleFunc("/api/fs/list", requireAuth(handleFileSystemList))
	http.HandleFunc("/api/fs/get", requireAuth(handleFileSystemGet))
	http.HandleFunc("/api/fs/dirs", requireAuth(handleFileSystemDirs))
	http.HandleFunc("/api/fs/mkdir", requireAuth(handleFileSystemMkdir))
	http.HandleFunc("/api/fs/rename", requireAuth(handleFileSystemRename))
	http.HandleFunc("/api/fs/move", requireAuth(handleFileSystemMove))
	http.HandleFunc("/api/fs/copy", requireAuth(handleFileSystemCopy))
	http.HandleFunc("/api/fs/remove", requireAuth(handleFileSystemRemove))
	http.HandleFunc("/api/fs/upload", requireAuth(handleFileSystemUpload))

	// 文件下载路由（需要认证）
	http.HandleFunc("/d/", requireAuth(handleFileSystemDownload))

	// 离线下载相关路由（需要认证）
	http.HandleFunc("/api/offline_download_tools", requireAuth(handleOfflineDownloadTools))
	http.HandleFunc("/api/user/offline_download/configs", requireAuth(handleUserOfflineDownloadConfigs))
	http.HandleFunc("/api/user/offline_download/tasks", requireAuth(handleUserOfflineDownloadTasks))
	http.HandleFunc("/api/user/offline_download/add_task", requireAuth(handleAddOfflineDownloadTask))
	http.HandleFunc("/api/user/offline_download/update_task", requireAuth(handleUpdateOfflineDownloadTaskStatus))
	http.HandleFunc("/api/user/offline_download/delete_task", requireAuth(handleDeleteOfflineDownloadTask))

	// 离线下载工具配置路由（需要认证）
	http.HandleFunc("/api/admin/setting/set_aria2", requireAuth(handleSetAria2Config))
	http.HandleFunc("/api/admin/setting/set_qbittorrent", requireAuth(handleSetQbittorrentConfig))
	http.HandleFunc("/api/admin/setting/set_transmission", requireAuth(handleSetTransmissionConfig))
	http.HandleFunc("/api/admin/setting/set_115", requireAuth(handleSet115Config))
	http.HandleFunc("/api/admin/setting/set_pikpak", requireAuth(handleSetPikPakConfig))
	http.HandleFunc("/api/admin/setting/set_thunder", requireAuth(handleSetThunderConfig))

	// 健康检查（无需认证）
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// 统计用户和驱动配置数量
		totalDriverConfigs := 0
		enabledDriverConfigs := 0
		for _, driver := range driversMap {
			totalDriverConfigs++
			if driver.Enabled {
				enabledDriverConfigs++
			}
		}

		// 统计离线下载配置和任务数量
		totalOfflineConfigs := len(offlineDownloadConfigs)
		totalOfflineTasks := len(offlineDownloadTasks)

		respondJSON(w, APIResponse{
			Code:    200,
			Message: "OpenList Workers is running",
			Data: map[string]interface{}{
				"users_count":              len(usersMap),
				"total_driver_configs":     totalDriverConfigs,
				"enabled_driver_configs":   enabledDriverConfigs,
				"driver_instances":         len(userDriverInstances),
				"offline_download_configs": totalOfflineConfigs,
				"offline_download_tasks":   totalOfflineTasks,
				"supported_offline_tools":  getSupportedOfflineDownloadTools(),
				"timestamp":                time.Now().Unix(),
				"version":                  "workers-1.0.0-auth",
				"auth_enabled":             true,
			},
		})
	})

	// 初始化端点（无需认证）
	http.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
		// 初始化 D1 数据库
		if err := initD1DatabaseWithManager("openlist-db"); err != nil {
			respondJSON(w, APIResponse{
				Code:    500,
				Message: "Failed to initialize database: " + err.Error(),
			})
			return
		}

		respondJSON(w, APIResponse{
			Code:    200,
			Message: "System initialized successfully",
			Data: map[string]interface{}{
				"users_count":              len(usersMap),
				"driver_configs_count":     len(driversMap),
				"database_tables":          4, // users, driver_configs, offline_download_configs, offline_download_tasks
				"filesystem_enabled":       true,
				"offline_download_enabled": true,
				"auth_enabled":             true,
				"supported_tools":          getSupportedOfflineDownloadTools(),
				"message":                  "Please register or login to use the system",
			},
		})
	})

	// 启动 Workers
	workers.Serve(nil)
}
