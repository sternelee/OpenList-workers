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

// NewD1DatabaseManager 创建开发环境的数据库管理器
func NewD1DatabaseManager(dbName string) (*D1DatabaseManager, error) {
	fmt.Printf("Creating development database manager for: %s\n", dbName)
	return &D1DatabaseManager{dbName: dbName}, nil
}

// CreateTables 创建数据库表（模拟）
func (dm *D1DatabaseManager) CreateTables(ctx context.Context) error {
	tables := []string{
		"users",
		"driver_configs (user-based)",
	}

	for _, table := range tables {
		fmt.Printf("Created table: %s (dev mode)\n", table)
	}

	return nil
}

// GetUserDriverConfigs 获取用户的驱动配置列表（模拟）
func (dm *D1DatabaseManager) GetUserDriverConfigs(ctx context.Context, userID uint, page, perPage int) ([]DriverConfig, int64, error) {
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

// CreateDriverConfig 创建驱动配置（模拟）
func (dm *D1DatabaseManager) CreateDriverConfig(ctx context.Context, config DriverConfig) error {
	config.ID = uint(time.Now().Unix())
	config.Created = time.Now().Format(time.RFC3339)
	config.Modified = time.Now().Format(time.RFC3339)

	key := fmt.Sprintf("%d_%s", config.UserID, config.Name)
	driversMap[key] = &config
	fmt.Printf("Created driver config: %s for user %d (dev mode)\n", config.Name, config.UserID)
	return nil
}

// UpdateDriverConfig 更新驱动配置（模拟）
func (dm *D1DatabaseManager) UpdateDriverConfig(ctx context.Context, config DriverConfig) error {
	key := fmt.Sprintf("%d_%s", config.UserID, config.Name)
	if existing, exists := driversMap[key]; exists {
		config.ID = existing.ID
		config.Created = existing.Created
		config.Modified = time.Now().Format(time.RFC3339)
		driversMap[key] = &config
		fmt.Printf("Updated driver config: %s for user %d (dev mode)\n", config.Name, config.UserID)
		return nil
	}
	return fmt.Errorf("driver config not found: %s for user %d", config.Name, config.UserID)
}

// DeleteUserDriverConfig 删除用户的驱动配置（模拟）
func (dm *D1DatabaseManager) DeleteUserDriverConfig(ctx context.Context, userID, id uint) error {
	for key, config := range driversMap {
		if config.ID == id && config.UserID == userID {
			delete(driversMap, key)
			fmt.Printf("Deleted driver config: %s for user %d (dev mode)\n", config.Name, userID)
			return nil
		}
	}
	return fmt.Errorf("driver config not found with id: %d for user %d", id, userID)
}

// GetUsers 获取用户列表（模拟）
func (dm *D1DatabaseManager) GetUsers(ctx context.Context, page, perPage int) ([]model.User, int64, error) {
	// 模拟用户数据
	users := []model.User{
		{
			ID:         1,
			Username:   "admin",
			BasePath:   "/",
			Role:       model.ADMIN,
			Disabled:   false,
			Permission: 0x30FF,
			Authn:      "[]",
		},
		{
			ID:         2,
			Username:   "guest",
			BasePath:   "/",
			Role:       model.GUEST,
			Disabled:   true,
			Permission: 0,
			Authn:      "[]",
		},
	}

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

// CreateUser 创建用户（模拟）
func (dm *D1DatabaseManager) CreateUser(ctx context.Context, user model.User) error {
	user.ID = uint(time.Now().Unix())
	if user.Authn == "" {
		user.Authn = "[]"
	}
	usersMap[user.ID] = &user
	fmt.Printf("Created user: %s (dev mode)\n", user.Username)
	return nil
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
