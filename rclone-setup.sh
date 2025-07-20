#!/bin/bash

# Rclone Setup Script for mdriver
# This script helps install rclone and configure cloud storage remotes

set -e

echo "🌩️ Rclone Setup for mdriver"
echo "=========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    printf "${1}${2}${NC}\n"
}

# Check if running on supported platform
check_platform() {
    case "$(uname -s)" in
        Linux*)     PLATFORM=linux;;
        Darwin*)    PLATFORM=macos;;
        CYGWIN*|MINGW*|MSYS*) PLATFORM=windows;;
        *)          PLATFORM=unknown;;
    esac
    
    print_color $BLUE "Detected platform: $PLATFORM"
}

# Check if rclone is installed
check_rclone() {
    if command -v rclone >/dev/null 2>&1; then
        RCLONE_VERSION=$(rclone version | head -n1)
        print_color $GREEN "✅ Rclone is installed: $RCLONE_VERSION"
        return 0
    else
        print_color $YELLOW "⚠️ Rclone is not installed"
        return 1
    fi
}

# Install rclone
install_rclone() {
    print_color $BLUE "Installing rclone..."
    
    case $PLATFORM in
        linux)
            if command -v apt-get >/dev/null 2>&1; then
                # Debian/Ubuntu
                sudo apt-get update
                sudo apt-get install -y rclone
            elif command -v yum >/dev/null 2>&1; then
                # RHEL/CentOS
                sudo yum install -y epel-release
                sudo yum install -y rclone
            elif command -v brew >/dev/null 2>&1; then
                # Linux with Homebrew
                brew install rclone
            else
                # Use curl install script
                curl https://rclone.org/install.sh | sudo bash
            fi
            ;;
        macos)
            if command -v brew >/dev/null 2>&1; then
                brew install rclone
            else
                print_color $YELLOW "Installing Homebrew first..."
                /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
                brew install rclone
            fi
            ;;
        windows)
            print_color $BLUE "For Windows, please download rclone from: https://rclone.org/downloads/"
            print_color $BLUE "Or use winget: winget install Rclone.Rclone"
            exit 1
            ;;
        *)
            print_color $RED "Unsupported platform: $PLATFORM"
            exit 1
            ;;
    esac
    
    print_color $GREEN "✅ Rclone installation completed"
}

# Show supported cloud providers
show_providers() {
    print_color $BLUE "\n☁️ Supported Cloud Storage Providers:"
    echo "====================================="
    
    cat << EOF
1. Amazon S3 (and compatible)
   - AWS S3, MinIO, DigitalOcean Spaces
   - Type: s3

2. Google Drive
   - Personal and G Suite accounts
   - Type: drive

3. Microsoft OneDrive
   - Personal and Business accounts
   - Type: onedrive

4. Dropbox
   - Personal and Business accounts
   - Type: dropbox

5. Google Cloud Storage
   - Enterprise cloud storage
   - Type: googlecloudstorage

6. Azure Blob Storage
   - Microsoft Azure storage
   - Type: azureblob

7. SFTP/SSH
   - Secure file transfer
   - Type: sftp

8. FTP
   - Traditional file transfer
   - Type: ftp

9. Box
   - Box cloud storage
   - Type: box

10. pCloud
    - pCloud storage
    - Type: pcloud
EOF
}

# Interactive remote configuration
configure_remote() {
    print_color $BLUE "\n🔧 Configuring rclone remote..."
    
    echo "Let's set up a new remote connection."
    read -p "Enter a name for your remote (e.g., 'mycloud'): " REMOTE_NAME
    
    if [ -z "$REMOTE_NAME" ]; then
        print_color $RED "Remote name cannot be empty"
        exit 1
    fi
    
    show_providers
    echo
    read -p "Enter the provider type from above (e.g., 's3', 'drive'): " PROVIDER_TYPE
    
    if [ -z "$PROVIDER_TYPE" ]; then
        print_color $RED "Provider type cannot be empty"
        exit 1
    fi
    
    print_color $BLUE "Starting interactive configuration..."
    rclone config create "$REMOTE_NAME" "$PROVIDER_TYPE"
    
    if [ $? -eq 0 ]; then
        print_color $GREEN "✅ Remote '$REMOTE_NAME' configured successfully"
        
        # Test the remote
        print_color $BLUE "Testing remote connection..."
        if rclone lsd "$REMOTE_NAME:" --max-depth 1 >/dev/null 2>&1; then
            print_color $GREEN "✅ Remote test successful"
        else
            print_color $YELLOW "⚠️ Remote test failed - please check your configuration"
        fi
    else
        print_color $RED "❌ Remote configuration failed"
        exit 1
    fi
}

# List existing remotes
list_remotes() {
    print_color $BLUE "\n📁 Existing rclone remotes:"
    
    if rclone listremotes >/dev/null 2>&1; then
        REMOTES=$(rclone listremotes)
        if [ -n "$REMOTES" ]; then
            echo "$REMOTES" | while read -r remote; do
                if [ -n "$remote" ]; then
                    remote_name=$(echo "$remote" | sed 's/:$//')
                    remote_type=$(rclone config show "$remote_name" 2>/dev/null | grep "type = " | cut -d' ' -f3)
                    echo "  • $remote_name ($remote_type)"
                fi
            done
        else
            print_color $YELLOW "No remotes configured"
        fi
    else
        print_color $RED "Failed to list remotes"
    fi
}

# Generate mdriver config
generate_config() {
    print_color $BLUE "\n📝 Generating mdriver configuration..."
    
    list_remotes
    echo
    read -p "Enter the remote name to use with mdriver: " SELECTED_REMOTE
    
    if [ -z "$SELECTED_REMOTE" ]; then
        print_color $RED "Remote name cannot be empty"
        exit 1
    fi
    
    # Remove trailing colon if present
    SELECTED_REMOTE=$(echo "$SELECTED_REMOTE" | sed 's/:$//')
    
    # Validate remote exists
    if ! rclone config show "$SELECTED_REMOTE" >/dev/null 2>&1; then
        print_color $RED "Remote '$SELECTED_REMOTE' not found"
        exit 1
    fi
    
    # Create config file
    CONFIG_FILE="mdriver-rclone-config.yaml"
    
    cat > "$CONFIG_FILE" << EOF
# mdriver configuration with rclone
tusd_url: "http://localhost:1080"
mount_point: "\$HOME/CloudStorage/$SELECTED_REMOTE"
mount_type: "rclone"
log_level: "info"

# Rclone configuration
rclone_config: "$SELECTED_REMOTE:"

# Mount options (adjust as needed)
vfs_cache: "auto"
mount_options:
  - "--vfs-cache-max-age=1h"
  - "--vfs-cache-max-size=1G"
  - "--buffer-size=64M"
  - "--dir-cache-time=72h"

# Performance settings
chunk_size: 16777216  # 16MB
max_retries: 3
retry_delay: 2

# Cache settings
cache_dir: "\$HOME/.cache/mdriver"
cache_size: 1073741824  # 1GB

# Advanced options
debug_mode: false
background_sync: true
health_check_interval: 30
stats_reporting_interval: 60
EOF
    
    print_color $GREEN "✅ Configuration saved to: $CONFIG_FILE"
    print_color $BLUE "You can now run: ./mdriver --mount --config $CONFIG_FILE"
}

# Test rclone with mdriver
test_integration() {
    print_color $BLUE "\n🧪 Testing rclone integration..."
    
    if [ ! -f "./mdriver" ]; then
        print_color $YELLOW "mdriver binary not found, building..."
        if [ -f "go.mod" ]; then
            go build .
        else
            print_color $RED "Not in mdriver directory or binary not found"
            exit 1
        fi
    fi
    
    # Test rclone detection
    ./mdriver --platform-info | grep -i rclone
    
    if [ $? -eq 0 ]; then
        print_color $GREEN "✅ Rclone integration detected"
    else
        print_color $YELLOW "⚠️ Rclone integration not detected"
    fi
}

# Main menu
show_menu() {
    echo
    print_color $BLUE "What would you like to do?"
    echo "1. Install rclone"
    echo "2. Configure new remote"
    echo "3. List existing remotes" 
    echo "4. Generate mdriver config"
    echo "5. Test integration"
    echo "6. Exit"
    echo
}

# Main script
main() {
    check_platform
    
    if check_rclone; then
        echo
    else
        read -p "Would you like to install rclone? (y/N): " INSTALL_CHOICE
        if [[ $INSTALL_CHOICE =~ ^[Yy]$ ]]; then
            install_rclone
        else
            print_color $YELLOW "Rclone is required for this integration"
            exit 1
        fi
    fi
    
    while true; do
        show_menu
        read -p "Choose an option (1-6): " CHOICE
        
        case $CHOICE in
            1)
                install_rclone
                ;;
            2)
                configure_remote
                ;;
            3)
                list_remotes
                ;;
            4)
                generate_config
                ;;
            5)
                test_integration
                ;;
            6)
                print_color $GREEN "Goodbye! 👋"
                exit 0
                ;;
            *)
                print_color $RED "Invalid choice. Please try again."
                ;;
        esac
    done
}

# Run main function
main "$@"