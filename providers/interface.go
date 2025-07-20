package providers

import (
	"context"
	"os"
	"time"
)

// Provider 定义挂载提供者接口
type Provider interface {
	// Mount 挂载文件系统到指定路径
	Mount(ctx context.Context, mountPoint string) error

	// Unmount 卸载文件系统
	Unmount() error

	// IsSupported 检查当前平台是否支持此挂载方式
	IsSupported() bool

	// GetName 获取挂载提供者名称
	GetName() string
}

// FileSystemHandler 定义文件系统操作接口
type FileSystemHandler interface {
	// 文件操作
	OpenFile(path string, mode int) (ReadCloser, error)
	CreateFile(path string, mode os.FileMode) (WriteCloser, error)
	DeleteFile(path string) error
	
	// 目录操作
	ListFiles(path string) ([]*FileInfo, error)
	CreateDir(path string, mode os.FileMode) error
	DeleteDir(path string) error

	// 文件信息
	GetFileInfo(path string) (*FileInfo, error)
	Exists(path string) bool
	
	// 文件拷贝
	CopyFile(dst WriteCloser, src ReadCloser) error
}

// ReadCloser 读取器接口
type ReadCloser interface {
	Read([]byte) (int, error)
	Close() error
}

// WriteCloser 写入器接口
type WriteCloser interface {
	Write([]byte) (int, error)
	Close() error
}

// FileInfo 文件信息结构
type FileInfo struct {
	Name    string
	Path    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
	ETag    string
}

// Config 挂载配置
type Config struct {
	MountPoint string
	MountType  string
	ReadOnly   bool
	CacheDir   string
	CacheSize  int64
	DebugMode  bool

	// macOS File Provider 配置
	BundleID    string
	DisplayName string

	// Windows Cloud Files 配置
	ProviderName string
	SyncRootPath string

	// rclone SDK 配置
	RcloneConfig string
}

// Platform 平台信息
type Platform struct {
	OS      string
	Version string
	Arch    string
}

// GetPlatform 获取当前平台信息
func GetPlatform() Platform {
	return Platform{
		OS:      "darwin", // 实际实现中会动态检测
		Version: "14.0",
		Arch:    "amd64",
	}
}

// Manager 挂载管理器
type Manager struct {
	config   *Config
	handler  FileSystemHandler
	provider Provider
	platform PlatformInfo
}

// NewManager 创建挂载管理器
func NewManager(config *Config, handler FileSystemHandler) *Manager {
	return &Manager{
		config:  config,
		handler: handler,
		platform: GetPlatformInfo(),
	}
}

// SelectBestProvider 选择最佳挂载提供者
func (mm *Manager) SelectBestProvider() Provider {
	providers := []Provider{
		NewRcloneSDKProvider(mm.config, mm.handler),
	}

	for _, provider := range providers {
		if provider.IsSupported() {
			return provider
		}
	}

	return NewRcloneSDKProvider(mm.config, mm.handler)
}

// Mount 执行挂载
func (mm *Manager) Mount(ctx context.Context) error {
	if mm.provider == nil {
		mm.provider = mm.SelectBestProvider()
	}

	return mm.provider.Mount(ctx, mm.config.MountPoint)
}

// Unmount 执行卸载
func (mm *Manager) Unmount() error {
	if mm.provider != nil {
		return mm.provider.Unmount()
	}
	return nil
}
