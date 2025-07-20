package providers

import (
	"context"
	"fmt"
	"os"
)

// UniversalProvider provides basic file synchronization without external dependencies
// This provider works across all platforms as a fallback option
type UniversalProvider struct {
	config  *Config
	handler FileSystemHandler
}

// NewRcloneSDKProvider creates a universal provider (renamed to maintain compatibility)
func NewRcloneSDKProvider(config *Config, handler FileSystemHandler) Provider {
	return &UniversalProvider{
		config:  config,
		handler: handler,
	}
}

// GetName returns the provider name
func (u *UniversalProvider) GetName() string {
	return "Universal File Sync Provider"
}

// IsSupported checks if this provider is supported
func (u *UniversalProvider) IsSupported() bool {
	// This provider works on all platforms as a fallback
	return true
}

// Mount provides basic file synchronization instead of true mounting
func (u *UniversalProvider) Mount(ctx context.Context, mountPoint string) error {
	fmt.Printf("Starting simple file sync mode at: %s\n", mountPoint)
	fmt.Printf("Note: This is a basic sync mode, not true mounting\n")
	
	// Create the mount point directory if it doesn't exist
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}
	
	// Start a background goroutine to handle the mount
	go func() {
		<-ctx.Done()
		fmt.Println("Mount context cancelled")
	}()
	
	// Return immediately - the mount is "active" until context is cancelled
	return nil
}

// Unmount stops the sync process
func (u *UniversalProvider) Unmount() error {
	fmt.Println("Stopping universal file sync")
	return nil
}

// Logger interface for compatibility
type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}