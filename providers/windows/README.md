# Windows Cloud Files API - Smart Folders

> https://learn.microsoft.com/en-us/windows/win32/cfapi/build-a-cloud-file-sync-engine

This directory contains the implementation of Windows Cloud Files API integration for creating Smart Folders with on-demand file synchronization.

## Overview

Windows Cloud Files API (introduced in Windows 10 version 1809) enables applications to create **Smart Folders** that provide:

- **On-demand file hydration** - Files download when accessed
- **Space-efficient storage** - Only metadata stored locally until needed
- **Native Windows Explorer integration** - Cloud status icons and context menus
- **Placeholder file management** - Seamless user experience

## Architecture

### Core Components

1. **cloudfiles_bridge.h/cpp** - C++ wrapper for Windows Cloud Files API
2. **windows_cloudfiles.go** - Go provider implementation with CGO bindings
3. **windows_factory.go** - Provider selection and instantiation
4. **windows_stub.go** - Cross-platform compatibility stubs

### Smart Folder Features

| Feature | Description |
|---------|-------------|
| **Sync Root Registration** | Register folder as cloud storage location |
| **Placeholder Files** | Create file entries with metadata only |
| **On-Demand Hydration** | Download content when file is accessed |
| **Dehydration** | Remove content while keeping metadata |
| **Progress Indication** | Show download/upload progress in Explorer |
| **Context Menus** | Right-click options for pin/unpin operations |
| **Cloud Status Icons** | Visual indicators for file sync state |

## API Integration

### Windows Cloud Files API Functions

```cpp
// Sync Root Management
CfRegisterSyncRoot()     // Register smart folder
CfUnregisterSyncRoot()   // Unregister smart folder

// Placeholder Operations  
CfCreatePlaceholders()   // Create file placeholders
CfHydratePlaceholder()   // Download file content
CfDehydratePlaceholder() // Remove content, keep metadata
CfUpdatePlaceholder()    // Update file metadata

// State Management
CfSetInSyncState()       // Mark file as synchronized
CfGetPlaceholderState()  // Get current file state
```

### File States

| State | Description | Explorer Icon |
|-------|-------------|---------------|
| **Placeholder** | Metadata only, no content | ☁️ Cloud icon |
| **Partially Hydrated** | Some content downloaded | ⏳ Sync icon |
| **Fully Hydrated** | Complete content available | ✅ Green checkmark |
| **Pinned** | Always keep on device | 📌 Pin icon |
| **Unpinned** | Can be dehydrated | 📤 Available online |

## Building

### Prerequisites

- **Windows 10 version 1809+** (build 17763 or later)
- **Visual Studio 2019/2022** with C++ build tools
- **Windows 10 SDK** (latest version)
- **Go 1.19+** with CGO enabled

### Build Steps

```powershell
# Check system compatibility
cd providers/windows
make test

# Build Cloud Files bridge
make bridge

# Build Go application with Cloud Files support
set CGO_ENABLED=1
go build -tags cloudfiles ../..
```

### Cross-Platform Building

```bash
# On non-Windows systems, stub implementation is used
go build .  # Builds with stubs

# For Windows with Cloud Files API
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -tags cloudfiles .
```

## Usage

### Configuration

```yaml
# config.yaml
mount_point: "C:\\Users\\Username\\SmartFolder"
mount_type: "windows-cloudfiles"
tusd_url: "https://your-tusd-server.com"
display_name: "My Cloud Storage"
```

### Running Smart Folder

```powershell
# Start smart folder
.\mdriver.exe --mount --config config.yaml

# The folder will appear in Windows Explorer with:
# - Cloud status icons
# - Right-click context menus
# - On-demand file access
```

### User Experience

1. **Initial Setup**
   - Smart folder appears in specified location
   - Files show as cloud icons (placeholder state)
   - Folder integrates with Windows Explorer

2. **File Access**
   - Double-click file → Automatic download begins
   - Progress shown in Explorer status
   - File becomes fully available

3. **Context Menu Options**
   - **"Always keep on this device"** → Pin file locally
   - **"Free up space"** → Convert to placeholder
   - **"Download"** → Force hydration

4. **Visual Indicators**
   - ☁️ **Cloud icon**: File available online only
   - ⏳ **Sync icon**: Download/upload in progress
   - ✅ **Green check**: File fully synced locally
   - 📌 **Pin icon**: File pinned to device

## Smart Folder Operations

### File Hydration (Download)

```go
// Trigger file download
err := provider.HydrateFile("document.pdf", HYDRATION_POLICY_FULL)
if err != nil {
    log.Printf("Failed to hydrate file: %v", err)
}
```

### File Dehydration (Free Space)

```go  
// Remove local content, keep metadata
err := provider.DehydrateFile("large-video.mp4")
if err != nil {
    log.Printf("Failed to dehydrate file: %v", err)
}
```

### Placeholder Creation

```go
// Create placeholder for remote file
placeholder := &FilePlaceholder{
    RelativePath:  "remote-document.pdf",
    FileSize:     2048576,  // 2MB
    FileID:       123,
    ModifiedTime: time.Now().Unix(),
    IsDirectory:  false,
}

err := provider.CreatePlaceholders([]*FilePlaceholder{placeholder})
```

## Implementation Details

### Callback Handling

The Cloud Files API uses callbacks for file operations:

```cpp
void CALLBACK CloudFileCallback(
    _In_ CONST CF_CALLBACK_INFO* CallbackInfo,
    _In_ CONST CF_CALLBACK_PARAMETERS* CallbackParameters
) {
    switch (CallbackInfo->CallbackType) {
        case CF_CALLBACK_TYPE_FETCH_DATA:
            // Handle file download request
            break;
        case CF_CALLBACK_TYPE_VALIDATE_DATA:
            // Validate downloaded content
            break;
        case CF_CALLBACK_TYPE_CANCEL_FETCH_DATA:
            // Handle download cancellation
            break;
    }
}
```

### CGO Integration

The Go implementation uses CGO to call Windows APIs:

```go
/*
#cgo CFLAGS: -I./windows
#cgo LDFLAGS: -L./windows -lole32 -lshell32
#include "cloudfiles_bridge.h"
*/
import "C"

func (w *WindowsCloudFilesProvider) Mount(ctx context.Context, mountPoint string) error {
    syncRootPathW := stringToWideChar(mountPoint)
    result := C.cf_register_sync_root(syncRootPathW, displayNameW, serverURIW)
    // Handle result...
}
```

## Testing

### Manual Testing

1. **Register Smart Folder**
   ```powershell
   .\mdriver.exe --mount --config test-config.yaml
   ```

2. **Verify Explorer Integration**
   - Open Windows Explorer
   - Navigate to mount point
   - Verify cloud icons appear
   - Test right-click context menus

3. **Test File Operations**
   - Click on placeholder file
   - Verify download progress
   - Test pin/unpin operations
   - Verify space savings with dehydration

### Automated Testing

```go
func TestCloudFilesAPI(t *testing.T) {
    provider := NewWindowsCloudFilesProvider(config, handler)
    
    // Test support detection
    assert.True(t, provider.IsSupported())
    
    // Test sync root registration
    err := provider.Mount(ctx, testPath)
    assert.NoError(t, err)
    
    // Test placeholder creation
    err = provider.CreatePlaceholders(placeholders)
    assert.NoError(t, err)
    
    // Cleanup
    provider.Unmount()
}
```

## Troubleshooting

### Common Issues

1. **"Cloud Files API not supported"**
   - Verify Windows 10 version 1809+ (build 17763+)
   - Check system requirements

2. **"Failed to register sync root"**
   - Ensure path exists and is accessible
   - Check permissions
   - Verify no other sync root at same path

3. **"CGO compilation failed"**
   - Install Visual Studio C++ build tools
   - Verify Windows SDK installation
   - Set correct CGO environment variables

### Debug Information

Enable debug logging to see Cloud Files operations:

```yaml
debug_mode: true
log_level: "debug"
```

### System Requirements Check

```powershell
# Check Windows version
winver

# Check Cloud Files API availability  
powershell -Command "[System.Environment]::OSVersion.Version"

# Verify build tools
where cl  # MSVC compiler
where link  # Linker
```

## Comparison with Other Providers

| Feature | Cloud Files | File Provider (macOS) | Rclone | WebDAV |
|---------|-------------|----------------------|---------|---------|
| **OS Integration** | ✅ Native Windows | ✅ Native macOS | ❌ FUSE | ❌ Network |
| **Space Efficiency** | ✅ Placeholders | ✅ On-demand | ❌ Full sync | ❌ Full sync |
| **Progress Indication** | ✅ Explorer UI | ✅ Finder UI | ❌ Terminal | ❌ None |
| **Context Menus** | ✅ Pin/Unpin | ✅ Download | ❌ None | ❌ None |
| **Cloud Icons** | ✅ Status icons | ✅ Status icons | ❌ None | ❌ None |
| **Background Sync** | ✅ Smart | ✅ Smart | ✅ Manual | ❌ Manual |

## Security Considerations

- **Sync root permissions**: Ensure proper access controls
- **Data transmission**: Use HTTPS for server communication  
- **Local storage**: Files stored in user's profile directory
- **API access**: Cloud Files requires elevated permissions for registration
- **Server authentication**: Implement proper auth for tusd server

## Future Enhancements

- [ ] **Conflict resolution** for concurrent modifications
- [ ] **Offline support** with queued operations  
- [ ] **Bandwidth throttling** for large file transfers
- [ ] **Selective sync** with folder filtering
- [ ] **Version history** integration
- [ ] **Sharing integration** with Windows sharing APIs
- [ ] **Search integration** with Windows Search
- [ ] **Thumbnail generation** for media files

## References

- [Windows Cloud Files API Documentation](https://docs.microsoft.com/en-us/windows/win32/cfapi/cloud-files-api-portal)
- [Cloud Files API Samples](https://github.com/microsoft/Windows-classic-samples/tree/master/Samples/CloudMirror)
- [File System Minifilter Drivers](https://docs.microsoft.com/en-us/windows-hardware/drivers/ifs/)
- [Windows Storage Spaces](https://docs.microsoft.com/en-us/windows-server/storage/storage-spaces/overview)