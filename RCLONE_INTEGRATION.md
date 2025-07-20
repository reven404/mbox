# Rclone Integration for mdriver

## Overview

The mdriver project now includes comprehensive rclone integration, providing universal cloud storage mounting capabilities across multiple platforms. This implementation supports mounting various cloud storage services as local file systems using rclone's powerful backend ecosystem.

## Architecture Components

### 1. Core Provider (`providers/rclone_provider.go`)
- **RcloneProvider**: Main implementation of the Provider interface for rclone
- **Features**:
  - Mount/unmount operations with health monitoring
  - VFS caching support with configurable policies
  - Progress tracking and statistics reporting
  - Automatic rclone binary detection
  - Cross-platform mount point management

### 2. Configuration Management (`providers/rclone/config_manager.go`)
- **ConfigManager**: Handles rclone remote configurations
- **Supported Providers**: S3, Google Drive, OneDrive, Dropbox, Google Cloud Storage, Azure Blob, SFTP, FTP, Box, pCloud
- **Features**:
  - Interactive configuration wizard
  - Remote validation and testing
  - Auto-configuration for common providers
  - Export/import functionality

### 3. Sync Service (`providers/rclone/sync_service.go`)
- **SyncService**: Advanced synchronization operations
- **Sync Modes**: sync, copy, move, backup, check
- **Features**:
  - Progress monitoring with real-time stats
  - Multiple concurrent sync sessions
  - Error tracking and retry mechanisms
  - Session management and cleanup

### 4. Serve Service (`providers/rclone/sync_service.go`)
- **ServeService**: Protocol server management
- **Supported Protocols**: HTTP, WebDAV, FTP, SFTP, Restic
- **Features**:
  - Multi-protocol server support
  - Dynamic port management
  - Session lifecycle management

## Setup and Usage

### Quick Setup
```bash
# 1. Make setup script executable
chmod +x rclone-setup.sh

# 2. Run interactive setup
./rclone-setup.sh

# 3. Test integration
./test-rclone-integration.sh
```

### Configuration Examples

#### Basic Configuration (`test-rclone-config.yaml`)
```yaml
tusd_url: "http://localhost:1080"
mount_point: "/tmp/rclone-mount"
mount_type: "rclone"
log_level: "info"
rclone_config: "myremote:"
vfs_cache: "auto"
```

#### Advanced Mount Options
```yaml
mount_options:
  - "--vfs-cache-max-age=1h"
  - "--vfs-cache-max-size=1G"
  - "--buffer-size=64M"
  - "--dir-cache-time=72h"
```

### Platform Integration

#### mdriver CLI Usage
```bash
# Check rclone availability
./mdriver --platform-info | grep rclone

# Mount with rclone provider
./mdriver --mount --config test-rclone-config.yaml

# Use with specific provider type
./mdriver --mount --provider rclone
```

## Supported Cloud Providers

| Provider | Type | Auth Method | Features |
|----------|------|-------------|----------|
| Amazon S3 | `s3` | Access Key | Object storage, versioning, encryption |
| Google Drive | `drive` | OAuth2 | Document sync, sharing, collaboration |
| Microsoft OneDrive | `onedrive` | OAuth2 | Office 365, sharing, versioning |
| Dropbox | `dropbox` | OAuth2 | File sync, sharing, mobile sync |
| Google Cloud Storage | `googlecloudstorage` | Service Account | Enterprise storage, analytics |
| Azure Blob Storage | `azureblob` | Access Key | Hot/cold/archive tiers, CDN |
| SFTP | `sftp` | SSH Key | Secure transfer, Unix permissions |
| FTP | `ftp` | Password | Legacy support, wide compatibility |

## Testing and Validation

### Integration Tests
- `test-rclone-integration.sh`: Comprehensive integration testing
- Rclone binary detection and version verification
- Configuration directory setup
- Remote listing and validation
- Mount capability testing

### Manual Testing
```bash
# List available remotes
rclone listremotes

# Test remote connectivity
rclone lsd myremote: --max-depth 1

# Mount manually for testing
rclone mount myremote: /tmp/test-mount
```

## Architecture Benefits

### Cross-Platform Support
- **Linux**: Native FUSE support
- **macOS**: FUSE integration with macFUSE
- **Windows**: WinFsp compatibility

### Performance Features
- VFS caching with multiple policies (off, minimal, writes, full, auto)
- Configurable buffer sizes and transfer limits
- Background sync capabilities
- Health monitoring and automatic recovery

### Enterprise Features
- Multiple authentication methods (OAuth2, access keys, service accounts)
- Encryption support for secure cloud storage
- Bandwidth limiting and transfer controls
- Comprehensive logging and monitoring

## Implementation Status

✅ **Completed Components**:
- Core rclone provider implementation
- Configuration management system
- Sync service with progress tracking
- Serve service for protocol support
- Interactive setup wizard
- Cross-platform build support
- Integration testing suite

## Next Steps

1. **Authentication Enhancement**: Implement OAuth2 flow automation
2. **Conflict Resolution**: Add advanced file conflict handling
3. **Performance Optimization**: Implement adaptive caching strategies
4. **Monitoring Integration**: Add telemetry and health metrics
5. **Documentation**: Expand user guides and API documentation

The rclone integration provides mdriver with powerful, production-ready cloud storage mounting capabilities that work seamlessly across all supported platforms.