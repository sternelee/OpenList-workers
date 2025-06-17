package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

type searchNodeRepository struct {
	db *sql.DB
}

// NewSearchNodeRepository 创建搜索节点仓库
func NewSearchNodeRepository(db *sql.DB) SearchNodeRepository {
	return &searchNodeRepository{db: db}
}

// Create 创建搜索节点
func (r *searchNodeRepository) Create(ctx context.Context, node *models.SearchNode) error {
	query := `
		INSERT INTO search_nodes (parent, name, is_dir, size)
		VALUES (?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		node.Parent, node.Name, node.IsDir, node.Size,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	node.ID = int(id)
	return nil
}

// GetByID 根据ID获取搜索节点
func (r *searchNodeRepository) GetByID(ctx context.Context, id int) (*models.SearchNode, error) {
	query := `
		SELECT id, parent, name, is_dir, size, created_at, updated_at
		FROM search_nodes WHERE id = ?
	`
	node := &models.SearchNode{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&node.ID, &node.Parent, &node.Name, &node.IsDir, &node.Size,
		&node.CreatedAt, &node.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// GetByParent 根据父路径获取搜索节点列表
func (r *searchNodeRepository) GetByParent(ctx context.Context, parent string) ([]*models.SearchNode, error) {
	query := `
		SELECT id, parent, name, is_dir, size, created_at, updated_at
		FROM search_nodes WHERE parent = ? ORDER BY is_dir DESC, name
	`
	rows, err := r.db.QueryContext(ctx, query, parent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.SearchNode
	for rows.Next() {
		node := &models.SearchNode{}
		err := rows.Scan(
			&node.ID, &node.Parent, &node.Name, &node.IsDir, &node.Size,
			&node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// Search 搜索节点
func (r *searchNodeRepository) Search(ctx context.Context, req *models.SearchReq) ([]*models.SearchNode, int64, error) {
	// 构建搜索条件
	whereConditions := []string{}
	args := []interface{}{}

	if req.Parent != "" {
		whereConditions = append(whereConditions, "parent LIKE ?")
		args = append(args, req.Parent+"%")
	}

	if req.Keywords != "" {
		whereConditions = append(whereConditions, "name LIKE ?")
		args = append(args, "%"+req.Keywords+"%")
	}

	// 根据 scope 过滤
	if req.Scope == 1 { // 只搜索目录
		whereConditions = append(whereConditions, "is_dir = TRUE")
	} else if req.Scope == 2 { // 只搜索文件
		whereConditions = append(whereConditions, "is_dir = FALSE")
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM search_nodes %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取搜索结果
	query := fmt.Sprintf(`
		SELECT id, parent, name, is_dir, size, created_at, updated_at
		FROM search_nodes %s
		ORDER BY is_dir DESC, name
		LIMIT ? OFFSET ?
	`, whereClause)

	// 添加分页参数
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var nodes []*models.SearchNode
	for rows.Next() {
		node := &models.SearchNode{}
		err := rows.Scan(
			&node.ID, &node.Parent, &node.Name, &node.IsDir, &node.Size,
			&node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		nodes = append(nodes, node)
	}

	return nodes, total, nil
}

// DeleteByParent 删除指定父路径下的所有节点
func (r *searchNodeRepository) DeleteByParent(ctx context.Context, parent string) error {
	query := `DELETE FROM search_nodes WHERE parent LIKE ?`
	_, err := r.db.ExecContext(ctx, query, parent+"%")
	return err
}

// Update 更新搜索节点
func (r *searchNodeRepository) Update(ctx context.Context, node *models.SearchNode) error {
	query := `
		UPDATE search_nodes SET parent = ?, name = ?, is_dir = ?, size = ?,
		       updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		node.Parent, node.Name, node.IsDir, node.Size, node.ID,
	)
	return err
}

// Delete 删除搜索节点
func (r *searchNodeRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM search_nodes WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List 获取搜索节点列表
func (r *searchNodeRepository) List(ctx context.Context, limit, offset int) ([]*models.SearchNode, error) {
	query := `
		SELECT id, parent, name, is_dir, size, created_at, updated_at
		FROM search_nodes ORDER BY parent, is_dir DESC, name LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.SearchNode
	for rows.Next() {
		node := &models.SearchNode{}
		err := rows.Scan(
			&node.ID, &node.Parent, &node.Name, &node.IsDir, &node.Size,
			&node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
} 