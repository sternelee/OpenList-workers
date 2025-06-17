package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/OpenListTeam/OpenList-workers/workers/db"
	"github.com/OpenListTeam/OpenList-workers/workers/drivers"
	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

// StorageHandler 存储管理处理器
type StorageHandler struct {
	repos         *db.Repositories
	driverService *drivers.DriverService
}

// NewStorageHandler 创建存储处理器
func NewStorageHandler(repos *db.Repositories, driverService *drivers.DriverService) *StorageHandler {
	return &StorageHandler{
		repos:         repos,
		driverService: driverService,
	}
}

// ListStorages 列出所有存储
func (h *StorageHandler) ListStorages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
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
	
	storages, err := h.repos.Storage.List(ctx, perPage, offset)
	if err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to list storages: %v", err))
		return
	}
	
	response := APIResponse{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"content": storages,
			"total":   len(storages),
			"page":    page,
			"per_page": perPage,
		},
	}
	
	writeJSONResponse(w, response)
}

// GetStorage 获取单个存储
func (h *StorageHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 获取存储ID
	storageID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeErrorResponse(w, 400, "invalid storage id")
		return
	}
	
	storage, err := h.repos.Storage.GetByID(ctx, storageID)
	if err != nil {
		writeErrorResponse(w, 404, "storage not found")
		return
	}
	
	response := APIResponse{
		Code:    200,
		Message: "success",
		Data:    storage,
	}
	
	writeJSONResponse(w, response)
}

// CreateStorage 创建存储
func (h *StorageHandler) CreateStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var storage models.Storage
	if err := json.NewDecoder(r.Body).Decode(&storage); err != nil {
		writeErrorResponse(w, 400, "invalid request body")
		return
	}
	
	// 设置默认值
	storage.Modified = time.Now()
	storage.Status = "pending"
	
	// 创建存储记录
	if err := h.repos.Storage.Create(ctx, &storage); err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to create storage: %v", err))
		return
	}
	
	// 尝试加载驱动
	go func() {
		bgCtx := context.Background()
		if err := h.driverService.Initialize(bgCtx); err != nil {
			fmt.Printf("Failed to reload drivers after storage creation: %v\n", err)
		}
	}()
	
	response := APIResponse{
		Code:    201,
		Message: "storage created successfully",
		Data:    storage,
	}
	
	writeJSONResponse(w, response)
}

// UpdateStorage 更新存储
func (h *StorageHandler) UpdateStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 获取存储ID
	storageID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeErrorResponse(w, 400, "invalid storage id")
		return
	}
	
	var storage models.Storage
	if err := json.NewDecoder(r.Body).Decode(&storage); err != nil {
		writeErrorResponse(w, 400, "invalid request body")
		return
	}
	
	// 设置ID
	storage.ID = storageID
	storage.Modified = time.Now()
	
	// 更新存储记录
	if err := h.repos.Storage.Update(ctx, &storage); err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to update storage: %v", err))
		return
	}
	
	// 重新加载驱动
	go func() {
		bgCtx := context.Background()
		if err := h.driverService.Initialize(bgCtx); err != nil {
			fmt.Printf("Failed to reload drivers after storage update: %v\n", err)
		}
	}()
	
	response := APIResponse{
		Code:    200,
		Message: "storage updated successfully",
		Data:    storage,
	}
	
	writeJSONResponse(w, response)
}

// DeleteStorage 删除存储
func (h *StorageHandler) DeleteStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 获取存储ID
	storageID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeErrorResponse(w, 400, "invalid storage id")
		return
	}
	
	// 删除存储记录
	if err := h.repos.Storage.Delete(ctx, storageID); err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to delete storage: %v", err))
		return
	}
	
	// 重新加载驱动
	go func() {
		bgCtx := context.Background()
		if err := h.driverService.Initialize(bgCtx); err != nil {
			fmt.Printf("Failed to reload drivers after storage deletion: %v\n", err)
		}
	}()
	
	response := APIResponse{
		Code:    200,
		Message: "storage deleted successfully",
	}
	
	writeJSONResponse(w, response)
}

// EnableStorage 启用存储
func (h *StorageHandler) EnableStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 获取存储ID
	storageID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeErrorResponse(w, 400, "invalid storage id")
		return
	}
	
	// 获取存储记录
	storage, err := h.repos.Storage.GetByID(ctx, storageID)
	if err != nil {
		writeErrorResponse(w, 404, "storage not found")
		return
	}
	
	// 启用存储
	storage.Disabled = false
	storage.Modified = time.Now()
	
	if err := h.repos.Storage.Update(ctx, storage); err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to enable storage: %v", err))
		return
	}
	
	// 重新加载驱动
	go func() {
		bgCtx := context.Background()
		if err := h.driverService.Initialize(bgCtx); err != nil {
			fmt.Printf("Failed to reload drivers after enabling storage: %v\n", err)
		}
	}()
	
	response := APIResponse{
		Code:    200,
		Message: "storage enabled successfully",
		Data:    storage,
	}
	
	writeJSONResponse(w, response)
}

// DisableStorage 禁用存储
func (h *StorageHandler) DisableStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 获取存储ID
	storageID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeErrorResponse(w, 400, "invalid storage id")
		return
	}
	
	// 获取存储记录
	storage, err := h.repos.Storage.GetByID(ctx, storageID)
	if err != nil {
		writeErrorResponse(w, 404, "storage not found")
		return
	}
	
	// 禁用存储
	storage.Disabled = true
	storage.Modified = time.Now()
	
	if err := h.repos.Storage.Update(ctx, storage); err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to disable storage: %v", err))
		return
	}
	
	// 重新加载驱动
	go func() {
		bgCtx := context.Background()
		if err := h.driverService.Initialize(bgCtx); err != nil {
			fmt.Printf("Failed to reload drivers after disabling storage: %v\n", err)
		}
	}()
	
	response := APIResponse{
		Code:    200,
		Message: "storage disabled successfully",
		Data:    storage,
	}
	
	writeJSONResponse(w, response)
}

// TestStorage 测试存储连接
func (h *StorageHandler) TestStorage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var storage models.Storage
	if err := json.NewDecoder(r.Body).Decode(&storage); err != nil {
		writeErrorResponse(w, 400, "invalid request body")
		return
	}
	
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
		},
	}
	
	writeJSONResponse(w, response)
} 