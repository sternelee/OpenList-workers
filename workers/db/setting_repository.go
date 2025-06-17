package db

import (
	"context"
	"database/sql"

	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

type settingRepository struct {
	db *sql.DB
}

// NewSettingRepository 创建设置仓库
func NewSettingRepository(db *sql.DB) SettingRepository {
	return &settingRepository{db: db}
}

// Create 创建设置项
func (r *settingRepository) Create(ctx context.Context, setting *models.SettingItem) error {
	query := `
		INSERT INTO settings (key, value, help, type, options, group_id, flag, index_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		setting.Key, setting.Value, setting.Help, setting.Type,
		setting.Options, setting.GroupID, setting.Flag, setting.IndexOrder,
	)
	return err
}

// GetByID 根据ID获取设置项 (这里使用Key作为主键)
func (r *settingRepository) GetByID(ctx context.Context, id int) (*models.SettingItem, error) {
	// 由于设置表使用key作为主键，这个方法可能不太适用
	// 这里提供一个基于索引的查询
	query := `
		SELECT key, value, help, type, options, group_id, flag, index_order, created_at, updated_at
		FROM settings WHERE index_order = ?
	`
	setting := &models.SettingItem{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&setting.Key, &setting.Value, &setting.Help, &setting.Type,
		&setting.Options, &setting.GroupID, &setting.Flag, &setting.IndexOrder,
		&setting.CreatedAt, &setting.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return setting, nil
}

// GetByKey 根据Key获取设置项
func (r *settingRepository) GetByKey(ctx context.Context, key string) (*models.SettingItem, error) {
	query := `
		SELECT key, value, help, type, options, group_id, flag, index_order, created_at, updated_at
		FROM settings WHERE key = ?
	`
	setting := &models.SettingItem{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&setting.Key, &setting.Value, &setting.Help, &setting.Type,
		&setting.Options, &setting.GroupID, &setting.Flag, &setting.IndexOrder,
		&setting.CreatedAt, &setting.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return setting, nil
}

// GetByGroup 根据组ID获取设置项列表
func (r *settingRepository) GetByGroup(ctx context.Context, groupID int) ([]*models.SettingItem, error) {
	query := `
		SELECT key, value, help, type, options, group_id, flag, index_order, created_at, updated_at
		FROM settings WHERE group_id = ? ORDER BY index_order
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*models.SettingItem
	for rows.Next() {
		setting := &models.SettingItem{}
		err := rows.Scan(
			&setting.Key, &setting.Value, &setting.Help, &setting.Type,
			&setting.Options, &setting.GroupID, &setting.Flag, &setting.IndexOrder,
			&setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}
	return settings, nil
}

// SetValue 设置配置值
func (r *settingRepository) SetValue(ctx context.Context, key, value string) error {
	query := `
		UPDATE settings SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE key = ?
	`
	_, err := r.db.ExecContext(ctx, query, value, key)
	return err
}

// Update 更新设置项
func (r *settingRepository) Update(ctx context.Context, setting *models.SettingItem) error {
	query := `
		UPDATE settings SET value = ?, help = ?, type = ?, options = ?, group_id = ?, 
		       flag = ?, index_order = ?, updated_at = CURRENT_TIMESTAMP
		WHERE key = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		setting.Value, setting.Help, setting.Type, setting.Options,
		setting.GroupID, setting.Flag, setting.IndexOrder, setting.Key,
	)
	return err
}

// Delete 删除设置项
func (r *settingRepository) Delete(ctx context.Context, id int) error {
	// 由于这个方法签名限制，我们使用index_order来删除
	query := `DELETE FROM settings WHERE index_order = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List 获取设置项列表
func (r *settingRepository) List(ctx context.Context, limit, offset int) ([]*models.SettingItem, error) {
	query := `
		SELECT key, value, help, type, options, group_id, flag, index_order, created_at, updated_at
		FROM settings ORDER BY group_id, index_order LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*models.SettingItem
	for rows.Next() {
		setting := &models.SettingItem{}
		err := rows.Scan(
			&setting.Key, &setting.Value, &setting.Help, &setting.Type,
			&setting.Options, &setting.GroupID, &setting.Flag, &setting.IndexOrder,
			&setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}
	return settings, nil
} 