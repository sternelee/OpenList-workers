package drivers

import (
	"context"
	"fmt"
	"sync"

	"github.com/OpenListTeam/OpenList-workers/workers/db"
	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

// DriverManager 驱动管理器
type DriverManager struct {
	mu       sync.RWMutex
	drivers  map[string]Driver // mountPath -> driver
	repos    *db.Repositories
}

// NewDriverManager 创建驱动管理器
func NewDriverManager(repos *db.Repositories) *DriverManager {
	return &DriverManager{
		drivers: make(map[string]Driver),
		repos:   repos,
	}
}

// LoadDrivers 加载所有启用的驱动
func (m *DriverManager) LoadDrivers(ctx context.Context) error {
	storages, err := m.repos.Storage.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to load storages: %v", err)
	}

	for _, storage := range storages {
		if err := m.LoadDriver(ctx, storage); err != nil {
			// 记录错误但继续加载其他驱动
			fmt.Printf("Failed to load driver %s: %v\n", storage.MountPath, err)
			// 更新存储状态为错误
			storage.SetStatus("error")
			m.repos.Storage.Update(ctx, storage)
		}
	}

	return nil
}

// LoadDriver 加载单个驱动
func (m *DriverManager) LoadDriver(ctx context.Context, storage *models.Storage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果驱动已存在，先卸载
	if existingDriver, exists := m.drivers[storage.MountPath]; exists {
		existingDriver.Drop(ctx)
		delete(m.drivers, storage.MountPath)
	}

	// 创建新驱动
	driver, err := CreateDriver(storage)
	if err != nil {
		return fmt.Errorf("failed to create driver: %v", err)
	}

	// 初始化驱动
	if err := InitializeDriver(ctx, driver); err != nil {
		return fmt.Errorf("failed to initialize driver: %v", err)
	}

	// 存储驱动实例
	m.drivers[storage.MountPath] = driver

	// 更新存储状态
	storage.SetStatus("work")
	if err := m.repos.Storage.Update(ctx, storage); err != nil {
		fmt.Printf("Failed to update storage status: %v\n", err)
	}

	return nil
}

// UnloadDriver 卸载驱动
func (m *DriverManager) UnloadDriver(ctx context.Context, mountPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	driver, exists := m.drivers[mountPath]
	if !exists {
		return fmt.Errorf("driver %s not found", mountPath)
	}

	// 停止驱动
	if err := driver.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop driver: %v", err)
	}

	// 从管理器中移除
	delete(m.drivers, mountPath)

	return nil
}

// GetDriver 获取驱动实例
func (m *DriverManager) GetDriver(mountPath string) (Driver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	driver, exists := m.drivers[mountPath]
	if !exists {
		return nil, fmt.Errorf("driver %s not found", mountPath)
	}

	return driver, nil
}

// GetDriverByPath 根据路径获取驱动
func (m *DriverManager) GetDriverByPath(path string) (Driver, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 查找最长匹配的挂载路径
	var bestMatch string
	var bestDriver Driver

	for mountPath, driver := range m.drivers {
		if len(mountPath) > len(bestMatch) && matchPath(path, mountPath) {
			bestMatch = mountPath
			bestDriver = driver
		}
	}

	if bestDriver == nil {
		return nil, "", fmt.Errorf("no driver found for path: %s", path)
	}

	// 计算相对路径
	relativePath := path[len(bestMatch):]
	if relativePath == "" {
		relativePath = "/"
	}

	return bestDriver, relativePath, nil
}

// ListDrivers 列出所有活动驱动
func (m *DriverManager) ListDrivers() map[string]Driver {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Driver)
	for mountPath, driver := range m.drivers {
		result[mountPath] = driver
	}
	return result
}

// ReloadDriver 重新加载驱动
func (m *DriverManager) ReloadDriver(ctx context.Context, mountPath string) error {
	// 获取存储配置
	storage, err := m.repos.Storage.GetByMountPath(ctx, mountPath)
	if err != nil {
		return fmt.Errorf("failed to get storage: %v", err)
	}

	// 重新加载驱动
	return m.LoadDriver(ctx, storage)
}

// RefreshAllDrivers 刷新所有驱动
func (m *DriverManager) RefreshAllDrivers(ctx context.Context) error {
	// 卸载所有现有驱动
	m.mu.Lock()
	for mountPath, driver := range m.drivers {
		driver.Drop(ctx)
		delete(m.drivers, mountPath)
	}
	m.mu.Unlock()

	// 重新加载所有驱动
	return m.LoadDrivers(ctx)
}

// matchPath 检查路径是否匹配挂载点
func matchPath(path, mountPath string) bool {
	if mountPath == "/" {
		return true
	}
	
	if len(path) < len(mountPath) {
		return false
	}
	
	if path[:len(mountPath)] != mountPath {
		return false
	}
	
	// 确保是完整路径匹配
	if len(path) == len(mountPath) {
		return true
	}
	
	return path[len(mountPath)] == '/'
}

// DriverService 驱动服务
type DriverService struct {
	manager *DriverManager
}

// NewDriverService 创建驱动服务
func NewDriverService(repos *db.Repositories) *DriverService {
	return &DriverService{
		manager: NewDriverManager(repos),
	}
}

// Initialize 初始化服务
func (s *DriverService) Initialize(ctx context.Context) error {
	return s.manager.LoadDrivers(ctx)
}

// ListFiles 列出文件
func (s *DriverService) ListFiles(ctx context.Context, path string, args ListArgs) ([]Obj, error) {
	driver, relativePath, err := s.manager.GetDriverByPath(path)
	if err != nil {
		return nil, err
	}

	// 创建目录对象
	dir := &Object{
		Path:     relativePath,
		Name:     relativePath,
		IsFolder: true,
	}

	return driver.List(ctx, dir, args)
}

// GetFileLink 获取文件链接
func (s *DriverService) GetFileLink(ctx context.Context, path string, args LinkArgs) (*Link, error) {
	driver, relativePath, err := s.manager.GetDriverByPath(path)
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

// GetFile 获取文件信息
func (s *DriverService) GetFile(ctx context.Context, path string) (Obj, error) {
	driver, relativePath, err := s.manager.GetDriverByPath(path)
	if err != nil {
		return nil, err
	}

	// 检查驱动是否支持Get操作
	if getter, ok := driver.(Getter); ok {
		return getter.Get(ctx, relativePath)
	}

	return nil, fmt.Errorf("driver does not support Get operation")
} 