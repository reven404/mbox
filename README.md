# mbox - Universal File Sync and Mount Driver for tusd

A Golang-based universal file synchronization and mounting driver that provides seamless integration with tusd servers across different operating systems.

## Features

### Core Features
- Real-time file system monitoring using fsnotify
- TUS protocol support for reliable file uploads
- Configurable chunk-based upload for large files
- Retry mechanism with exponential backoff
- Comprehensive logging with configurable levels
- YAML-based configuration

### Platform-Specific Mount Integration
- **macOS**: Native File Provider API integration (macOS 10.15+)
- **Windows**: Windows Cloud Files API integration (Windows 10 1809+)
- **Universal**: rclone-based mounting for all platforms
- **Legacy**: WebDAV server for network drive access

### Mount Capabilities
- Native OS integration for optimal user experience
- Automatic platform detection and provider selection
- Caching for improved performance
- Support for large files and resumable transfers
- Cross-platform compatibility

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](../../releases/latest).

#### Platform Support Matrix

| Platform | Architecture | Support Level | Notes |
|----------|-------------|---------------|-------|
| **macOS** | Intel (x64) | Native | File Provider API integration |
| **macOS** | Apple Silicon (ARM64) | Native | Universal binary available |
| **Windows 10+** | x64 | Native | Cloud Files API integration |
| **Windows XP+** | x86/x64 | Legacy | Basic sync functionality |
| **Linux** | x64/ARM64 | Full | FUSE + rclone SDK support |

#### Quick Installation

```bash
# Linux/macOS - Download and install latest version
curl -fsSL https://raw.githubusercontent.com/username/mbox/main/install.sh | bash

# Windows - Download from releases page or use Scoop
scoop bucket add username https://github.com/username/scoop-bucket
scoop install mbox

# macOS - Use Homebrew
brew tap username/tap
brew install mbox
```

### Build from Source

Requirements:
- Go 1.20 or later
- CGO enabled for platform-specific features

```bash
# Clone the repository
git clone https://github.com/username/mbox.git
cd mbox

# Build for your platform
go build -o mbox .

# Or use GoReleaser for multi-platform builds
goreleaser build --single-target
```

## Usage

### File Synchronization Mode (Default)

Monitor local directory and sync files to tusd server:

```bash
# Basic sync mode
./mbox

# With custom configuration
./mbox -config /path/to/config.yaml
```

### Mount Mode

Mount tusd server as local filesystem:

```bash
# Enable mount mode
./mbox -mount -config config.yaml

# Or set mount_point in config.yaml
./mbox  # Will auto-detect mount mode if mount_point is set
```

### Platform Information

Check available mount capabilities on your system:

```bash
./mbox -platform-info
```

### Show Version

```bash
./mbox -version
```

## Configuration

Create a `config.yaml` file with the following structure:

```yaml
tusd_url: "http://localhost:1080/files/"
watch_dir: "./sync"
log_level: "info"
chunk_size: 1048576  # 1MB
max_retries: 3
retry_delay_seconds: 5
```

### Configuration Options

- `tusd_url`: URL of the tusd server endpoint
- `watch_dir`: Directory to monitor for file changes
- `log_level`: Logging level (debug, info, warn, error)
- `chunk_size`: Size of chunks for file upload (in bytes)
- `max_retries`: Maximum number of retry attempts for failed uploads
- `retry_delay_seconds`: Delay between retry attempts (in seconds)

## Dependencies

- `github.com/fsnotify/fsnotify` - File system notifications
- `github.com/sirupsen/logrus` - Structured logging
- `gopkg.in/yaml.v3` - YAML configuration parsing

## How It Works

1. **File System Monitoring**: Uses fsnotify to watch for file system events in the specified directory
2. **Event Processing**: Filters and processes file events (create, write, rename)
3. **Upload Queue**: Queues files for upload to prevent overwhelming the system
4. **TUS Protocol**: Implements TUS resumable upload protocol for reliable file transfers
5. **Error Handling**: Provides retry mechanisms and comprehensive error logging

## File Filtering

The driver automatically ignores:
- Hidden files (starting with `.`)
- Temporary files (ending with `~` or `.tmp`)
- System files (like `.DS_Store`)
- Directories

## Architecture

```
main.go           - Application entry point and signal handling
config.go         - Configuration management
logger.go         - Logging setup and configuration
watcher.go        - File system monitoring using fsnotify
tusd_client.go    - TUS protocol client implementation
sync_engine.go    - Main synchronization logic and coordination
```

## Development

### Building

```bash
# Standard build
go build .

# Build with specific platform tags
go build -tags=legacy,windowsxp .  # Windows XP compatibility
go build -tags=cgo,static .        # Static build with CGO

# Cross-compilation examples
GOOS=windows GOARCH=386 go build -tags=legacy,windowsxp .
GOOS=darwin GOARCH=arm64 go build .
GOOS=linux GOARCH=arm64 go build .
```

### Release Process

This project uses [GoReleaser](https://goreleaser.com/) for automated releases:

```bash
# Test the release process without publishing
goreleaser release --snapshot --clean

# Create a release (requires git tag)
git tag v1.0.0
git push origin v1.0.0  # This triggers GitHub Actions release workflow
```

#### Release Artifacts

Each release includes:
- **Binaries**: Pre-compiled for all supported platforms
- **Archives**: tar.gz (Linux/macOS) and zip (Windows) with configs
- **Docker Images**: Multi-arch containers for Linux
- **Package Manager**: Homebrew (macOS), Scoop (Windows), Snapcraft (Linux)
- **Checksums**: SHA256 verification files

### Running Tests

```bash
go test ./...
```

### Running with Debug Logging

```bash
./mbox -config config.yaml
# Set log_level: "debug" in config.yaml for verbose logging
```

## 参考

- https://github.com/libfuse/libfuse
- https://github.com/rclone/rclone
- https://github.com/juicedata/juicefs

## License

This project is licensed under the MIT License.