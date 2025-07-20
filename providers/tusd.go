package providers

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TusdClient interface for tusd operations
type TusdClient interface {
	UploadFile(filePath string) error
	ListFiles() ([]TusdFileInfo, error)
}

// TusdFileInfo represents file information from tusd
type TusdFileInfo struct {
	ID       string
	Name     string
	Size     int64
	ModTime  time.Time
	Metadata map[string]string
}

// TusdHandler implements FileSystemHandler interface
type TusdHandler struct {
	tusdClient TusdClient
	cacheDir   string
	logger     Logger
}

// NewTusdFileSystemHandler creates a tusd handler
func NewTusdFileSystemHandler(tusdClient TusdClient, cacheDir string, logger Logger) FileSystemHandler {
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "tusd-cache")
	}
	
	os.MkdirAll(cacheDir, 0755)
	
	return &TusdHandler{
		tusdClient: tusdClient,
		cacheDir:   cacheDir,
		logger:     logger,
	}
}

// Simple implementations for FileSystemHandler interface

// OpenFile opens a file for reading
func (h *TusdHandler) OpenFile(path string, mode int) (ReadCloser, error) {
	h.logger.Infof("OpenFile: %s", path)
	
	// For simplicity, return a file from cache or create empty
	cachedPath := filepath.Join(h.cacheDir, strings.TrimPrefix(path, "/"))
	
	file, err := os.Open(cachedPath)
	if err != nil {
		// Create empty file if not exists
		if os.IsNotExist(err) {
			dir := filepath.Dir(cachedPath)
			os.MkdirAll(dir, 0755)
			file, err = os.Create(cachedPath)
		}
	}
	
	return file, err
}

// CreateFile creates a file for writing
func (h *TusdHandler) CreateFile(path string, mode os.FileMode) (WriteCloser, error) {
	h.logger.Infof("CreateFile: %s", path)
	
	cachedPath := filepath.Join(h.cacheDir, strings.TrimPrefix(path, "/"))
	dir := filepath.Dir(cachedPath)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	return os.Create(cachedPath)
}

// DeleteFile deletes a file
func (h *TusdHandler) DeleteFile(path string) error {
	h.logger.Infof("DeleteFile: %s", path)
	
	cachedPath := filepath.Join(h.cacheDir, strings.TrimPrefix(path, "/"))
	return os.Remove(cachedPath)
}

// ListFiles lists files in a directory
func (h *TusdHandler) ListFiles(path string) ([]*FileInfo, error) {
	h.logger.Infof("ListFiles: %s", path)
	
	// For simplicity, list cached files
	cachedPath := filepath.Join(h.cacheDir, strings.TrimPrefix(path, "/"))
	
	entries, err := os.ReadDir(cachedPath)
	if err != nil {
		return nil, err
	}
	
	var files []*FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		files = append(files, &FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}
	
	return files, nil
}

// CreateDir creates a directory
func (h *TusdHandler) CreateDir(path string, mode os.FileMode) error {
	h.logger.Infof("CreateDir: %s", path)
	
	cachedPath := filepath.Join(h.cacheDir, strings.TrimPrefix(path, "/"))
	return os.MkdirAll(cachedPath, mode)
}

// DeleteDir deletes a directory
func (h *TusdHandler) DeleteDir(path string) error {
	h.logger.Infof("DeleteDir: %s", path)
	
	cachedPath := filepath.Join(h.cacheDir, strings.TrimPrefix(path, "/"))
	return os.RemoveAll(cachedPath)
}

// GetFileInfo gets file information
func (h *TusdHandler) GetFileInfo(path string) (*FileInfo, error) {
	h.logger.Infof("GetFileInfo: %s", path)
	
	cachedPath := filepath.Join(h.cacheDir, strings.TrimPrefix(path, "/"))
	info, err := os.Stat(cachedPath)
	if err != nil {
		return nil, err
	}
	
	return &FileInfo{
		Name:    info.Name(),
		Path:    path,
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// Exists checks if a file exists
func (h *TusdHandler) Exists(path string) bool {
	cachedPath := filepath.Join(h.cacheDir, strings.TrimPrefix(path, "/"))
	_, err := os.Stat(cachedPath)
	return err == nil
}

// CopyFile copies from src to dst
func (h *TusdHandler) CopyFile(dst WriteCloser, src ReadCloser) error {
	h.logger.Infof("CopyFile")
	
	_, err := io.Copy(dst, src)
	return err
}