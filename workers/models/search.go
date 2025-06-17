package models

import (
	"fmt"
	"time"
)

// SearchNode 搜索节点
type SearchNode struct {
	ID       int       `json:"id" db:"id"`
	Parent   string    `json:"parent" db:"parent"`
	Name     string    `json:"name" db:"name"`
	IsDir    bool      `json:"is_dir" db:"is_dir"`
	Size     int64     `json:"size" db:"size"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Type 返回类型名称
func (s *SearchNode) Type() string {
	return "SearchNode"
}

// SearchReq 搜索请求
type SearchReq struct {
	Parent   string `json:"parent"`
	Keywords string `json:"keywords"`
	// 0 for all, 1 for dir, 2 for file
	Scope int `json:"scope"`
	PageReq
}

// PageReq 分页请求
type PageReq struct {
	Page    int `json:"page" form:"page"`
	PerPage int `json:"per_page" form:"per_page"`
}

// 常量定义
const MaxUint = ^uint(0)
const MinUint = 0
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

// Validate 验证搜索请求
func (p *SearchReq) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page can't < 1")
	}
	if p.PerPage < 1 {
		return fmt.Errorf("per_page can't < 1")
	}
	return nil
}

// Validate 验证分页请求
func (p *PageReq) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = MaxInt
	}
}

// IndexProgress 索引进度
type IndexProgress struct {
	ObjCount     uint64     `json:"obj_count"`
	IsDone       bool       `json:"is_done"`
	LastDoneTime *time.Time `json:"last_done_time"`
	Error        string     `json:"error"`
} 