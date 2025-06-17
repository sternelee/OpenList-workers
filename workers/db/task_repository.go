package db

import (
	"context"
	"database/sql"

	"github.com/sternelee/OpenList-workers/workers/models"
)

type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository 创建任务仓库
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create 创建任务
func (r *taskRepository) Create(ctx context.Context, task *models.TaskItem) error {
	query := `
		INSERT INTO tasks (key, persist_data)
		VALUES (?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, task.Key, task.PersistData)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = int(id)
	return nil
}

// GetByID 根据ID获取任务
func (r *taskRepository) GetByID(ctx context.Context, id int) (*models.TaskItem, error) {
	query := `
		SELECT id, key, persist_data, created_at, updated_at
		FROM tasks WHERE id = ?
	`
	task := &models.TaskItem{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Key, &task.PersistData, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetByKey 根据Key获取任务
func (r *taskRepository) GetByKey(ctx context.Context, key string) (*models.TaskItem, error) {
	query := `
		SELECT id, key, persist_data, created_at, updated_at
		FROM tasks WHERE key = ?
	`
	task := &models.TaskItem{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&task.ID, &task.Key, &task.PersistData, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// Update 更新任务
func (r *taskRepository) Update(ctx context.Context, task *models.TaskItem) error {
	query := `
		UPDATE tasks SET key = ?, persist_data = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, task.Key, task.PersistData, task.ID)
	return err
}

// Delete 删除任务
func (r *taskRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List 获取任务列表
func (r *taskRepository) List(ctx context.Context, limit, offset int) ([]*models.TaskItem, error) {
	query := `
		SELECT id, key, persist_data, created_at, updated_at
		FROM tasks ORDER BY id LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.TaskItem
	for rows.Next() {
		task := &models.TaskItem{}
		err := rows.Scan(
			&task.ID, &task.Key, &task.PersistData, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
} 