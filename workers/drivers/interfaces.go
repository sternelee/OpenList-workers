package drivers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/sternelee/OpenList-workers/workers/models"
)

// Driver 驱动主接口
type Driver interface {
	Config() DriverConfig
	GetStorage() *models.Storage
	SetStorage(storage *models.Storage)
	GetAddition() interface{}
	Init(ctx context.Context) error
	Drop(ctx context.Context) error

	// 基础读取操作
	List(ctx context.Context, dir Obj, args ListArgs) ([]Obj, error)
	Link(ctx context.Context, file Obj, args LinkArgs) (*Link, error)
}

// Writer 写入操作接口
type Writer interface {
	MakeDir(ctx context.Context, parentDir Obj, dirName string) error
	Move(ctx context.Context, srcObj, dstDir Obj) error
	Rename(ctx context.Context, srcObj Obj, newName string) error
	Copy(ctx context.Context, srcObj, dstDir Obj) error
	Remove(ctx context.Context, obj Obj) error
	Put(ctx context.Context, dstDir Obj, file FileStreamer, up UpdateProgress) error
}

// Getter 获取文件接口
type Getter interface {
	Get(ctx context.Context, path string) (Obj, error)
}

// Other 其他操作接口
type Other interface {
	Other(ctx context.Context, args OtherArgs) (interface{}, error)
}

// DriverConfig 驱动配置
type DriverConfig struct {
	Name              string `json:"name"`
	LocalSort         bool   `json:"local_sort"`
	OnlyLocal         bool   `json:"only_local"`
	OnlyProxy         bool   `json:"only_proxy"`
	NoCache           bool   `json:"no_cache"`
	NoUpload          bool   `json:"no_upload"`
	NeedMs            bool   `json:"need_ms"`
	DefaultRoot       string `json:"default_root"`
	CheckStatus       bool   `json:"check_status"`
	Alert             string `json:"alert"`
	NoOverwriteUpload bool   `json:"no_overwrite_upload"`
}

// Obj 文件对象接口
type Obj interface {
	GetSize() int64
	GetName() string
	ModTime() time.Time
	CreateTime() time.Time
	IsDir() bool
	GetID() string
	GetPath() string
}

// FileStreamer 文件流接口
type FileStreamer interface {
	io.Reader
	io.Closer
	Obj
	GetMimetype() string
	NeedStore() bool
	IsForceStreamUpload() bool
	GetExist() Obj
	SetExist(Obj)
}

// Link 链接信息
type Link struct {
	URL         string        `json:"url"`
	Header      http.Header   `json:"header"`
	MFile       io.ReadCloser `json:"-"`
	Concurrency int           `json:"concurrency"`
	PartSize    int           `json:"part_size"`
}

// ListArgs 列表参数
type ListArgs struct {
	ReqPath           string
	S3ShowPlaceholder bool
	Refresh           bool
}

// LinkArgs 链接参数
type LinkArgs struct {
	IP       string
	Header   http.Header
	Type     string
	HttpReq  *http.Request
	Redirect bool
}

// OtherArgs 其他操作参数
type OtherArgs struct {
	Obj    Obj
	Method string
	Data   interface{}
}

// UpdateProgress 更新进度函数
type UpdateProgress func(percentage float64)

// Object 基础对象实现
type Object struct {
	ID       string
	Path     string
	Name     string
	Size     int64
	Modified time.Time
	Ctime    time.Time
	IsFolder bool
}

func (o *Object) GetSize() int64     { return o.Size }
func (o *Object) GetName() string    { return o.Name }
func (o *Object) ModTime() time.Time { return o.Modified }
func (o *Object) CreateTime() time.Time {
	if o.Ctime.IsZero() {
		return o.ModTime()
	}
	return o.Ctime
}
func (o *Object) IsDir() bool     { return o.IsFolder }
func (o *Object) GetID() string   { return o.ID }
func (o *Object) GetPath() string { return o.Path }

// BaseDriver 基础驱动实现
type BaseDriver struct {
	storage *models.Storage
}

func (d *BaseDriver) GetStorage() *models.Storage {
	return d.storage
}

func (d *BaseDriver) SetStorage(storage *models.Storage) {
	d.storage = storage
}

// Additional 附加配置接口
type Additional interface{}

// RootPath 根路径配置
type RootPath struct {
	RootFolderPath string `json:"root_folder_path" required:"true"`
}

func (r *RootPath) GetRootPath() string {
	return r.RootFolderPath
}

// RootID 根ID配置
type RootID struct {
	RootFolderID string `json:"root_folder_id" required:"true"`
}

func (r *RootID) GetRootID() string {
	return r.RootFolderID
}

