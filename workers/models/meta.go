package models

import "time"

// Meta 路径元数据
type Meta struct {
	ID        int       `json:"id" db:"id"`
	Path      string    `json:"path" db:"path" binding:"required"`
	Password  string    `json:"password" db:"password"`
	PSub      bool      `json:"p_sub" db:"p_sub"`
	Write     bool      `json:"write" db:"write"`
	WSub      bool      `json:"w_sub" db:"w_sub"`
	Hide      string    `json:"hide" db:"hide"`
	HSub      bool      `json:"h_sub" db:"h_sub"`
	Readme    string    `json:"readme" db:"readme"`
	RSub      bool      `json:"r_sub" db:"r_sub"`
	Header    string    `json:"header" db:"header"`
	HeaderSub bool      `json:"header_sub" db:"header_sub"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
} 