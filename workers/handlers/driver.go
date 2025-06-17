package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList-workers/workers/drivers"
	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

// DriverHandler 驱动处理器
type DriverHandler struct {
	driverService *drivers.DriverService
}

// NewDriverHandler 创建驱动处理器
func NewDriverHandler(driverService *drivers.DriverService) *DriverHandler {
	return &DriverHandler{
		driverService: driverService,
	}
}

// ListDriverInfo 列出所有驱动信息
func (h *DriverHandler) ListDriverInfo(w http.ResponseWriter, r *http.Request) {
	driverInfos := drivers.GetDriverInfos()
	
	response := APIResponse{
		Code:    200,
		Message: "success",
		Data:    driverInfos,
	}
	
	writeJSONResponse(w, response)
}

// ListDriverNames 列出所有驱动名称
func (h *DriverHandler) ListDriverNames(w http.ResponseWriter, r *http.Request) {
	driverNames := drivers.GetDriverNames()
	
	response := APIResponse{
		Code:    200,
		Message: "success",
		Data:    driverNames,
	}
	
	writeJSONResponse(w, response)
}

// GetDriverInfo 获取指定驱动信息
func (h *DriverHandler) GetDriverInfo(w http.ResponseWriter, r *http.Request) {
	driverName := r.URL.Query().Get("driver")
	if driverName == "" {
		writeErrorResponse(w, 400, "driver parameter is required")
		return
	}
	
	driverInfo, err := drivers.GetDriverInfo(driverName)
	if err != nil {
		writeErrorResponse(w, 404, fmt.Sprintf("driver %s not found", driverName))
		return
	}
	
	response := APIResponse{
		Code:    200,
		Message: "success",
		Data:    driverInfo,
	}
	
	writeJSONResponse(w, response)
}

// ListFiles 列出文件
func (h *DriverHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
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
	
	// 调用驱动服务
	files, err := h.driverService.ListFiles(ctx, path, args)
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
		},
	}
	
	writeJSONResponse(w, response)
}

// GetFileLink 获取文件链接
func (h *DriverHandler) GetFileLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		writeErrorResponse(w, 400, "path parameter is required")
		return
	}
	
	// 构建链接参数
	args := drivers.LinkArgs{
		IP:      getClientIP(r),
		Header:  r.Header,
		Type:    r.URL.Query().Get("type"),
		HttpReq: r,
	}
	
	// 调用驱动服务
	link, err := h.driverService.GetFileLink(ctx, path, args)
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

// GetFile 获取文件信息
func (h *DriverHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		writeErrorResponse(w, 400, "path parameter is required")
		return
	}
	
	// 调用驱动服务
	file, err := h.driverService.GetFile(ctx, path)
	if err != nil {
		writeErrorResponse(w, 500, fmt.Sprintf("failed to get file: %v", err))
		return
	}
	
	// 构建响应
	fileInfo := FileInfo{
		Name:     file.GetName(),
		Size:     file.GetSize(),
		IsDir:    file.IsDir(),
		Modified: file.ModTime().Unix(),
		Path:     file.GetPath(),
	}
	
	response := APIResponse{
		Code:    200,
		Message: "success",
		Data:    fileInfo,
	}
	
	writeJSONResponse(w, response)
}

// DownloadFile 下载文件
func (h *DriverHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		writeErrorResponse(w, 400, "path parameter is required")
		return
	}
	
	// 构建链接参数
	args := drivers.LinkArgs{
		IP:      getClientIP(r),
		Header:  r.Header,
		HttpReq: r,
	}
	
	// 获取文件链接
	link, err := h.driverService.GetFileLink(ctx, path, args)
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

// FileInfo 文件信息响应
type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"is_dir"`
	Modified int64  `json:"modified"`
	Path     string `json:"path"`
}

// 辅助函数
func getClientIP(r *http.Request) string {
	// 尝试从各种头部获取真实IP
	ip := r.Header.Get("CF-Connecting-IP")
	if ip != "" {
		return ip
	}
	
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For可能包含多个IP，取第一个
		if comma := strings.Index(ip, ","); comma != -1 {
			ip = ip[:comma]
		}
		return strings.TrimSpace(ip)
	}
	
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	
	// fallback到RemoteAddr
	if colon := strings.LastIndex(r.RemoteAddr, ":"); colon != -1 {
		return r.RemoteAddr[:colon]
	}
	
	return r.RemoteAddr
}

func getFileName(path string) string {
	if slash := strings.LastIndex(path, "/"); slash != -1 {
		return path[slash+1:]
	}
	return path
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSONResponse(w http.ResponseWriter, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	json.NewEncoder(w).Encode(response)
}

func writeErrorResponse(w http.ResponseWriter, code int, message string) {
	response := APIResponse{
		Code:    code,
		Message: message,
	}
	writeJSONResponse(w, response)
} 