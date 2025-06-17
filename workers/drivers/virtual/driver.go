package virtual

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sternelee/OpenList-workers/workers/drivers"
)

// Virtual 虚拟驱动
type Virtual struct {
	drivers.BaseDriver
	Addition
}

// Addition 虚拟驱动配置
type Addition struct {
	drivers.RootPath
	Files string `json:"files" type:"text" help:"JSON format file list"`
}

// Config 驱动配置
var config = drivers.DriverConfig{
	Name:        "Virtual",
	LocalSort:   true,
	OnlyLocal:   true,
	NoCache:     true,
	DefaultRoot: "/",
}

// VirtualFile 虚拟文件
type VirtualFile struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"is_dir"`
	Modified string `json:"modified"`
	Content  string `json:"content,omitempty"`
}

func (d *Virtual) Config() drivers.DriverConfig {
	return config
}

func (d *Virtual) GetAddition() interface{} {
	return &d.Addition
}

func (d *Virtual) Init(ctx context.Context) error {
	return nil
}

func (d *Virtual) Drop(ctx context.Context) error {
	return nil
}

func (d *Virtual) List(ctx context.Context, dir drivers.Obj, args drivers.ListArgs) ([]drivers.Obj, error) {
	var files []VirtualFile
	if d.Files != "" {
		if err := json.Unmarshal([]byte(d.Files), &files); err != nil {
			return nil, err
		}
	}

	var objs []drivers.Obj
	for _, file := range files {
		modTime, _ := time.Parse("2006-01-02 15:04:05", file.Modified)
		if modTime.IsZero() {
			modTime = time.Now()
		}

		obj := &drivers.Object{
			ID:       file.Name,
			Path:     dir.GetPath() + "/" + file.Name,
			Name:     file.Name,
			Size:     file.Size,
			Modified: modTime,
			IsFolder: file.IsDir,
		}
		objs = append(objs, obj)
	}

	return objs, nil
}

func (d *Virtual) Link(ctx context.Context, file drivers.Obj, args drivers.LinkArgs) (*drivers.Link, error) {
	// 虚拟驱动返回空链接
	return &drivers.Link{
		URL: "",
	}, nil
}

func (d *Virtual) Get(ctx context.Context, path string) (drivers.Obj, error) {
	// 简单实现：返回路径信息
	obj := &drivers.Object{
		ID:       path,
		Path:     path,
		Name:     path,
		Size:     0,
		Modified: time.Now(),
		IsFolder: true,
	}
	return obj, nil
}

// 实现Writer接口的方法
func (d *Virtual) MakeDir(ctx context.Context, parentDir drivers.Obj, dirName string) error {
	return nil // 虚拟操作，直接返回成功
}

func (d *Virtual) Move(ctx context.Context, srcObj, dstDir drivers.Obj) error {
	return nil
}

func (d *Virtual) Rename(ctx context.Context, srcObj drivers.Obj, newName string) error {
	return nil
}

func (d *Virtual) Copy(ctx context.Context, srcObj, dstDir drivers.Obj) error {
	return nil
}

func (d *Virtual) Remove(ctx context.Context, obj drivers.Obj) error {
	return nil
}

func (d *Virtual) Put(ctx context.Context, dstDir drivers.Obj, file drivers.FileStreamer, up drivers.UpdateProgress) error {
	return nil
}

// 确保实现了所有接口
var _ drivers.Driver = (*Virtual)(nil)
var _ drivers.Writer = (*Virtual)(nil)
var _ drivers.Getter = (*Virtual)(nil)

// 注册驱动
func init() {
	drivers.RegisterDriver(func() drivers.Driver {
		return &Virtual{}
	})
}

