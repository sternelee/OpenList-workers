package db

import (
	"context"
	"database/sql"

	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

// Repository 接口定义了基本的数据库操作
type Repository[T any] interface {
	Create(ctx context.Context, item *T) error
	GetByID(ctx context.Context, id int) (*T, error)
	Update(ctx context.Context, item *T) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*T, error)
}

// UserRepository 用户仓库接口
type UserRepository interface {
	Repository[models.User]
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetBySsoID(ctx context.Context, ssoID string) (*models.User, error)
}

// StorageRepository 存储仓库接口
type StorageRepository interface {
	Repository[models.Storage]
	GetByMountPath(ctx context.Context, mountPath string) (*models.Storage, error)
	GetByUserAndPath(ctx context.Context, userID int, mountPath string) (*models.Storage, error)
	ListByDriver(ctx context.Context, driver string) ([]*models.Storage, error)
	ListEnabled(ctx context.Context) ([]*models.Storage, error)
	ListByUser(ctx context.Context, userID int, limit, offset int) ([]*models.Storage, error)
	ListUserEnabled(ctx context.Context, userID int) ([]*models.Storage, error)
	ListPublic(ctx context.Context) ([]*models.Storage, error)
	CheckUserAccess(ctx context.Context, userID int, storageID int) (bool, error)
	DeleteByUser(ctx context.Context, userID int) error
}

// SettingRepository 设置仓库接口
type SettingRepository interface {
	Repository[models.SettingItem]
	GetByKey(ctx context.Context, key string) (*models.SettingItem, error)
	GetByGroup(ctx context.Context, groupID int) ([]*models.SettingItem, error)
	SetValue(ctx context.Context, key, value string) error
}

// MetaRepository 元数据仓库接口
type MetaRepository interface {
	Repository[models.Meta]
	GetByPath(ctx context.Context, path string) (*models.Meta, error)
}

// SearchNodeRepository 搜索节点仓库接口
type SearchNodeRepository interface {
	Repository[models.SearchNode]
	GetByParent(ctx context.Context, parent string) ([]*models.SearchNode, error)
	Search(ctx context.Context, req *models.SearchReq) ([]*models.SearchNode, int64, error)
	DeleteByParent(ctx context.Context, parent string) error
}

// TaskRepository 任务仓库接口
type TaskRepository interface {
	Repository[models.TaskItem]
	GetByKey(ctx context.Context, key string) (*models.TaskItem, error)
}

// SSHKeyRepository SSH密钥仓库接口
type SSHKeyRepository interface {
	Repository[models.SSHPublicKey]
	GetByUserID(ctx context.Context, userID int) ([]*models.SSHPublicKey, error)
	GetByFingerprint(ctx context.Context, fingerprint string) (*models.SSHPublicKey, error)
	DeleteByUserID(ctx context.Context, userID int) error
}

// Repositories 包含所有仓库接口
type Repositories struct {
	User       UserRepository
	Storage    StorageRepository
	Setting    SettingRepository
	Meta       MetaRepository
	SearchNode SearchNodeRepository
	Task       TaskRepository
	SSHKey     SSHKeyRepository
}

// NewRepositories 创建新的仓库实例
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:       NewUserRepository(db),
		Storage:    NewStorageRepository(db),
		Setting:    NewSettingRepository(db),
		Meta:       NewMetaRepository(db),
		SearchNode: NewSearchNodeRepository(db),
		Task:       NewTaskRepository(db),
		SSHKey:     NewSSHKeyRepository(db),
	}
} 