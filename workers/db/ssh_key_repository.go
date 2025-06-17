package db

import (
	"context"
	"database/sql"

	"github.com/sternelee/OpenList-workers/workers/models"
)

type sshKeyRepository struct {
	db *sql.DB
}

// NewSSHKeyRepository 创建SSH密钥仓库
func NewSSHKeyRepository(db *sql.DB) SSHKeyRepository {
	return &sshKeyRepository{db: db}
}

// Create 创建SSH密钥
func (r *sshKeyRepository) Create(ctx context.Context, key *models.SSHPublicKey) error {
	query := `
		INSERT INTO ssh_keys (user_id, title, fingerprint, key_str, added_time, last_used_time)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		key.UserID, key.Title, key.Fingerprint, key.KeyStr,
		key.AddedTime, key.LastUsedTime,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	key.ID = int(id)
	return nil
}

// GetByID 根据ID获取SSH密钥
func (r *sshKeyRepository) GetByID(ctx context.Context, id int) (*models.SSHPublicKey, error) {
	query := `
		SELECT id, user_id, title, fingerprint, key_str, added_time, last_used_time, created_at, updated_at
		FROM ssh_keys WHERE id = ?
	`
	key := &models.SSHPublicKey{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&key.ID, &key.UserID, &key.Title, &key.Fingerprint, &key.KeyStr,
		&key.AddedTime, &key.LastUsedTime, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GetByUserID 根据用户ID获取SSH密钥列表
func (r *sshKeyRepository) GetByUserID(ctx context.Context, userID int) ([]*models.SSHPublicKey, error) {
	query := `
		SELECT id, user_id, title, fingerprint, key_str, added_time, last_used_time, created_at, updated_at
		FROM ssh_keys WHERE user_id = ? ORDER BY added_time DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.SSHPublicKey
	for rows.Next() {
		key := &models.SSHPublicKey{}
		err := rows.Scan(
			&key.ID, &key.UserID, &key.Title, &key.Fingerprint, &key.KeyStr,
			&key.AddedTime, &key.LastUsedTime, &key.CreatedAt, &key.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// GetByFingerprint 根据指纹获取SSH密钥
func (r *sshKeyRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*models.SSHPublicKey, error) {
	query := `
		SELECT id, user_id, title, fingerprint, key_str, added_time, last_used_time, created_at, updated_at
		FROM ssh_keys WHERE fingerprint = ?
	`
	key := &models.SSHPublicKey{}
	err := r.db.QueryRowContext(ctx, query, fingerprint).Scan(
		&key.ID, &key.UserID, &key.Title, &key.Fingerprint, &key.KeyStr,
		&key.AddedTime, &key.LastUsedTime, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DeleteByUserID 删除用户的所有SSH密钥
func (r *sshKeyRepository) DeleteByUserID(ctx context.Context, userID int) error {
	query := `DELETE FROM ssh_keys WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// Update 更新SSH密钥
func (r *sshKeyRepository) Update(ctx context.Context, key *models.SSHPublicKey) error {
	query := `
		UPDATE ssh_keys SET user_id = ?, title = ?, fingerprint = ?, key_str = ?, 
		       added_time = ?, last_used_time = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		key.UserID, key.Title, key.Fingerprint, key.KeyStr,
		key.AddedTime, key.LastUsedTime, key.ID,
	)
	return err
}

// Delete 删除SSH密钥
func (r *sshKeyRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM ssh_keys WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List 获取SSH密钥列表
func (r *sshKeyRepository) List(ctx context.Context, limit, offset int) ([]*models.SSHPublicKey, error) {
	query := `
		SELECT id, user_id, title, fingerprint, key_str, added_time, last_used_time, created_at, updated_at
		FROM ssh_keys ORDER BY added_time DESC LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.SSHPublicKey
	for rows.Next() {
		key := &models.SSHPublicKey{}
		err := rows.Scan(
			&key.ID, &key.UserID, &key.Title, &key.Fingerprint, &key.KeyStr,
			&key.AddedTime, &key.LastUsedTime, &key.CreatedAt, &key.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
} 