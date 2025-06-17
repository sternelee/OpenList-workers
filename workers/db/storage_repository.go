package db

import (
	"context"
	"database/sql"

	"github.com/sternelee/OpenList-workers/workers/models"
)

type storageRepository struct {
	db *sql.DB
}

// NewStorageRepository 创建存储仓库
func NewStorageRepository(db *sql.DB) StorageRepository {
	return &storageRepository{db: db}
}

// Create 创建存储配置
func (r *storageRepository) Create(ctx context.Context, storage *models.Storage) error {
	query := `
		INSERT INTO storages (user_id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		                     modified, disabled, disable_index, enable_sign, is_public, allow_guest, require_auth,
		                     order_by, order_direction, extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		storage.UserID, storage.MountPath, storage.OrderIndex, storage.Driver, storage.CacheExpiration,
		storage.Status, storage.Addition, storage.Remark, storage.Modified,
		storage.Disabled, storage.DisableIndex, storage.EnableSign,
		storage.IsPublic, storage.AllowGuest, storage.RequireAuth,
		storage.OrderBy, storage.OrderDirection, storage.ExtractFolder,
		storage.WebProxy, storage.WebdavPolicy, storage.ProxyRange, storage.DownProxyUrl,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	storage.ID = int(id)
	return nil
}

// GetByID 根据ID获取存储配置
func (r *storageRepository) GetByID(ctx context.Context, id int) (*models.Storage, error) {
	query := `
		SELECT id, user_id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, is_public, allow_guest, require_auth,
		       order_by, order_direction, extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages WHERE id = ?
	`
	storage := &models.Storage{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&storage.ID, &storage.UserID, &storage.MountPath, &storage.OrderIndex, &storage.Driver,
		&storage.CacheExpiration, &storage.Status, &storage.Addition, &storage.Remark,
		&storage.Modified, &storage.Disabled, &storage.DisableIndex, &storage.EnableSign,
		&storage.IsPublic, &storage.AllowGuest, &storage.RequireAuth,
		&storage.OrderBy, &storage.OrderDirection, &storage.ExtractFolder,
		&storage.WebProxy, &storage.WebdavPolicy, &storage.ProxyRange, &storage.DownProxyUrl,
		&storage.CreatedAt, &storage.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return storage, nil
}

// GetByMountPath 根据挂载路径获取存储配置
func (r *storageRepository) GetByMountPath(ctx context.Context, mountPath string) (*models.Storage, error) {
	query := `
		SELECT id, user_id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, is_public, allow_guest, require_auth,
		       order_by, order_direction, extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages WHERE mount_path = ?
	`
	storage := &models.Storage{}
	err := r.db.QueryRowContext(ctx, query, mountPath).Scan(
		&storage.ID, &storage.UserID, &storage.MountPath, &storage.OrderIndex, &storage.Driver,
		&storage.CacheExpiration, &storage.Status, &storage.Addition, &storage.Remark,
		&storage.Modified, &storage.Disabled, &storage.DisableIndex, &storage.EnableSign,
		&storage.IsPublic, &storage.AllowGuest, &storage.RequireAuth,
		&storage.OrderBy, &storage.OrderDirection, &storage.ExtractFolder,
		&storage.WebProxy, &storage.WebdavPolicy, &storage.ProxyRange, &storage.DownProxyUrl,
		&storage.CreatedAt, &storage.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return storage, nil
}

// GetByUserAndPath 根据用户ID和挂载路径获取存储配置
func (r *storageRepository) GetByUserAndPath(ctx context.Context, userID int, mountPath string) (*models.Storage, error) {
	query := `
		SELECT id, user_id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, is_public, allow_guest, require_auth,
		       order_by, order_direction, extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages WHERE user_id = ? AND mount_path = ?
	`
	storage := &models.Storage{}
	err := r.db.QueryRowContext(ctx, query, userID, mountPath).Scan(
		&storage.ID, &storage.UserID, &storage.MountPath, &storage.OrderIndex, &storage.Driver,
		&storage.CacheExpiration, &storage.Status, &storage.Addition, &storage.Remark,
		&storage.Modified, &storage.Disabled, &storage.DisableIndex, &storage.EnableSign,
		&storage.IsPublic, &storage.AllowGuest, &storage.RequireAuth,
		&storage.OrderBy, &storage.OrderDirection, &storage.ExtractFolder,
		&storage.WebProxy, &storage.WebdavPolicy, &storage.ProxyRange, &storage.DownProxyUrl,
		&storage.CreatedAt, &storage.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return storage, nil
}

// ListByDriver 根据驱动类型获取存储配置列表
func (r *storageRepository) ListByDriver(ctx context.Context, driver string) ([]*models.Storage, error) {
	query := `
		SELECT id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, order_by, order_direction, 
		       extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages WHERE driver = ? ORDER BY order_index
	`
	return r.queryStorages(ctx, query, driver)
}

// ListEnabled 获取启用的存储配置列表
func (r *storageRepository) ListEnabled(ctx context.Context) ([]*models.Storage, error) {
	query := `
		SELECT id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, order_by, order_direction, 
		       extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages WHERE disabled = FALSE ORDER BY order_index
	`
	return r.queryStorages(ctx, query)
}

// Update 更新存储配置
func (r *storageRepository) Update(ctx context.Context, storage *models.Storage) error {
	query := `
		UPDATE storages SET mount_path = ?, order_index = ?, driver = ?, cache_expiration = ?, 
		       status = ?, addition = ?, remark = ?, modified = ?, disabled = ?, 
		       disable_index = ?, enable_sign = ?, order_by = ?, order_direction = ?, 
		       extract_folder = ?, web_proxy = ?, webdav_policy = ?, proxy_range = ?, 
		       down_proxy_url = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		storage.MountPath, storage.OrderIndex, storage.Driver, storage.CacheExpiration,
		storage.Status, storage.Addition, storage.Remark, storage.Modified,
		storage.Disabled, storage.DisableIndex, storage.EnableSign,
		storage.OrderBy, storage.OrderDirection, storage.ExtractFolder,
		storage.WebProxy, storage.WebdavPolicy, storage.ProxyRange, storage.DownProxyUrl,
		storage.ID,
	)
	return err
}

// Delete 删除存储配置
func (r *storageRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM storages WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List 获取存储配置列表
func (r *storageRepository) List(ctx context.Context, limit, offset int) ([]*models.Storage, error) {
	query := `
		SELECT id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, order_by, order_direction, 
		       extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages ORDER BY order_index LIMIT ? OFFSET ?
	`
	return r.queryStorages(ctx, query, limit, offset)
}

// queryStorages 查询存储配置的通用方法
func (r *storageRepository) queryStorages(ctx context.Context, query string, args ...interface{}) ([]*models.Storage, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var storages []*models.Storage
	for rows.Next() {
		storage := &models.Storage{}
		err := rows.Scan(
			&storage.ID, &storage.MountPath, &storage.OrderIndex, &storage.Driver,
			&storage.CacheExpiration, &storage.Status, &storage.Addition, &storage.Remark,
			&storage.Modified, &storage.Disabled, &storage.DisableIndex, &storage.EnableSign,
			&storage.OrderBy, &storage.OrderDirection, &storage.ExtractFolder,
			&storage.WebProxy, &storage.WebdavPolicy, &storage.ProxyRange, &storage.DownProxyUrl,
			&storage.CreatedAt, &storage.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		storages = append(storages, storage)
	}
	return storages, nil
} 