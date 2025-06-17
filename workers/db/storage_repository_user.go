package db

import (
	"context"
	"database/sql"

	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

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

// ListByUser 根据用户ID获取存储配置列表
func (r *storageRepository) ListByUser(ctx context.Context, userID int, limit, offset int) ([]*models.Storage, error) {
	query := `
		SELECT id, user_id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, is_public, allow_guest, require_auth,
		       order_by, order_direction, extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages WHERE user_id = ? ORDER BY order_index LIMIT ? OFFSET ?
	`
	return r.queryStorages(ctx, query, userID, limit, offset)
}

// ListUserEnabled 获取用户启用的存储配置列表
func (r *storageRepository) ListUserEnabled(ctx context.Context, userID int) ([]*models.Storage, error) {
	query := `
		SELECT id, user_id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, is_public, allow_guest, require_auth,
		       order_by, order_direction, extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages WHERE user_id = ? AND disabled = FALSE ORDER BY order_index
	`
	return r.queryStorages(ctx, query, userID)
}

// ListPublic 获取公开访问的存储配置列表
func (r *storageRepository) ListPublic(ctx context.Context) ([]*models.Storage, error) {
	query := `
		SELECT id, user_id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, is_public, allow_guest, require_auth,
		       order_by, order_direction, extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
		FROM storages WHERE is_public = TRUE AND disabled = FALSE ORDER BY order_index
	`
	return r.queryStorages(ctx, query)
}

// CheckUserAccess 检查用户是否有存储访问权限
func (r *storageRepository) CheckUserAccess(ctx context.Context, userID int, storageID int) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM storages 
		WHERE id = ? AND (user_id = ? OR is_public = TRUE OR allow_guest = TRUE)
	`
	err := r.db.QueryRowContext(ctx, query, storageID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteByUser 删除用户的所有存储配置
func (r *storageRepository) DeleteByUser(ctx context.Context, userID int) error {
	query := `DELETE FROM storages WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// UpdateStorageFields 更新存储的所有字段（包括新增字段）
func (r *storageRepository) updateStorageQuery() string {
	return `
		UPDATE storages SET user_id = ?, mount_path = ?, order_index = ?, driver = ?, cache_expiration = ?, 
		       status = ?, addition = ?, remark = ?, modified = ?, disabled = ?, 
		       disable_index = ?, enable_sign = ?, is_public = ?, allow_guest = ?, require_auth = ?,
		       order_by = ?, order_direction = ?, extract_folder = ?, web_proxy = ?, webdav_policy = ?, 
		       proxy_range = ?, down_proxy_url = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
}

// queryStorageFields 更新查询存储的字段列表
func (r *storageRepository) queryStorageFields() string {
	return `
		SELECT id, user_id, mount_path, order_index, driver, cache_expiration, status, addition, remark, 
		       modified, disabled, disable_index, enable_sign, is_public, allow_guest, require_auth,
		       order_by, order_direction, extract_folder, web_proxy, webdav_policy, proxy_range, down_proxy_url, 
		       created_at, updated_at
	`
}

// scanStorage 扫描存储字段到结构体
func (r *storageRepository) scanStorage(rows *sql.Rows) (*models.Storage, error) {
	storage := &models.Storage{}
	err := rows.Scan(
		&storage.ID, &storage.UserID, &storage.MountPath, &storage.OrderIndex, &storage.Driver,
		&storage.CacheExpiration, &storage.Status, &storage.Addition, &storage.Remark,
		&storage.Modified, &storage.Disabled, &storage.DisableIndex, &storage.EnableSign,
		&storage.IsPublic, &storage.AllowGuest, &storage.RequireAuth,
		&storage.OrderBy, &storage.OrderDirection, &storage.ExtractFolder,
		&storage.WebProxy, &storage.WebdavPolicy, &storage.ProxyRange, &storage.DownProxyUrl,
		&storage.CreatedAt, &storage.UpdatedAt,
	)
	return storage, err
} 