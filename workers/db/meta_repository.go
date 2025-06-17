package db

import (
	"context"
	"database/sql"

	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

type metaRepository struct {
	db *sql.DB
}

// NewMetaRepository 创建元数据仓库
func NewMetaRepository(db *sql.DB) MetaRepository {
	return &metaRepository{db: db}
}

// Create 创建元数据
func (r *metaRepository) Create(ctx context.Context, meta *models.Meta) error {
	query := `
		INSERT INTO metas (path, password, p_sub, write, w_sub, hide, h_sub, readme, r_sub, header, header_sub)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		meta.Path, meta.Password, meta.PSub, meta.Write, meta.WSub,
		meta.Hide, meta.HSub, meta.Readme, meta.RSub, meta.Header, meta.HeaderSub,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	meta.ID = int(id)
	return nil
}

// GetByID 根据ID获取元数据
func (r *metaRepository) GetByID(ctx context.Context, id int) (*models.Meta, error) {
	query := `
		SELECT id, path, password, p_sub, write, w_sub, hide, h_sub, readme, r_sub, 
		       header, header_sub, created_at, updated_at
		FROM metas WHERE id = ?
	`
	meta := &models.Meta{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&meta.ID, &meta.Path, &meta.Password, &meta.PSub, &meta.Write,
		&meta.WSub, &meta.Hide, &meta.HSub, &meta.Readme, &meta.RSub,
		&meta.Header, &meta.HeaderSub, &meta.CreatedAt, &meta.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// GetByPath 根据路径获取元数据
func (r *metaRepository) GetByPath(ctx context.Context, path string) (*models.Meta, error) {
	query := `
		SELECT id, path, password, p_sub, write, w_sub, hide, h_sub, readme, r_sub, 
		       header, header_sub, created_at, updated_at
		FROM metas WHERE path = ?
	`
	meta := &models.Meta{}
	err := r.db.QueryRowContext(ctx, query, path).Scan(
		&meta.ID, &meta.Path, &meta.Password, &meta.PSub, &meta.Write,
		&meta.WSub, &meta.Hide, &meta.HSub, &meta.Readme, &meta.RSub,
		&meta.Header, &meta.HeaderSub, &meta.CreatedAt, &meta.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// Update 更新元数据
func (r *metaRepository) Update(ctx context.Context, meta *models.Meta) error {
	query := `
		UPDATE metas SET path = ?, password = ?, p_sub = ?, write = ?, w_sub = ?, 
		       hide = ?, h_sub = ?, readme = ?, r_sub = ?, header = ?, header_sub = ?,
		       updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		meta.Path, meta.Password, meta.PSub, meta.Write, meta.WSub,
		meta.Hide, meta.HSub, meta.Readme, meta.RSub, meta.Header, meta.HeaderSub,
		meta.ID,
	)
	return err
}

// Delete 删除元数据
func (r *metaRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM metas WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List 获取元数据列表
func (r *metaRepository) List(ctx context.Context, limit, offset int) ([]*models.Meta, error) {
	query := `
		SELECT id, path, password, p_sub, write, w_sub, hide, h_sub, readme, r_sub, 
		       header, header_sub, created_at, updated_at
		FROM metas ORDER BY id LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metas []*models.Meta
	for rows.Next() {
		meta := &models.Meta{}
		err := rows.Scan(
			&meta.ID, &meta.Path, &meta.Password, &meta.PSub, &meta.Write,
			&meta.WSub, &meta.Hide, &meta.HSub, &meta.Readme, &meta.RSub,
			&meta.Header, &meta.HeaderSub, &meta.CreatedAt, &meta.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		metas = append(metas, meta)
	}
	return metas, nil
} 