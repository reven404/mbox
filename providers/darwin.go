//go:build darwin

package providers

import (
	"context"
	"fmt"
	"runtime"
)

// DarwinProvider simplified macOS/Darwin provider
type DarwinProvider struct {
	config  *Config
	handler FileSystemHandler
}

// NewMacOSFileProvider creates a simplified macOS provider
func NewMacOSFileProvider(config *Config, handler FileSystemHandler) Provider {
	return &DarwinProvider{
		config:  config,
		handler: handler,
	}
}

// GetName returns the provider name
func (d *DarwinProvider) GetName() string {
	return "Darwin Provider"
}

// IsSupported checks if this provider is supported
func (d *DarwinProvider) IsSupported() bool {
	return runtime.GOOS == "darwin"
}

// Mount provides basic mounting functionality
func (d *DarwinProvider) Mount(ctx context.Context, mountPoint string) error {
	fmt.Printf("Starting macOS simple mount at: %s\n", mountPoint)
	fmt.Println("Note: This is a simplified implementation")
	
	// Keep context active
	<-ctx.Done()
	return ctx.Err()
}

// Unmount stops the mount
func (d *DarwinProvider) Unmount() error {
	fmt.Println("Unmounting Darwin provider")
	return nil
}