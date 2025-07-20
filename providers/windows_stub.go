//go:build !windows

package providers

import (
	"context"
	"fmt"
)

// WindowsCloudFilesStub provides a stub implementation for non-Windows platforms
type WindowsCloudFilesStub struct {
	config *Config
}

// WindowsLegacyStub provides a stub for Windows legacy provider
type WindowsLegacyStub struct {
	config *Config
}

// NewWindowsCloudFiles creates a stub provider for non-Windows platforms
func NewWindowsCloudFiles(config *Config, handler FileSystemHandler) Provider {
	return &WindowsCloudFilesStub{config: config}
}

// NewWindowsLegacyProvider creates a stub provider for non-Windows platforms
func NewWindowsLegacyProvider(config *Config, handler FileSystemHandler) Provider {
	return &WindowsLegacyStub{config: config}
}

// GetName returns the provider name
func (w *WindowsCloudFilesStub) GetName() string {
	return "Windows Cloud Files API (Unsupported)"
}

// IsSupported always returns false for non-Windows platforms
func (w *WindowsCloudFilesStub) IsSupported() bool {
	return false
}

// Mount returns an error indicating this is not supported
func (w *WindowsCloudFilesStub) Mount(ctx context.Context, mountPoint string) error {
	return fmt.Errorf("Windows Cloud Files API is only available on Windows 10 1809+ with cloudfiles build tag")
}

// Unmount is a no-op
func (w *WindowsCloudFilesStub) Unmount() error {
	return nil
}

// Windows Legacy Stub methods
func (w *WindowsLegacyStub) GetName() string {
	return "Windows Legacy Provider (Unsupported)"
}

func (w *WindowsLegacyStub) IsSupported() bool {
	return false
}

func (w *WindowsLegacyStub) Mount(ctx context.Context, mountPoint string) error {
	return fmt.Errorf("Windows Legacy provider is only available on Windows with legacy build tags")
}

func (w *WindowsLegacyStub) Unmount() error {
	return nil
}
