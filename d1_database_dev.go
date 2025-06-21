//go:build !workers
// +build !workers

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenListTeam/OpenList/internal/model"
)

// D1DatabaseManager 开发环境的模拟数据库管理器
type D1DatabaseManager struct {
	dbName string
}

// NewD1DatabaseManager 创建新的 D1 数据库管理器
func NewD1DatabaseManager(dbName string) (*D1DatabaseManager, error) {
	return &D1DatabaseManager{
		dbName: dbName,
	}, nil
}

// CreateTables 创建数据库表（开发环境模拟）
func (d *D1DatabaseManager) CreateTables(ctx context.Context) error {
	fmt.Printf("Creating tables for database: %s (dev mode)\n", d.dbName)
	fmt.Println("✓ Created users table")
	fmt.Println("✓ Created driver_configs table")
	fmt.Println("✓ Created offline_download_configs table")
	fmt.Println("✓ Created offline_download_tasks table")
	return nil
}

// User management methods (dev mode)
func (d *D1DatabaseManager) GetUsers(ctx context.Context, page, perPage int) ([]model.User, int64, error) {
	fmt.Printf("Getting users (page %d, per_page %d) - dev mode\n", page, perPage)
	// 模拟用户数据
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
			Username:   "testuser",
			BasePath:   "/",
			Role:       model.GENERAL,
			Disabled:   false,
			Permission: 0,
		},
	}

	// 简单分页
	start := (page - 1) * perPage
	end := start + perPage
	total := int64(len(users))

	if start > len(users) {
		return []model.User{}, total, nil
	}
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], total, nil
}

func (d *D1DatabaseManager) CreateUser(ctx context.Context, user model.User) error {
	fmt.Printf("Creating user: %s - dev mode\n", user.Username)
	user.ID = uint(time.Now().Unix())
	if user.Authn == "" {
		user.Authn = "[]"
	}
	usersMap[user.ID] = &user
	return nil
}

func (d *D1DatabaseManager) UpdateUser(ctx context.Context, user model.User) error {
	fmt.Printf("Updating user: %s - dev mode\n", user.Username)
	usersMap[user.ID] = &user
	return nil
}

func (d *D1DatabaseManager) DeleteUser(ctx context.Context, userID uint) error {
	fmt.Printf("Deleting user: %d - dev mode\n", userID)
	delete(usersMap, userID)
	return nil
}

// Driver config management methods (dev mode)
func (d *D1DatabaseManager) GetUserDriverConfigs(ctx context.Context, userID uint, page, perPage int) ([]DriverConfig, int64, error) {
	fmt.Printf("Getting driver configs for user %d (page %d, per_page %d) - dev mode\n", userID, page, perPage)

	// 返回用户的驱动配置
	var configs []DriverConfig
	for _, config := range driversMap {
		if config.UserID == userID {
			configs = append(configs, *config)
		}
	}

	// 简单排序
	for i := 0; i < len(configs)-1; i++ {
		for j := i + 1; j < len(configs); j++ {
			if configs[i].Order > configs[j].Order {
				configs[i], configs[j] = configs[j], configs[i]
			}
		}
	}

	start := (page - 1) * perPage
	end := start + perPage
	total := int64(len(configs))

	if start > len(configs) {
		return []DriverConfig{}, total, nil
	}
	if end > len(configs) {
		end = len(configs)
	}

	return configs[start:end], total, nil
}

func (d *D1DatabaseManager) CreateDriverConfig(ctx context.Context, config DriverConfig) error {
	fmt.Printf("Creating driver config: %s for user %d - dev mode\n", config.Name, config.UserID)
	config.ID = uint(time.Now().Unix())
	config.Created = time.Now().Format(time.RFC3339)
	config.Modified = time.Now().Format(time.RFC3339)

	key := fmt.Sprintf("%d_%s", config.UserID, config.Name)
	driversMap[key] = &config
	return nil
}

func (d *D1DatabaseManager) UpdateDriverConfig(ctx context.Context, config DriverConfig) error {
	fmt.Printf("Updating driver config: %s for user %d - dev mode\n", config.Name, config.UserID)
	key := fmt.Sprintf("%d_%s", config.UserID, config.Name)
	if existing, exists := driversMap[key]; exists {
		config.ID = existing.ID
		config.Created = existing.Created
		config.Modified = time.Now().Format(time.RFC3339)
		driversMap[key] = &config
		return nil
	}
	return fmt.Errorf("driver config not found: %s for user %d", config.Name, config.UserID)
}

func (d *D1DatabaseManager) DeleteUserDriverConfig(ctx context.Context, userID, configID uint) error {
	fmt.Printf("Deleting driver config: %d for user %d - dev mode\n", configID, userID)
	for key, config := range driversMap {
		if config.ID == configID && config.UserID == userID {
			delete(driversMap, key)
			return nil
		}
	}
	return fmt.Errorf("driver config not found with id: %d for user %d", configID, userID)
}

// Offline download config management methods (dev mode)
func (d *D1DatabaseManager) GetUserOfflineDownloadConfigs(ctx context.Context, userID uint) ([]*OfflineDownloadConfig, error) {
	fmt.Printf("Getting offline download configs for user %d - dev mode\n", userID)

	// 返回一些模拟数据
	configs := []*OfflineDownloadConfig{
		{
			ID:          1,
			UserID:      userID,
			ToolName:    "aria2",
			Config:      `{"uri": "http://localhost:6800/jsonrpc", "secret": ""}`,
			TempDirPath: "",
			Enabled:     true,
			Created:     time.Now().Format(time.RFC3339),
			Modified:    time.Now().Format(time.RFC3339),
		},
		{
			ID:          2,
			UserID:      userID,
			ToolName:    "115",
			Config:      `{}`,
			TempDirPath: "/downloads/temp",
			Enabled:     false,
			Created:     time.Now().Format(time.RFC3339),
			Modified:    time.Now().Format(time.RFC3339),
		},
	}

	return configs, nil
}

func (d *D1DatabaseManager) CreateOfflineDownloadConfig(ctx context.Context, config OfflineDownloadConfig) error {
	fmt.Printf("Creating offline download config: %s for user %d - dev mode\n", config.ToolName, config.UserID)
	config.ID = uint(time.Now().Unix())
	config.Created = time.Now().Format(time.RFC3339)
	config.Modified = time.Now().Format(time.RFC3339)

	key := fmt.Sprintf("%d_%s", config.UserID, config.ToolName)
	offlineDownloadConfigs[key] = &config
	return nil
}

func (d *D1DatabaseManager) UpdateOfflineDownloadConfig(ctx context.Context, config OfflineDownloadConfig) error {
	fmt.Printf("Updating offline download config: %s for user %d - dev mode\n", config.ToolName, config.UserID)
	key := fmt.Sprintf("%d_%s", config.UserID, config.ToolName)
	if existing, exists := offlineDownloadConfigs[key]; exists {
		config.ID = existing.ID
		config.Created = existing.Created
		config.Modified = time.Now().Format(time.RFC3339)
		offlineDownloadConfigs[key] = &config
		return nil
	}
	return fmt.Errorf("offline download config not found: %s for user %d", config.ToolName, config.UserID)
}

func (d *D1DatabaseManager) DeleteUserOfflineDownloadConfig(ctx context.Context, userID uint, toolName string) error {
	fmt.Printf("Deleting offline download config: %s for user %d - dev mode\n", toolName, userID)
	key := fmt.Sprintf("%d_%s", userID, toolName)
	if _, exists := offlineDownloadConfigs[key]; exists {
		delete(offlineDownloadConfigs, key)
		return nil
	}
	return fmt.Errorf("offline download config not found: %s for user %d", toolName, userID)
}

// Offline download task management methods (dev mode)
func (d *D1DatabaseManager) GetUserOfflineDownloadTasks(ctx context.Context, userID uint, page, perPage int) ([]*OfflineDownloadTask, int64, error) {
	fmt.Printf("Getting offline download tasks for user %d (page %d, per_page %d) - dev mode\n", userID, page, perPage)

	// 返回一些模拟数据
	tasks := []*OfflineDownloadTask{
		{
			ID:           1,
			UserID:       userID,
			ConfigID:     1,
			URLs:         `["http://example.com/file1.zip", "http://example.com/file2.zip"]`,
			DstPath:      "/downloads",
			Tool:         "aria2",
			Status:       "completed",
			Progress:     100,
			DeletePolicy: "delete_on_complete",
			Error:        "",
			Created:      time.Now().Add(-time.Hour).Format(time.RFC3339),
			Updated:      time.Now().Format(time.RFC3339),
		},
		{
			ID:           2,
			UserID:       userID,
			ConfigID:     1,
			URLs:         `["http://example.com/file3.zip"]`,
			DstPath:      "/downloads",
			Tool:         "aria2",
			Status:       "running",
			Progress:     45,
			DeletePolicy: "keep",
			Error:        "",
			Created:      time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			Updated:      time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
	}

	// 简单分页
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

func (d *D1DatabaseManager) CreateOfflineDownloadTask(ctx context.Context, task OfflineDownloadTask) error {
	fmt.Printf("Creating offline download task for user %d - dev mode\n", task.UserID)
	task.ID = uint(time.Now().Unix())
	task.Status = "pending"
	task.Progress = 0
	task.Created = time.Now().Format(time.RFC3339)
	task.Updated = time.Now().Format(time.RFC3339)

	offlineDownloadTasks[task.ID] = &task
	return nil
}

func (d *D1DatabaseManager) UpdateOfflineDownloadTaskStatus(ctx context.Context, userID, taskID uint, status string, progress int, errorMsg string) error {
	fmt.Printf("Updating offline download task %d status to %s (progress: %d%%) for user %d - dev mode\n",
		taskID, status, progress, userID)

	if task, exists := offlineDownloadTasks[taskID]; exists && task.UserID == userID {
		task.Status = status
		task.Progress = progress
		task.Error = errorMsg
		task.Updated = time.Now().Format(time.RFC3339)
		return nil
	}
	return fmt.Errorf("offline download task not found: %d for user %d", taskID, userID)
}

func (d *D1DatabaseManager) DeleteOfflineDownloadTask(ctx context.Context, userID, taskID uint) error {
	fmt.Printf("Deleting offline download task: %d for user %d - dev mode\n", taskID, userID)
	if task, exists := offlineDownloadTasks[taskID]; exists && task.UserID == userID {
		delete(offlineDownloadTasks, taskID)
		return nil
	}
	return fmt.Errorf("offline download task not found: %d for user %d", taskID, userID)
}

// GetStorages 获取存储列表（模拟）
func (dm *D1DatabaseManager) GetStorages(ctx context.Context, page, perPage int) ([]model.Storage, int64, error) {
	// 模拟存储数据
	storages := []model.Storage{
		{
			ID:        1,
			MountPath: "/test",
			Driver:    "S3",
			Order:     1,
			Status:    "WORK",
			Disabled:  false,
			Modified:  time.Now(),
		},
		{
			ID:        2,
			MountPath: "/local",
			Driver:    "Local",
			Order:     2,
			Status:    "WORK",
			Disabled:  false,
			Modified:  time.Now(),
		},
	}

	start := (page - 1) * perPage
	end := start + perPage
	total := int64(len(storages))

	if start > len(storages) {
		return []model.Storage{}, total, nil
	}
	if end > len(storages) {
		end = len(storages)
	}

	return storages[start:end], total, nil
}

// CreateStorage 创建存储（模拟）
func (dm *D1DatabaseManager) CreateStorage(ctx context.Context, storage model.Storage) (uint, error) {
	id := uint(time.Now().Unix())
	storage.ID = id
	storage.Modified = time.Now()

	fmt.Printf("Created storage: %s with driver: %s (dev mode)\n", storage.MountPath, storage.Driver)
	return id, nil
}

// Close 关闭数据库连接（模拟）
func (dm *D1DatabaseManager) Close() error {
	fmt.Printf("Closed database connection: %s (dev mode)\n", dm.dbName)
	return nil
}
