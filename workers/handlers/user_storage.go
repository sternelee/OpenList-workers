package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sternelee/OpenList-workers/workers/db"
	"github.com/sternelee/OpenList-workers/workers/drivers"
	"github.com/sternelee/OpenList-workers/workers/models"
)

// UserStorageHandler 用户存储处理器
type UserStorageHandler struct {
	repos             *db.Repositories
	userDriverService *drivers.UserDriverService
}

// NewUserStorageHandler 创建用户存储处理器
func NewUserStorageHandler(repos *db.Repositories, userDriverService *drivers.UserDriverService) *UserStorageHandler {
	return &UserStorageHandler{
		repos:             repos,
		userDriverService: userDriverService,
	}
}

// ListUserStorages 列出用户的存储配置
func (h *UserStorageHandler) ListUserStorages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)
	if user == nil {
		writeErrorResponse(w, 401, "user not found")
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	storages, err := h.repos.Storage.ListByUser(ctx, user.ID, perPage, offset)
	if err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to list storages: %v", err))
		return
	}

	response := APIResponse{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"content":  storages,
			"total":    len(storages),
			"page":     page,
			"per_page": perPage,
		},
	}

	writeJSONResponse(w, response)
}

// CreateUserStorage 创建用户存储
func (h *UserStorageHandler) CreateUserStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)
	if user == nil {
		writeErrorResponse(w, 401, "user not found")
		return
	}

	var storage models.Storage
	if err := json.NewDecoder(r.Body).Decode(&storage); err != nil {
		writeErrorResponse(w, 400, "invalid request body")
		return
	}

	// 设置用户ID和默认值
	storage.UserID = user.ID
	storage.Modified = time.Now()
	storage.Status = "pending"

	// 检查挂载路径是否重复
	if _, err := h.repos.Storage.GetByUserAndPath(ctx, user.ID, storage.MountPath); err == nil {
		writeErrorResponse(w, 400, "mount path already exists")
		return
	}

	// 创建存储记录
	if err := h.repos.Storage.Create(ctx, &storage); err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to create storage: %v", err))
		return
	}

	// 尝试初始化用户驱动
	go func() {
		bgCtx := context.Background()
		if err := h.userDriverService.InitializeUser(bgCtx, user.ID); err != nil {
			fmt.Printf("Failed to reload user %d drivers: %v\n", user.ID, err)
		}
	}()

	response := APIResponse{
		Code:    201,
		Message: "storage created successfully",
		Data:    storage,
	}

	writeJSONResponse(w, response)
}

// UpdateUserStorage 更新用户存储
func (h *UserStorageHandler) UpdateUserStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)
	if user == nil {
		writeErrorResponse(w, 401, "user not found")
		return
	}

	// 获取存储ID
	storageID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeErrorResponse(w, 400, "invalid storage id")
		return
	}

	// 检查存储是否属于当前用户
	existingStorage, err := h.repos.Storage.GetByID(ctx, storageID)
	if err != nil {
		writeErrorResponse(w, 404, "storage not found")
		return
	}

	if existingStorage.UserID != user.ID && !user.IsAdmin() {
		writeErrorResponse(w, 403, "access denied")
		return
	}

	var storage models.Storage
	if err := json.NewDecoder(r.Body).Decode(&storage); err != nil {
		writeErrorResponse(w, 400, "invalid request body")
		return
	}

	// 设置ID和用户ID
	storage.ID = storageID
	storage.UserID = user.ID
	storage.Modified = time.Now()

	// 更新存储记录
	if err := h.repos.Storage.Update(ctx, &storage); err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to update storage: %v", err))
		return
	}

	// 重新加载用户驱动
	go func() {
		bgCtx := context.Background()
		if err := h.userDriverService.InitializeUser(bgCtx, user.ID); err != nil {
			fmt.Printf("Failed to reload user %d drivers: %v\n", user.ID, err)
		}
	}()

	response := APIResponse{
		Code:    200,
		Message: "storage updated successfully",
		Data:    storage,
	}

	writeJSONResponse(w, response)
}

// DeleteUserStorage 删除用户存储
func (h *UserStorageHandler) DeleteUserStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)
	if user == nil {
		writeErrorResponse(w, 401, "user not found")
		return
	}

	// 获取存储ID
	storageID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeErrorResponse(w, 400, "invalid storage id")
		return
	}

	// 检查存储是否属于当前用户
	existingStorage, err := h.repos.Storage.GetByID(ctx, storageID)
	if err != nil {
		writeErrorResponse(w, 404, "storage not found")
		return
	}

	if existingStorage.UserID != user.ID && !user.IsAdmin() {
		writeErrorResponse(w, 403, "access denied")
		return
	}

	// 删除存储记录
	if err := h.repos.Storage.Delete(ctx, storageID); err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to delete storage: %v", err))
		return
	}

	// 重新加载用户驱动
	go func() {
		bgCtx := context.Background()
		if err := h.userDriverService.InitializeUser(bgCtx, user.ID); err != nil {
			fmt.Printf("Failed to reload user %d drivers: %v\n", user.ID, err)
		}
	}()

	response := APIResponse{
		Code:    200,
		Message: "storage deleted successfully",
	}

	writeJSONResponse(w, response)
}

// TestUserStorage 测试用户存储连接
func (h *UserStorageHandler) TestUserStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)
	if user == nil {
		writeErrorResponse(w, 401, "user not found")
		return
	}

	var storage models.Storage
	if err := json.NewDecoder(r.Body).Decode(&storage); err != nil {
		writeErrorResponse(w, 400, "invalid request body")
		return
	}

	// 设置用户ID
	storage.UserID = user.ID

	// 创建临时驱动实例进行测试
	driver, err := drivers.CreateDriver(&storage)
	if err != nil {
		writeErrorResponse(w, 400, fmt.Sprintf("failed to create driver: %v", err))
		return
	}

	// 测试初始化
	if err := drivers.InitializeDriver(ctx, driver); err != nil {
		writeErrorResponse(w, 400, fmt.Sprintf("driver initialization failed: %v", err))
		return
	}

	// 清理资源
	defer driver.Drop(ctx)

	// 尝试列出根目录
	rootObj := &drivers.Object{
		Path:     "/",
		Name:     "/",
		IsFolder: true,
	}

	_, err = driver.List(ctx, rootObj, drivers.ListArgs{})
	if err != nil {
		writeErrorResponse(w, 400, fmt.Sprintf("failed to list root directory: %v", err))
		return
	}

	response := APIResponse{
		Code:    200,
		Message: "storage test successful",
		Data: map[string]interface{}{
			"status": "success",
			"driver": storage.Driver,
			"user":   user.Username,
		},
	}

	writeJSONResponse(w, response)
}

// UserFileHandler 用户文件处理器
type UserFileHandler struct {
	userDriverService *drivers.UserDriverService
}

// NewUserFileHandler 创建用户文件处理器
func NewUserFileHandler(userDriverService *drivers.UserDriverService) *UserFileHandler {
	return &UserFileHandler{
		userDriverService: userDriverService,
	}
}

// ListUserFiles 列出用户文件
func (h *UserFileHandler) ListUserFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)
	if user == nil {
		writeErrorResponse(w, 401, "user not found")
		return
	}

	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	// 获取其他参数
	refresh := r.URL.Query().Get("refresh") == "true"

	// 构建列表参数
	args := drivers.ListArgs{
		ReqPath: path,
		Refresh: refresh,
	}

	// 调用用户驱动服务
	files, err := h.userDriverService.ListUserFiles(ctx, user.ID, path, args)
	if err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to list files: %v", err))
		return
	}

	// 转换为响应格式
	var fileInfos []FileInfo
	for _, file := range files {
		fileInfos = append(fileInfos, FileInfo{
			Name:     file.GetName(),
			Size:     file.GetSize(),
			IsDir:    file.IsDir(),
			Modified: file.ModTime().Unix(),
			Path:     file.GetPath(),
		})
	}

	response := APIResponse{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"content": fileInfos,
			"total":   len(fileInfos),
			"user":    user.Username,
		},
	}

	writeJSONResponse(w, response)
}

// GetUserFileLink 获取用户文件链接
func (h *UserFileHandler) GetUserFileLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)

	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		writeErrorResponse(w, 400, "path parameter is required")
		return
	}

	// 如果没有用户信息，检查是否为公开资源
	var userID int
	if user != nil {
		userID = user.ID
	} else {
		userID = 0 // 匿名用户
	}

	// 构建链接参数
	args := drivers.LinkArgs{
		IP:      getClientIP(r),
		Header:  r.Header,
		Type:    r.URL.Query().Get("type"),
		HttpReq: r,
	}

	// 调用用户驱动服务
	link, err := h.userDriverService.GetUserFileLink(ctx, userID, path, args)
	if err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to get file link: %v", err))
		return
	}

	// 如果有直接URL，重定向
	if link.URL != "" {
		http.Redirect(w, r, link.URL, http.StatusFound)
		return
	}

	// 如果有文件流，直接返回
	if link.MFile != nil {
		defer link.MFile.Close()

		// 设置响应头
		for key, values := range link.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// 复制文件内容
		http.ServeContent(w, r, "", time.Time{}, link.MFile)
		return
	}

	writeErrorResponse(w, 404, "file not found")
}

// DownloadUserFile 下载用户文件
func (h *UserFileHandler) DownloadUserFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetUserFromContext(ctx)

	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		writeErrorResponse(w, 400, "path parameter is required")
		return
	}

	// 如果没有用户信息，检查是否为公开资源
	var userID int
	if user != nil {
		userID = user.ID
	} else {
		userID = 0 // 匿名用户
	}

	// 构建链接参数
	args := drivers.LinkArgs{
		IP:      getClientIP(r),
		Header:  r.Header,
		HttpReq: r,
	}

	// 获取文件链接
	link, err := h.userDriverService.GetUserFileLink(ctx, userID, path, args)
	if err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to get download link: %v", err))
		return
	}

	// 处理下载
	if link.URL != "" {
		// 重定向到下载URL
		http.Redirect(w, r, link.URL, http.StatusFound)
		return
	}

	if link.MFile != nil {
		defer link.MFile.Close()

		// 设置下载响应头
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", getFileName(path)))

		// 设置其他响应头
		for key, values := range link.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// 复制文件内容
		http.ServeContent(w, r, getFileName(path), time.Time{}, link.MFile)
		return
	}

	writeErrorResponse(w, 404, "file not found")
}

