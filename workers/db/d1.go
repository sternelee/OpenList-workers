package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sternelee/OpenList-workers/workers/models"
	"github.com/syumai/workers/cloudflare"
)

type D1Client struct {
	db cloudflare.D1Database
}

func NewD1Client(db cloudflare.D1Database) *D1Client {
	return &D1Client{db: db}
}

// User operations

func (c *D1Client) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	stmt := c.db.Prepare("SELECT id, username, password_hash, salt, role, permission, base_path, disabled, otp_secret, sso_id, authn, created_at, updated_at FROM users WHERE username = ?")
	result, err := stmt.Bind(username).First()
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("user not found")
	}

	user := &models.User{}
	if err := scanUser(result, user); err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return user, nil
}

func (c *D1Client) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	stmt := c.db.Prepare("SELECT id, username, password_hash, salt, role, permission, base_path, disabled, otp_secret, sso_id, authn, created_at, updated_at FROM users WHERE id = ?")
	result, err := stmt.Bind(id).First()
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("user not found")
	}

	user := &models.User{}
	if err := scanUser(result, user); err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return user, nil
}

func (c *D1Client) CreateUser(ctx context.Context, user *models.User) error {
	stmt := c.db.Prepare(`
		INSERT INTO users (username, password_hash, salt, role, permission, base_path, disabled, otp_secret, sso_id, authn, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := stmt.Bind(
		user.Username,
		user.PasswordHash,
		user.Salt,
		user.Role,
		user.Permission,
		user.BasePath,
		user.Disabled,
		user.OtpSecret,
		user.SsoID,
		user.Authn,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	).Run()
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (c *D1Client) UpdateUser(ctx context.Context, user *models.User) error {
	stmt := c.db.Prepare(`
		UPDATE users 
		SET username = ?, password_hash = ?, salt = ?, role = ?, permission = ?, base_path = ?, disabled = ?, otp_secret = ?, sso_id = ?, authn = ?, updated_at = ?
		WHERE id = ?
	`)
	
	user.UpdatedAt = time.Now()

	_, err := stmt.Bind(
		user.Username,
		user.PasswordHash,
		user.Salt,
		user.Role,
		user.Permission,
		user.BasePath,
		user.Disabled,
		user.OtpSecret,
		user.SsoID,
		user.Authn,
		user.UpdatedAt.Format(time.RFC3339),
		user.ID,
	).Run()
	
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (c *D1Client) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	stmt := c.db.Prepare("SELECT id, username, password_hash, salt, role, permission, base_path, disabled, otp_secret, sso_id, authn, created_at, updated_at FROM users ORDER BY id LIMIT ? OFFSET ?")
	result, err := stmt.Bind(limit, offset).All()
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	var users []*models.User
	for _, row := range result {
		user := &models.User{}
		if err := scanUser(row, user); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Storage operations

func (c *D1Client) GetStorageByMountPath(ctx context.Context, mountPath string) (*models.Storage, error) {
	stmt := c.db.Prepare("SELECT id, mount_path, name, driver, config, enabled, created_at, updated_at FROM storages WHERE mount_path = ?")
	result, err := stmt.Bind(mountPath).First()
	if err != nil {
		return nil, fmt.Errorf("failed to query storage: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("storage not found")
	}

	storage := &models.Storage{}
	if err := scanStorage(result, storage); err != nil {
		return nil, fmt.Errorf("failed to scan storage: %w", err)
	}

	return storage, nil
}

func (c *D1Client) ListStorages(ctx context.Context) ([]*models.Storage, error) {
	stmt := c.db.Prepare("SELECT id, mount_path, name, driver, config, enabled, created_at, updated_at FROM storages ORDER BY mount_path")
	result, err := stmt.All()
	if err != nil {
		return nil, fmt.Errorf("failed to query storages: %w", err)
	}

	var storages []*models.Storage
	for _, row := range result {
		storage := &models.Storage{}
		if err := scanStorage(row, storage); err != nil {
			return nil, fmt.Errorf("failed to scan storage: %w", err)
		}
		storages = append(storages, storage)
	}

	return storages, nil
}

// Setting operations

func (c *D1Client) GetSetting(ctx context.Context, key string) (*models.Setting, error) {
	stmt := c.db.Prepare("SELECT id, key, value, description, created_at, updated_at FROM settings WHERE key = ?")
	result, err := stmt.Bind(key).First()
	if err != nil {
		return nil, fmt.Errorf("failed to query setting: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("setting not found")
	}

	setting := &models.Setting{}
	if err := scanSetting(result, setting); err != nil {
		return nil, fmt.Errorf("failed to scan setting: %w", err)
	}

	return setting, nil
}

func (c *D1Client) SetSetting(ctx context.Context, key, value, description string) error {
	stmt := c.db.Prepare(`
		INSERT OR REPLACE INTO settings (key, value, description, created_at, updated_at)
		VALUES (?, ?, ?, COALESCE((SELECT created_at FROM settings WHERE key = ?), ?), ?)
	`)
	
	now := time.Now().Format(time.RFC3339)
	_, err := stmt.Bind(key, value, description, key, now, now).Run()
	if err != nil {
		return fmt.Errorf("failed to set setting: %w", err)
	}

	return nil
}

// Helper functions

func scanUser(row map[string]interface{}, user *models.User) error {
	if id, ok := row["id"].(float64); ok {
		user.ID = int(id)
	}
	if username, ok := row["username"].(string); ok {
		user.Username = username
	}
	if passwordHash, ok := row["password_hash"].(string); ok {
		user.PasswordHash = passwordHash
	}
	if salt, ok := row["salt"].(string); ok {
		user.Salt = salt
	}
	if role, ok := row["role"].(float64); ok {
		user.Role = int(role)
	}
	if permission, ok := row["permission"].(float64); ok {
		user.Permission = int32(permission)
	}
	if basePath, ok := row["base_path"].(string); ok {
		user.BasePath = basePath
	}
	if disabled, ok := row["disabled"].(float64); ok {
		user.Disabled = disabled != 0
	}
	if otpSecret, ok := row["otp_secret"].(string); ok {
		user.OtpSecret = otpSecret
	}
	if ssoID, ok := row["sso_id"].(string); ok {
		user.SsoID = ssoID
	}
	if authn, ok := row["authn"].(string); ok {
		user.Authn = authn
	}
	if createdAt, ok := row["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			user.CreatedAt = t
		}
	}
	if updatedAt, ok := row["updated_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			user.UpdatedAt = t
		}
	}
	return nil
}

func scanStorage(row map[string]interface{}, storage *models.Storage) error {
	if id, ok := row["id"].(float64); ok {
		storage.ID = int(id)
	}
	if mountPath, ok := row["mount_path"].(string); ok {
		storage.MountPath = mountPath
	}
	if name, ok := row["name"].(string); ok {
		storage.Name = name
	}
	if driver, ok := row["driver"].(string); ok {
		storage.Driver = driver
	}
	if config, ok := row["config"].(string); ok {
		storage.Config = config
	}
	if enabled, ok := row["enabled"].(float64); ok {
		storage.Enabled = enabled != 0
	}
	if createdAt, ok := row["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			storage.CreatedAt = t
		}
	}
	if updatedAt, ok := row["updated_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			storage.UpdatedAt = t
		}
	}
	return nil
}

func scanSetting(row map[string]interface{}, setting *models.Setting) error {
	if id, ok := row["id"].(float64); ok {
		setting.ID = int(id)
	}
	if key, ok := row["key"].(string); ok {
		setting.Key = key
	}
	if value, ok := row["value"].(string); ok {
		setting.Value = value
	}
	if description, ok := row["description"].(string); ok {
		setting.Description = description
	}
	if createdAt, ok := row["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			setting.CreatedAt = t
		}
	}
	if updatedAt, ok := row["updated_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			setting.UpdatedAt = t
		}
	}
	return nil
} 