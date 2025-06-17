package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/sternelee/OpenList-workers/workers/models"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, pwd_hash, pwd_ts, salt, role, permission, base_path, disabled, otp_secret, sso_id, authn)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Username, user.PwdHash, user.PwdTS, user.Salt, user.Role,
		user.Permission, user.BasePath, user.Disabled, user.OtpSecret,
		user.SsoID, user.Authn,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, username, pwd_hash, pwd_ts, salt, role, permission, base_path, 
		       disabled, otp_secret, sso_id, authn, created_at, updated_at
		FROM users WHERE id = ?
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PwdHash, &user.PwdTS, &user.Salt,
		&user.Role, &user.Permission, &user.BasePath, &user.Disabled,
		&user.OtpSecret, &user.SsoID, &user.Authn, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, pwd_hash, pwd_ts, salt, role, permission, base_path, 
		       disabled, otp_secret, sso_id, authn, created_at, updated_at
		FROM users WHERE username = ?
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PwdHash, &user.PwdTS, &user.Salt,
		&user.Role, &user.Permission, &user.BasePath, &user.Disabled,
		&user.OtpSecret, &user.SsoID, &user.Authn, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetBySsoID 根据SSO ID获取用户
func (r *userRepository) GetBySsoID(ctx context.Context, ssoID string) (*models.User, error) {
	query := `
		SELECT id, username, pwd_hash, pwd_ts, salt, role, permission, base_path, 
		       disabled, otp_secret, sso_id, authn, created_at, updated_at
		FROM users WHERE sso_id = ?
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, ssoID).Scan(
		&user.ID, &user.Username, &user.PwdHash, &user.PwdTS, &user.Salt,
		&user.Role, &user.Permission, &user.BasePath, &user.Disabled,
		&user.OtpSecret, &user.SsoID, &user.Authn, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET username = ?, pwd_hash = ?, pwd_ts = ?, salt = ?, role = ?, 
		       permission = ?, base_path = ?, disabled = ?, otp_secret = ?, 
		       sso_id = ?, authn = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Username, user.PwdHash, user.PwdTS, user.Salt, user.Role,
		user.Permission, user.BasePath, user.Disabled, user.OtpSecret,
		user.SsoID, user.Authn, user.ID,
	)
	return err
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, username, pwd_hash, pwd_ts, salt, role, permission, base_path, 
		       disabled, otp_secret, sso_id, authn, created_at, updated_at
		FROM users ORDER BY id LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.PwdHash, &user.PwdTS, &user.Salt,
			&user.Role, &user.Permission, &user.BasePath, &user.Disabled,
			&user.OtpSecret, &user.SsoID, &user.Authn, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
} 