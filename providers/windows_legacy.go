//go:build windows && (windowsxp || legacy)

package providers

import (
	"context"
	"fmt"
	"runtime"
)

// WindowsLegacyProvider provides basic file sync for Windows XP+
// This provider focuses on compatibility rather than advanced mounting features
type WindowsLegacyProvider struct {
	config  *Config
	handler FileSystemHandler
}

// NewWindowsLegacyProvider creates a new Windows legacy provider
func NewWindowsLegacyProvider(config *Config, handler FileSystemHandler) Provider {
	return &WindowsLegacyProvider{
		config:  config,
		handler: handler,
	}
}

// GetName returns the provider name
func (w *WindowsLegacyProvider) GetName() string {
	return "Windows Legacy (XP+)"
}

// IsSupported checks if this provider is supported on the current system
func (w *WindowsLegacyProvider) IsSupported() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	
	// For Windows XP+ compatibility, we provide basic sync functionality
	// Advanced mounting features are not available on older Windows versions
	return true
}

// Mount provides basic file synchronization instead of true mounting
// This is a fallback for older Windows systems that don't support Cloud Files API
func (w *WindowsLegacyProvider) Mount(ctx context.Context, mountPoint string) error {
	// On legacy Windows, we can't provide true mounting
	// Instead, we provide enhanced file synchronization with a local mirror
	fmt.Printf("Warning: True mounting is not available on Windows XP/Vista/7/8\n")
	fmt.Printf("Using legacy sync mode with local mirror at: %s\n", mountPoint)
	
	// Create the mount point directory if it doesn't exist
	if err := w.handler.CreateDir(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}
	
	// Start a background sync process
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Perform periodic sync operations
				w.syncFiles(ctx, mountPoint)
			}
		}
	}()
	
	return nil
}

// Unmount stops the sync process
func (w *WindowsLegacyProvider) Unmount() error {
	// Clean up any resources
	return nil
}

// GetCapabilities returns the capabilities of this provider
func (w *WindowsLegacyProvider) GetCapabilities() []string {
	return []string{
		"basic-sync",
		"legacy-windows",
		"file-mirroring",
	}
}

// syncFiles performs periodic synchronization
func (w *WindowsLegacyProvider) syncFiles(ctx context.Context, mountPoint string) error {
	// This would implement basic file synchronization logic
	// For the scope of this example, we'll keep it simple
	
	// List files from the tusd server
	files, err := w.handler.ListFiles("/")
	if err != nil {
		return err
	}
	
	// Sync each file to local mirror
	for _, file := range files {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		localPath := mountPoint + "/" + file.Name
		// Check if file needs updating
		if w.needsUpdate(file, localPath) {
			if err := w.downloadFile(file, localPath); err != nil {
				fmt.Printf("Failed to sync file %s: %v\n", file.Name, err)
			}
		}
	}
	
	return nil
}

// needsUpdate checks if a local file needs to be updated
func (w *WindowsLegacyProvider) needsUpdate(remoteFile *FileInfo, localPath string) bool {
	// Compare modification times, sizes, etc.
	// This is a simplified check
	localInfo, err := w.handler.GetFileInfo(localPath)
	if err != nil {
		return true // File doesn't exist locally
	}
	
	return remoteFile.ModTime.After(localInfo.ModTime)
}

// downloadFile downloads a file from the server to local path
func (w *WindowsLegacyProvider) downloadFile(file *FileInfo, localPath string) error {
	reader, err := w.handler.OpenFile(file.Path, 0)
	if err != nil {
		return err
	}
	defer reader.Close()
	
	writer, err := w.handler.CreateFile(localPath, 0644)
	if err != nil {
		return err
	}
	defer writer.Close()
	
	// Copy file content
	return w.handler.CopyFile(writer, reader)
}