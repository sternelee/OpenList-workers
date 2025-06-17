package drivers

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/sternelee/OpenList-workers/workers/models"
)

// DriverConstructor 驱动构造函数类型
type DriverConstructor func() Driver

// DriverRegistry 驱动注册表
type DriverRegistry struct {
	mu          sync.RWMutex
	drivers     map[string]DriverConstructor
	driverInfos map[string]DriverInfo
}

// 全局驱动注册表
var globalRegistry = NewDriverRegistry()

// NewDriverRegistry 创建新的驱动注册表
func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{
		drivers:     make(map[string]DriverConstructor),
		driverInfos: make(map[string]DriverInfo),
	}
}

// RegisterDriver 注册驱动
func RegisterDriver(constructor DriverConstructor) {
	globalRegistry.RegisterDriver(constructor)
}

// RegisterDriver 注册驱动
func (r *DriverRegistry) RegisterDriver(constructor DriverConstructor) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 创建临时驱动实例获取配置信息
	tempDriver := constructor()
	config := tempDriver.Config()

	// 注册驱动
	r.drivers[config.Name] = constructor

	// 注册驱动信息
	r.driverInfos[config.Name] = DriverInfo{
		Config: config,
		Items:  r.parseDriverFields(tempDriver.GetAddition()),
	}
}

// GetDriver 获取驱动构造函数
func GetDriver(name string) (DriverConstructor, error) {
	return globalRegistry.GetDriver(name)
}

// GetDriver 获取驱动构造函数
func (r *DriverRegistry) GetDriver(name string) (DriverConstructor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	constructor, ok := r.drivers[name]
	if !ok {
		return nil, fmt.Errorf("driver %s not found", name)
	}
	return constructor, nil
}

// GetDriverNames 获取所有驱动名称
func GetDriverNames() []string {
	return globalRegistry.GetDriverNames()
}

// GetDriverNames 获取所有驱动名称
func (r *DriverRegistry) GetDriverNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.drivers))
	for name := range r.drivers {
		names = append(names, name)
	}
	return names
}

// GetDriverInfos 获取所有驱动信息
func GetDriverInfos() map[string]DriverInfo {
	return globalRegistry.GetDriverInfos()
}

// GetDriverInfos 获取所有驱动信息
func (r *DriverRegistry) GetDriverInfos() map[string]DriverInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make(map[string]DriverInfo)
	for name, info := range r.driverInfos {
		infos[name] = info
	}
	return infos
}

// GetDriverInfo 获取指定驱动信息
func GetDriverInfo(name string) (DriverInfo, error) {
	return globalRegistry.GetDriverInfo(name)
}

// GetDriverInfo 获取指定驱动信息
func (r *DriverRegistry) GetDriverInfo(name string) (DriverInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.driverInfos[name]
	if !ok {
		return DriverInfo{}, fmt.Errorf("driver %s not found", name)
	}
	return info, nil
}

// CreateDriver 创建驱动实例
func CreateDriver(storage *models.Storage) (Driver, error) {
	return globalRegistry.CreateDriver(storage)
}

// CreateDriver 创建驱动实例
func (r *DriverRegistry) CreateDriver(storage *models.Storage) (Driver, error) {
	constructor, err := r.GetDriver(storage.Driver)
	if err != nil {
		return nil, err
	}

	driver := constructor()

	// 设置存储配置
	driver.SetStorage(storage)

	return driver, nil
}

// InitializeDriver 初始化驱动
func InitializeDriver(ctx context.Context, driver Driver) error {
	return driver.Init(ctx)
}

// parseDriverFields 解析驱动字段信息
func (r *DriverRegistry) parseDriverFields(addition interface{}) []DriverItem {
	if addition == nil {
		return nil
	}

	t := reflect.TypeOf(addition)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var items []DriverItem
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过内嵌字段
		if field.Anonymous {
			continue
		}

		item := DriverItem{
			Name:     field.Name,
			Type:     r.getFieldType(field.Type),
			Required: r.isFieldRequired(field),
			Help:     field.Tag.Get("help"),
			Default:  field.Tag.Get("default"),
			Options:  field.Tag.Get("options"),
		}

		// 从JSON标签获取字段名
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			item.Name = jsonTag
		}

		items = append(items, item)
	}

	return items
}

// getFieldType 获取字段类型
func (r *DriverRegistry) getFieldType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "number"
	case reflect.Bool:
		return "bool"
	default:
		return "string"
	}
}

// isFieldRequired 检查字段是否必需
func (r *DriverRegistry) isFieldRequired(field reflect.StructField) bool {
	tag := field.Tag.Get("required")
	return tag == "true"
}

// DriverItem 驱动配置项
type DriverItem struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Help     string `json:"help"`
	Default  string `json:"default"`
	Options  string `json:"options"`
}

// DriverInfo 驱动信息
type DriverInfo struct {
	Config DriverConfig `json:"config"`
	Items  []DriverItem `json:"items"`
}

