# macOS File Provider Extension

> https://developer.apple.com/documentation/fileprovider

This directory contains the implementation of a native macOS File Provider extension for integrating with the tusd server through Finder.

## Architecture

The File Provider implementation consists of several components:

### Swift Components

1. **FileProviderExtension.swift** - Main extension class implementing `NSFileProviderExtension`
2. **TusdFileProviderItem.swift** - File/folder representation implementing `NSFileProviderItem`
3. **TusdFileProviderEnumerator.swift** - Directory enumeration and change tracking
4. **TusdFileProviderClient.swift** - Communication bridge with tusd server

### Objective-C Bridge

1. **fileprovider_bridge.h/m** - C interface for Go ↔ Swift communication
2. **darwin_fileprovider.go** - Go wrapper using CGO to access the bridge

## Features

- **Native Finder Integration** - Files appear directly in Finder sidebar
- **On-demand Download** - Files are downloaded only when accessed
- **Upload Support** - Drag & drop files to upload via tusd protocol
- **Real-time Sync** - Changes on server reflected immediately in Finder
- **Progress Reporting** - Upload/download progress shown in Finder
- **Conflict Resolution** - Handles concurrent modifications gracefully

## Building

### Prerequisites

- macOS 10.15+ (Catalina or later)
- Xcode 11.0+ or Xcode Command Line Tools
- Swift 5.0+
- Go 1.19+ with CGO support

### Build the Extension

```bash
cd providers/macos
make extension
```

### Install for Development

```bash
make install-extension
```

**Note:** Installing extensions requires administrator privileges and may require additional steps:

1. Disable System Integrity Protection (SIP) temporarily if needed
2. Sign the extension with a valid developer certificate
3. Enable the extension in System Preferences > Extensions

### Remove Extension

```bash
make remove-extension
```

## Configuration

The extension reads configuration from UserDefaults with the suite name `group.com.mdriver.fileprovider`:

- `{domain_id}.serverURL` - Tusd server URL
- `{domain_id}.enabled` - Whether the domain is enabled

## Usage in Go Code

```go
import "mdriver/providers"

// Create File Provider
config := &providers.Config{
    MountPoint: "/tmp/mount",
    // ... other config
}

provider := providers.NewMacOSFileProvider(config, handler)

// Check if supported
if provider.IsSupported() {
    // Mount (registers domain with system)
    ctx := context.Background()
    err := provider.Mount(ctx, "/tmp/mount")
    if err != nil {
        log.Fatal(err)
    }
    
    // The extension will now appear in Finder sidebar
}
```

## Build Tags

The implementation uses build tags to conditionally compile:

- **Default build**: Uses fallback implementation with build tag `!fileprovider`
- **File Provider build**: Use `go build -tags fileprovider` to enable native File Provider

## Development Notes

### Debugging

1. **Console.app** - View extension logs with subsystem `com.mdriver.fileprovider`
2. **Activity Monitor** - Monitor extension processes
3. **Xcode Debugger** - Attach to extension process for debugging

### Extension Lifecycle

1. **Registration** - Go calls `fp_register_domain()` via CGO
2. **System Integration** - macOS loads the extension automatically
3. **File Operations** - Extension receives File Provider API calls
4. **Server Communication** - Extension communicates with tusd server
5. **Cleanup** - Go calls `fp_remove_domain()` to unregister

### Limitations

- Requires macOS 10.15+ for full File Provider API support
- Extension must be signed for distribution
- May require user approval in System Preferences
- Network operations are performed by the extension process, not the main app

## File Provider API Implementation Status

- ✅ Domain registration/removal
- ✅ File/folder enumeration  
- ✅ Item metadata retrieval
- ✅ File content download
- ✅ File upload/creation
- ✅ File modification
- ✅ File deletion
- ✅ Progress reporting
- ✅ Change notifications
- ⚠️ Conflict resolution (basic)
- ⚠️ Offline support (limited)
- ❌ Search integration
- ❌ Sharing/collaboration features

## Security Considerations

- Extension runs in sandboxed environment
- Network access requires appropriate entitlements
- File system access is controlled by File Provider API
- Server authentication should be handled securely
- Consider using App Groups for shared configuration

## Troubleshooting

### Extension Not Loading

1. Check Console.app for error messages
2. Verify extension is properly installed
3. Check System Preferences > Extensions > File Provider
4. Restart Finder: `killall Finder`

### Files Not Appearing

1. Verify server URL is accessible
2. Check extension logs in Console.app
3. Force refresh: `signalEnumerationChange()`
4. Verify File Provider domain is registered

### Upload/Download Failures

1. Check network connectivity
2. Verify tusd server is running and accessible
3. Check file permissions
4. Monitor server logs for errors

For more detailed troubleshooting, enable debug logging and check the system console.