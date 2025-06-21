//go:build workers
// +build workers

package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OpenListTeam/OpenList/internal/model"
	"github.com/syumai/workers/cloudflare/d1"
)

// D1DatabaseManager D1 数据库管理器
type D1DatabaseManager struct {
	db *sql.DB
}

// NewD1DatabaseManager 创建 D1 数据库管理器
func NewD1DatabaseManager(dbName string) (*D1DatabaseManager, error) {
	connector, err := d1.OpenConnector(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to open D1 connector: %w", err)
	}

	db := sql.OpenDB(connector)
	return &D1DatabaseManager{db: db}, nil
}

// CreateTables 创建数据库表
func (dm *D1DatabaseManager) CreateTables(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			pwd_hash TEXT NOT NULL,
			pwd_ts INTEGER NOT NULL,
			salt TEXT NOT NULL,
			base_path TEXT DEFAULT '/',
			role INTEGER DEFAULT 0,
			disabled BOOLEAN DEFAULT FALSE,
			permission INTEGER DEFAULT 0,
			otp_secret TEXT,
			sso_id TEXT,
			authn TEXT DEFAULT '[]'
		)`,
		`CREATE TABLE IF NOT EXISTS driver_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			display_name TEXT NOT NULL,
			description TEXT,
			config TEXT,
			icon TEXT,
			enabled BOOLEAN DEFAULT TRUE,
			order_num INTEGER DEFAULT 0,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, name),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}

	// 执行 DDL 语句创建表
	for _, query := range queries {
		_, err := dm.db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// GetUserDriverConfigs 获取用户的驱动配置列表
func (dm *D1DatabaseManager) GetUserDriverConfigs(ctx context.Context, userID uint, page, perPage int) ([]DriverConfig, int64, error) {
	// 获取总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM driver_configs WHERE user_id = ?`
	err := dm.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user driver configs: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * perPage
	query := `SELECT id, user_id, name, display_name, description, config, icon, enabled, order_num, created, modified
			  FROM driver_configs WHERE user_id = ? ORDER BY order_num, name LIMIT ? OFFSET ?`

	rows, err := dm.db.QueryContext(ctx, query, userID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user driver configs: %w", err)
	}
	defer rows.Close()

	var configs []DriverConfig
	for rows.Next() {
		var config DriverConfig
		err := rows.Scan(
			&config.ID,
			&config.UserID,
			&config.Name,
			&config.DisplayName,
			&config.Description,
			&config.Config,
			&config.Icon,
			&config.Enabled,
			&config.Order,
			&config.Created,
			&config.Modified,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan driver config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, total, nil
}

// CreateDriverConfig 创建驱动配置
func (dm *D1DatabaseManager) CreateDriverConfig(ctx context.Context, config DriverConfig) error {
	query := `INSERT INTO driver_configs (user_id, name, display_name, description, config, icon, enabled, order_num)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := dm.db.ExecContext(ctx, query,
		config.UserID,
		config.Name,
		config.DisplayName,
		config.Description,
		config.Config,
		config.Icon,
		config.Enabled,
		config.Order,
	)

	if err != nil {
		return fmt.Errorf("failed to create driver config: %w", err)
	}

	return nil
}

// UpdateDriverConfig 更新驱动配置
func (dm *D1DatabaseManager) UpdateDriverConfig(ctx context.Context, config DriverConfig) error {
	query := `UPDATE driver_configs
			  SET display_name = ?, description = ?, config = ?, icon = ?, enabled = ?, order_num = ?, modified = CURRENT_TIMESTAMP
			  WHERE id = ? AND user_id = ?`

	_, err := dm.db.ExecContext(ctx, query,
		config.DisplayName,
		config.Description,
		config.Config,
		config.Icon,
		config.Enabled,
		config.Order,
		config.ID,
		config.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update driver config: %w", err)
	}

	return nil
}

// DeleteUserDriverConfig 删除用户的驱动配置
func (dm *D1DatabaseManager) DeleteUserDriverConfig(ctx context.Context, userID, id uint) error {
	query := `DELETE FROM driver_configs WHERE id = ? AND user_id = ?`

	_, err := dm.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete driver config: %w", err)
	}

	return nil
}

// GetUsers 获取用户列表
func (dm *D1DatabaseManager) GetUsers(ctx context.Context, page, perPage int) ([]model.User, int64, error) {
	// 获取总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM users`
	err := dm.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * perPage
	query := `SELECT id, username, pwd_hash, pwd_ts, salt, base_path, role, disabled, permission, otp_secret, sso_id, authn
			  FROM users ORDER BY id LIMIT ? OFFSET ?`

	rows, err := dm.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		var otpSecret, ssoId sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PwdHash,
			&user.PwdTs,
			&user.Salt,
			&user.BasePath,
			&user.Role,
			&user.Disabled,
			&user.Permission,
			&otpSecret,
			&ssoId,
			&user.Authn,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		if otpSecret.Valid {
			user.OtpSecret = otpSecret.String
		}
		if ssoId.Valid {
			user.SsoID = ssoId.String
		}

		users = append(users, user)
	}

	return users, total, nil
}

// CreateUser 创建用户
func (dm *D1DatabaseManager) CreateUser(ctx context.Context, user model.User) error {
	query := `INSERT INTO users (username, pwd_hash, pwd_ts, salt, base_path, role, disabled, permission, otp_secret, sso_id, authn)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := dm.db.ExecContext(ctx, query,
		user.Username,
		user.PwdHash,
		user.PwdTs,
		user.Salt,
		user.BasePath,
		user.Role,
		user.Disabled,
		user.Permission,
		user.OtpSecret,
		user.SsoID,
		user.Authn,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetStorages 获取存储列表
func (dm *D1DatabaseManager) GetStorages(ctx context.Context, page, perPage int) ([]model.Storage, int64, error) {
	// 获取总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM storages`
	err := dm.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count storages: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * perPage
	query := `SELECT id, mount_path, driver, addition, order_num, remark, disabled, status, modified
			  FROM storages ORDER BY order_num, mount_path LIMIT ? OFFSET ?`

	rows, err := dm.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query storages: %w", err)
	}
	defer rows.Close()

	var storages []model.Storage
	for rows.Next() {
		var storage model.Storage
		var addition, remark sql.NullString

		err := rows.Scan(
			&storage.ID,
			&storage.MountPath,
			&storage.Driver,
			&addition,
			&storage.Order,
			&remark,
			&storage.Disabled,
			&storage.Status,
			&storage.Modified,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan storage: %w", err)
		}

		if addition.Valid {
			storage.Addition = addition.String
		}
		if remark.Valid {
			storage.Remark = remark.String
		}

		storages = append(storages, storage)
	}

	return storages, total, nil
}

// CreateStorage 创建存储
func (dm *D1DatabaseManager) CreateStorage(ctx context.Context, storage model.Storage) (uint, error) {
	query := `INSERT INTO storages (mount_path, driver, addition, order_num, remark, disabled)
			  VALUES (?, ?, ?, ?, ?, ?)`

	result, err := dm.db.ExecContext(ctx, query,
		storage.MountPath,
		storage.Driver,
		storage.Addition,
		storage.Order,
		storage.Remark,
		storage.Disabled,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create storage: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint(id), nil
}

// Close 关闭数据库连接
func (dm *D1DatabaseManager) Close() error {
	if dm.db != nil {
		return dm.db.Close()
	}
	return nil
}
