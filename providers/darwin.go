//go:build darwin && !fileprovider

package providers

import (
	"context"
	"fmt"
	"runtime"
)

// DarwinProvider simplified macOS/Darwin provider (fallback)
type DarwinProvider struct {
	config  *Config
	handler FileSystemHandler
}

// NewMacOSFileProvider creates a provider - will use File Provider API if available
func NewMacOSFileProvider(config *Config, handler FileSystemHandler) Provider {
	// For fallback build, always use the simple implementation
	// Real File Provider is only available with -tags fileprovider

	// Fall back to simplified provider
	return &DarwinProvider{
		config:  config,
		handler: handler,
	}
}

// GetName returns the provider name
func (d *DarwinProvider) GetName() string {
	return "Darwin Fallback Provider"
}

// IsSupported checks if this provider is supported
func (d *DarwinProvider) IsSupported() bool {
	return runtime.GOOS == "darwin"
}

// Mount provides basic mounting functionality
func (d *DarwinProvider) Mount(ctx context.Context, mountPoint string) error {
	fmt.Printf("Starting macOS fallback mount at: %s\n", mountPoint)
	fmt.Println("Note: Using fallback implementation - File Provider API not available")
	fmt.Println("For native Finder integration, build with -tags fileprovider on macOS 10.15+")

	// Keep context active
	<-ctx.Done()
	return ctx.Err()
}

// Unmount stops the mount
func (d *DarwinProvider) Unmount() error {
	fmt.Println("Unmounting Darwin fallback provider")
	return nil
}
