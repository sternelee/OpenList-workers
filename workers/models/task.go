package models

import "time"

// TaskItem 任务项
type TaskItem struct {
	ID          int       `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	PersistData string    `json:"persist_data" db:"persist_data"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
} 