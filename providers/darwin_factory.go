//go:build darwin && fileprovider

package providers

// NewMacOSFileProvider creates the best available macOS provider
// When built with fileprovider tag, uses native File Provider API
func NewMacOSFileProvider(config *Config, handler FileSystemHandler) Provider {
	// Use the real File Provider implementation
	return NewMacOSRealFileProvider(config, handler)
}
