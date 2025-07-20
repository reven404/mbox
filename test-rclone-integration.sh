#!/bin/bash

# Simple test script for rclone integration in mdriver
echo "🧪 Testing Rclone Integration"
echo "============================="

# Test 1: Check if mdriver detects rclone
echo "1. Testing rclone detection..."
./mdriver --platform-info | grep -i rclone
if [ $? -eq 0 ]; then
    echo "✅ Rclone integration detected"
else
    echo "❌ Rclone integration not detected"
    exit 1
fi

# Test 2: Check rclone binary detection
echo -e "\n2. Testing rclone binary detection..."
if command -v rclone >/dev/null 2>&1; then
    RCLONE_VERSION=$(rclone version | head -n1)
    echo "✅ Rclone found: $RCLONE_VERSION"
else
    echo "❌ Rclone not found in PATH"
    exit 1
fi

# Test 3: Test rclone help output
echo -e "\n3. Testing rclone functionality..."
if rclone --help >/dev/null 2>&1; then
    echo "✅ Rclone is functioning properly"
else
    echo "❌ Rclone has issues"
    exit 1
fi

# Test 4: Test configuration directory
echo -e "\n4. Testing rclone config path..."
RCLONE_CONFIG_DIR="$HOME/.config/rclone"
if [ ! -d "$RCLONE_CONFIG_DIR" ]; then
    echo "📁 Creating rclone config directory: $RCLONE_CONFIG_DIR"
    mkdir -p "$RCLONE_CONFIG_DIR"
fi

if [ -d "$RCLONE_CONFIG_DIR" ]; then
    echo "✅ Rclone config directory exists: $RCLONE_CONFIG_DIR"
else
    echo "❌ Could not create rclone config directory"
    exit 1
fi

# Test 5: List remotes (should work even with no remotes)
echo -e "\n5. Testing rclone remote listing..."
REMOTE_COUNT=$(rclone listremotes 2>/dev/null | wc -l)
echo "📋 Found $REMOTE_COUNT configured remotes"

echo -e "\n✅ All rclone integration tests passed!"
echo "📝 Next steps:"
echo "   • Run './rclone-setup.sh' to configure cloud storage remotes"
echo "   • Use './mdriver --mount --config test-rclone-config.yaml' to test mounting"
echo "   • Check providers/rclone/ for advanced configuration options"