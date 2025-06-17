package drivers

import (
	"context"
	"fmt"
	"sync"

	"github.com/sternelee/OpenList-workers/workers/db"
	"github.com/sternelee/OpenList-workers/workers/models"
)

// UserDriverManager 用户级别的驱动管理器
type UserDriverManager struct {
	mu         sync.RWMutex
	userStores map[int]map[string]Driver // userID -> mountPath -> driver
	repos      *db.Repositories
}

// NewUserDriverManager 创建用户驱动管理器
func NewUserDriverManager(repos *db.Repositories) *UserDriverManager {
	return &UserDriverManager{
		userStores: make(map[int]map[string]Driver),
		repos:      repos,
	}
}

// LoadUserDrivers 加载用户的所有驱动
func (m *UserDriverManager) LoadUserDrivers(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清理现有驱动
	if userDrivers, exists := m.userStores[userID]; exists {
		for _, driver := range userDrivers {
			driver.Drop(ctx)
		}
		delete(m.userStores, userID)
	}

	// 加载用户启用的存储
	storages, err := m.repos.Storage.ListUserEnabled(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user storages: %v", err)
	}

	// 创建用户存储映射
	userDrivers := make(map[string]Driver)
	for _, storage := range storages {
		if err := m.loadSingleDriver(ctx, storage, userDrivers); err != nil {
			fmt.Printf("Failed to load driver for user %d, storage %s: %v\n", userID, storage.MountPath, err)
			// 继续加载其他驱动
		}
	}

	m.userStores[userID] = userDrivers
	return nil
}

// loadSingleDriver 加载单个驱动
func (m *UserDriverManager) loadSingleDriver(ctx context.Context, storage *models.Storage, userDrivers map[string]Driver) error {
	// 创建驱动实例
	driver, err := CreateDriver(storage)
	if err != nil {
		return fmt.Errorf("failed to create driver: %v", err)
	}

	// 初始化驱动
	if err := InitializeDriver(ctx, driver); err != nil {
		return fmt.Errorf("failed to initialize driver: %v", err)
	}

	// 存储驱动实例
	userDrivers[storage.MountPath] = driver

	// 更新存储状态
	storage.SetStatus("work")
	if err := m.repos.Storage.Update(ctx, storage); err != nil {
		fmt.Printf("Failed to update storage status: %v\n", err)
	}

	return nil
}

// GetUserDriver 获取用户的特定驱动
func (m *UserDriverManager) GetUserDriver(userID int, mountPath string) (Driver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userDrivers, exists := m.userStores[userID]
	if !exists {
		return nil, fmt.Errorf("user %d has no loaded drivers", userID)
	}

	driver, exists := userDrivers[mountPath]
	if !exists {
		return nil, fmt.Errorf("driver %s not found for user %d", mountPath, userID)
	}

	return driver, nil
}

// GetUserDriverByPath 根据路径获取用户驱动
func (m *UserDriverManager) GetUserDriverByPath(userID int, path string) (Driver, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userDrivers, exists := m.userStores[userID]
	if !exists {
		return nil, "", fmt.Errorf("user %d has no loaded drivers", userID)
	}

	// 查找最长匹配的挂载路径
	var bestMatch string
	var bestDriver Driver

	for mountPath, driver := range userDrivers {
		if len(mountPath) > len(bestMatch) && matchPath(path, mountPath) {
			bestMatch = mountPath
			bestDriver = driver
		}
	}

	if bestDriver == nil {
		return nil, "", fmt.Errorf("no driver found for user %d, path: %s", userID, path)
	}

	// 计算相对路径
	relativePath := path[len(bestMatch):]
	if relativePath == "" {
		relativePath = "/"
	}

	return bestDriver, relativePath, nil
}

// CheckUserAccess 检查用户对路径的访问权限
func (m *UserDriverManager) CheckUserAccess(ctx context.Context, userID int, path string) (bool, *models.Storage, error) {
	// 先检查用户自己的存储
	if storage, err := m.checkUserOwnedStorage(ctx, userID, path); err == nil {
		return true, storage, nil
	}

	// 检查公开存储
	if storage, err := m.checkPublicStorage(ctx, path); err == nil {
		return true, storage, nil
	}

	return false, nil, fmt.Errorf("access denied to path: %s", path)
}

// checkUserOwnedStorage 检查用户拥有的存储
func (m *UserDriverManager) checkUserOwnedStorage(ctx context.Context, userID int, path string) (*models.Storage, error) {
	// 找到最匹配的挂载路径
	storages, err := m.repos.Storage.ListUserEnabled(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, storage := range storages {
		if matchPath(path, storage.MountPath) {
			return storage, nil
		}
	}

	return nil, fmt.Errorf("no user storage found")
}

// checkPublicStorage 检查公开存储
func (m *UserDriverManager) checkPublicStorage(ctx context.Context, path string) (*models.Storage, error) {
	storages, err := m.repos.Storage.ListPublic(ctx)
	if err != nil {
		return nil, err
	}

	for _, storage := range storages {
		if matchPath(path, storage.MountPath) {
			return storage, nil
		}
	}

	return nil, fmt.Errorf("no public storage found")
}

// UnloadUserDrivers 卸载用户的所有驱动
func (m *UserDriverManager) UnloadUserDrivers(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userDrivers, exists := m.userStores[userID]
	if !exists {
		return nil // 用户没有加载的驱动
	}

	// 停止所有驱动
	for _, driver := range userDrivers {
		if err := driver.Drop(ctx); err != nil {
			fmt.Printf("Failed to drop driver: %v\n", err)
		}
	}

	// 从管理器中移除
	delete(m.userStores, userID)
	return nil
}

// ListUserDrivers 列出用户的所有驱动
func (m *UserDriverManager) ListUserDrivers(userID int) map[string]Driver {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userDrivers, exists := m.userStores[userID]
	if !exists {
		return make(map[string]Driver)
	}

	result := make(map[string]Driver)
	for mountPath, driver := range userDrivers {
		result[mountPath] = driver
	}
	return result
}

// ReloadUserDriver 重新加载用户的特定驱动
func (m *UserDriverManager) ReloadUserDriver(ctx context.Context, userID int, mountPath string) error {
	// 获取存储配置
	storage, err := m.repos.Storage.GetByUserAndPath(ctx, userID, mountPath)
	if err != nil {
		return fmt.Errorf("failed to get storage: %v", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取用户驱动映射
	userDrivers, exists := m.userStores[userID]
	if !exists {
		userDrivers = make(map[string]Driver)
		m.userStores[userID] = userDrivers
	}

	// 如果驱动已存在，先卸载
	if existingDriver, exists := userDrivers[mountPath]; exists {
		existingDriver.Drop(ctx)
		delete(userDrivers, mountPath)
	}

	// 重新加载驱动
	return m.loadSingleDriver(ctx, storage, userDrivers)
}

// UserDriverService 用户驱动服务
type UserDriverService struct {
	manager *UserDriverManager
}

// NewUserDriverService 创建用户驱动服务
func NewUserDriverService(repos *db.Repositories) *UserDriverService {
	return &UserDriverService{
		manager: NewUserDriverManager(repos),
	}
}

// InitializeUser 初始化用户的驱动
func (s *UserDriverService) InitializeUser(ctx context.Context, userID int) error {
	return s.manager.LoadUserDrivers(ctx, userID)
}

// ListUserFiles 列出用户文件
func (s *UserDriverService) ListUserFiles(ctx context.Context, userID int, path string, args ListArgs) ([]Obj, error) {
	// 检查访问权限
	hasAccess, storage, err := s.manager.CheckUserAccess(ctx, userID, path)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}

	// 获取驱动
	driver, relativePath, err := s.manager.GetUserDriverByPath(userID, path)
	if err != nil {
		// 如果用户没有该驱动，尝试从公开存储获取
		return s.getFromPublicStorage(ctx, path, args)
	}

	// 创建目录对象
	dir := &Object{
		Path:     relativePath,
		Name:     relativePath,
		IsFolder: true,
	}

	return driver.List(ctx, dir, args)
}

// getFromPublicStorage 从公开存储获取文件
func (s *UserDriverService) getFromPublicStorage(ctx context.Context, path string, args ListArgs) ([]Obj, error) {
	// 获取公开存储
	storages, err := s.manager.repos.Storage.ListPublic(ctx)
	if err != nil {
		return nil, err
	}

	for _, storage := range storages {
		if matchPath(path, storage.MountPath) {
			// 创建临时驱动
			driver, err := CreateDriver(storage)
			if err != nil {
				continue
			}

			if err := InitializeDriver(ctx, driver); err != nil {
				continue
			}
			defer driver.Drop(ctx)

			// 计算相对路径
			relativePath := path[len(storage.MountPath):]
			if relativePath == "" {
				relativePath = "/"
			}

			dir := &Object{
				Path:     relativePath,
				Name:     relativePath,
				IsFolder: true,
			}

			return driver.List(ctx, dir, args)
		}
	}

	return nil, fmt.Errorf("no accessible storage found for path: %s", path)
}

// GetUserFileLink 获取用户文件链接
func (s *UserDriverService) GetUserFileLink(ctx context.Context, userID int, path string, args LinkArgs) (*Link, error) {
	// 检查访问权限
	hasAccess, _, err := s.manager.CheckUserAccess(ctx, userID, path)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}

	// 获取驱动
	driver, relativePath, err := s.manager.GetUserDriverByPath(userID, path)
	if err != nil {
		return nil, err
	}

	// 创建文件对象
	file := &Object{
		Path: relativePath,
		Name: relativePath,
	}

	return driver.Link(ctx, file, args)
}

