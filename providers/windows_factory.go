//go:build windows && cloudfiles

package providers

// NewWindowsCloudFiles creates the best available Windows provider
// When built with cloudfiles tag, uses native Cloud Files API
func NewWindowsCloudFiles(config *Config, handler FileSystemHandler) Provider {
	// Use the real Cloud Files implementation
	return NewWindowsCloudFilesProvider(config, handler)
}

// NewWindowsLegacyProvider is also available for older Windows versions
func NewWindowsLegacyProvider(config *Config, handler FileSystemHandler) Provider {
	// Check if we should use Cloud Files API first
	cloudProvider := NewWindowsCloudFilesProvider(config, handler)
	if cloudProvider.IsSupported() {
		return cloudProvider
	}

	// Fall back to legacy provider for Windows XP/Vista/7/8/Early 10
	return NewWindowsLegacyProvider(config, handler)
}
