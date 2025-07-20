package providers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

// RcloneProvider implements cloud storage mounting using rclone
type RcloneProvider struct {
	config     *Config
	handler    FileSystemHandler
	rclonePath string
	mountCmd   *exec.Cmd
	mounted    bool
	mutex      sync.RWMutex

	// rclone specific configuration
	remoteName   string
	remoteType   string
	mountOptions []string
	vfsCache     string
	logLevel     string
}

// NewRcloneProvider creates a new rclone provider
func NewRcloneProvider(config *Config, handler FileSystemHandler) Provider {
	provider := &RcloneProvider{
		config:       config,
		handler:      handler,
		rclonePath:   findRcloneBinary(),
		mounted:      false,
		remoteName:   config.RcloneConfig,
		vfsCache:     "auto",
		logLevel:     "info",
		mountOptions: []string{},
	}

	// Parse remote configuration
	provider.parseRemoteConfig()

	return provider
}

// GetName returns the provider name
func (r *RcloneProvider) GetName() string {
	if r.remoteType != "" {
		return fmt.Sprintf("Rclone (%s)", strings.Title(r.remoteType))
	}
	return "Rclone Cloud Storage"
}

// IsSupported checks if rclone is available and configured
func (r *RcloneProvider) IsSupported() bool {
	if r.rclonePath == "" {
		return false
	}

	// Check if rclone binary works
	cmd := exec.Command(r.rclonePath, "version")
	err := cmd.Run()
	return err == nil
}

// Mount starts the rclone mount
func (r *RcloneProvider) Mount(ctx context.Context, mountPoint string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.mounted {
		return fmt.Errorf("rclone already mounted")
	}

	if r.rclonePath == "" {
		return fmt.Errorf("rclone binary not found")
	}

	if r.remoteName == "" {
		return fmt.Errorf("rclone remote not configured")
	}

	fmt.Printf("🌩️ Starting rclone mount\n")
	fmt.Printf("📁 Mount Point: %s\n", mountPoint)
	fmt.Printf("☁️ Remote: %s (%s)\n", r.remoteName, r.remoteType)
	fmt.Printf("🛠️ Rclone: %s\n", r.rclonePath)

	// Create mount point
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Prepare rclone mount command
	args := r.buildMountArgs(mountPoint)
	r.mountCmd = exec.CommandContext(ctx, r.rclonePath, args...)

	// Set up logging
	if err := r.setupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	// Start the mount process
	fmt.Printf("🚀 Executing: %s %s\n", r.rclonePath, strings.Join(args, " "))

	if err := r.mountCmd.Start(); err != nil {
		return fmt.Errorf("failed to start rclone mount: %w", err)
	}

	// Wait for mount to be ready
	if err := r.waitForMount(mountPoint, 30*time.Second); err != nil {
		r.mountCmd.Process.Kill()
		return fmt.Errorf("mount failed to become ready: %w", err)
	}

	r.mounted = true

	fmt.Printf("✅ Rclone mount successful\n")
	fmt.Printf("📂 Files available at: %s\n", mountPoint)
	fmt.Printf("🔄 VFS Cache: %s\n", r.vfsCache)

	// Start monitoring services
	go r.startHealthMonitor(ctx)
	go r.startStatsReporter(ctx)

	// Wait for mount process or context cancellation
	go func() {
		<-ctx.Done()
		r.Unmount()
	}()

	return r.mountCmd.Wait()
}

// Unmount stops the rclone mount
func (r *RcloneProvider) Unmount() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.mounted || r.mountCmd == nil {
		return nil
	}

	fmt.Printf("🔻 Unmounting rclone...\n")

	// Try graceful shutdown first
	if r.mountCmd.Process != nil {
		r.mountCmd.Process.Signal(os.Interrupt)

		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- r.mountCmd.Wait()
		}()

		select {
		case <-done:
			fmt.Printf("✅ Rclone unmounted gracefully\n")
		case <-time.After(10 * time.Second):
			// Force kill if graceful shutdown fails
			fmt.Printf("⚠️ Forcing rclone termination\n")
			r.mountCmd.Process.Kill()
			<-done
		}
	}

	r.mounted = false
	r.mountCmd = nil

	return nil
}

// parseRemoteConfig parses the rclone remote configuration
func (r *RcloneProvider) parseRemoteConfig() {
	if r.remoteName == "" {
		return
	}

	// Extract remote name (format: "remote:path" or just "remote:")
	parts := strings.SplitN(r.remoteName, ":", 2)
	if len(parts) > 0 {
		remoteName := parts[0]

		// Get remote type from rclone config
		cmd := exec.Command(r.rclonePath, "config", "show", remoteName)
		output, err := cmd.Output()
		if err == nil {
			r.remoteType = r.extractRemoteType(string(output))
		}
	}
}

// extractRemoteType extracts the remote type from rclone config output
func (r *RcloneProvider) extractRemoteType(configOutput string) string {
	re := regexp.MustCompile(`type\s*=\s*(\w+)`)
	matches := re.FindStringSubmatch(configOutput)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}

// buildMountArgs builds the rclone mount command arguments
func (r *RcloneProvider) buildMountArgs(mountPoint string) []string {
	args := []string{"mount", r.remoteName, mountPoint}

	// VFS cache settings
	args = append(args, "--vfs-cache-mode", r.vfsCache)
	args = append(args, "--vfs-cache-max-age", "1h")
	args = append(args, "--vfs-cache-max-size", "1G")

	// Performance optimizations
	args = append(args, "--buffer-size", "64M")
	args = append(args, "--dir-cache-time", "72h")
	args = append(args, "--vfs-read-chunk-size", "16M")
	args = append(args, "--vfs-read-chunk-size-limit", "0")

	// Reliability settings
	args = append(args, "--checkers", "8")
	args = append(args, "--transfers", "4")
	args = append(args, "--retries", "3")

	// Platform-specific mount options
	switch r.getPlatform() {
	case "darwin":
		args = append(args, "--volname", "Rclone-"+r.remoteName)
		args = append(args, "--daemon-timeout", "10m")
	case "linux":
		args = append(args, "--allow-other")
		args = append(args, "--default-permissions")
	case "windows":
		args = append(args, "--network-mode")
	}

	// Logging
	if r.logLevel != "" {
		args = append(args, "--log-level", strings.ToUpper(r.logLevel))
	}

	// Add custom mount options
	args = append(args, r.mountOptions...)

	return args
}

// setupLogging configures rclone logging
func (r *RcloneProvider) setupLogging() error {
	if r.mountCmd == nil {
		return nil
	}

	// Create log pipes
	stdout, err := r.mountCmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := r.mountCmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start log processors
	go r.processLogs("stdout", stdout)
	go r.processLogs("stderr", stderr)

	return nil
}

// processLogs processes rclone log output
func (r *RcloneProvider) processLogs(source string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse and format rclone log messages
		if r.shouldLogMessage(line) {
			fmt.Printf("📋 [rclone-%s] %s\n", source, line)
		}
	}
}

// shouldLogMessage determines if a log message should be displayed
func (r *RcloneProvider) shouldLogMessage(message string) bool {
	// Filter out verbose messages based on log level
	lowerMsg := strings.ToLower(message)

	switch strings.ToLower(r.logLevel) {
	case "error":
		return strings.Contains(lowerMsg, "error") || strings.Contains(lowerMsg, "fatal")
	case "warn":
		return strings.Contains(lowerMsg, "error") || strings.Contains(lowerMsg, "warn")
	case "info":
		return !strings.Contains(lowerMsg, "debug")
	default:
		return true
	}
}

// waitForMount waits for the mount to be ready
func (r *RcloneProvider) waitForMount(mountPoint string, timeout time.Duration) error {
	start := time.Now()

	for time.Since(start) < timeout {
		// Check if mount point is accessible
		if r.isMountReady(mountPoint) {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("mount did not become ready within %v", timeout)
}

// isMountReady checks if the mount point is ready
func (r *RcloneProvider) isMountReady(mountPoint string) bool {
	// Try to stat the mount point
	_, err := os.Stat(mountPoint)
	if err != nil {
		return false
	}

	// Try to list directory contents
	entries, err := os.ReadDir(mountPoint)
	if err != nil {
		return false
	}

	// Mount is ready if we can list contents (even if empty)
	_ = entries
	return true
}

// startHealthMonitor monitors mount health
func (r *RcloneProvider) startHealthMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.checkMountHealth()
		}
	}
}

// checkMountHealth performs mount health checks
func (r *RcloneProvider) checkMountHealth() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.mounted || r.mountCmd == nil {
		return
	}

	// Check if process is still running
	if r.mountCmd.Process != nil {
		err := r.mountCmd.Process.Signal(syscall.Signal(0)) // Signal 0 checks if process exists
		if err != nil {
			fmt.Printf("⚠️ Rclone process appears to have died: %v\n", err)
		}
	}
}

// startStatsReporter reports mount statistics
func (r *RcloneProvider) startStatsReporter(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.reportStats()
		}
	}
}

// reportStats reports rclone statistics
func (r *RcloneProvider) reportStats() {
	// Get rclone stats via rc (remote control) if available
	cmd := exec.Command(r.rclonePath, "rc", "core/stats")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		fmt.Printf("📊 Rclone stats: %s\n", strings.TrimSpace(string(output)))
	}
}

// getPlatform returns the current platform
func (r *RcloneProvider) getPlatform() string {
	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		return "windows"
	case fileExists("/System/Library/CoreServices/SystemVersion.plist"):
		return "darwin"
	default:
		return "linux"
	}
}

// findRcloneBinary locates the rclone binary
func findRcloneBinary() string {
	// Try common locations and PATH
	candidates := []string{
		"rclone",                                      // In PATH
		"/usr/bin/rclone",                             // Linux/Unix
		"/usr/local/bin/rclone",                       // macOS Homebrew
		"/opt/homebrew/bin/rclone",                    // macOS Apple Silicon
		"C:\\Program Files\\rclone\\rclone.exe",       // Windows
		"C:\\Program Files (x86)\\rclone\\rclone.exe", // Windows 32-bit
	}

	for _, path := range candidates {
		if fileExists(path) || isInPath(path) {
			return path
		}
	}

	return ""
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isInPath checks if a command is in PATH
func isInPath(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetCapabilities returns rclone provider capabilities
func (r *RcloneProvider) GetCapabilities() []string {
	capabilities := []string{
		"multi-cloud-support",
		"fuse-mounting",
		"vfs-caching",
		"cross-platform",
		"high-performance",
		"streaming-support",
		"encryption-support",
	}

	// Add remote-specific capabilities
	switch r.remoteType {
	case "s3":
		capabilities = append(capabilities, "aws-s3", "object-storage")
	case "drive":
		capabilities = append(capabilities, "google-drive", "document-sync")
	case "dropbox":
		capabilities = append(capabilities, "dropbox-sync", "sharing")
	case "onedrive":
		capabilities = append(capabilities, "microsoft-onedrive", "office-365")
	case "azure":
		capabilities = append(capabilities, "azure-blob", "enterprise-storage")
	}

	return capabilities
}

// SetMountOptions allows customizing rclone mount options
func (r *RcloneProvider) SetMountOptions(options []string) {
	r.mountOptions = options
}

// SetVFSCache configures VFS cache mode
func (r *RcloneProvider) SetVFSCache(mode string) {
	validModes := []string{"off", "minimal", "writes", "full", "auto"}
	for _, valid := range validModes {
		if mode == valid {
			r.vfsCache = mode
			return
		}
	}
}

// GetRemoteInfo returns information about the configured remote
func (r *RcloneProvider) GetRemoteInfo() map[string]string {
	info := map[string]string{
		"remote_name": r.remoteName,
		"remote_type": r.remoteType,
		"vfs_cache":   r.vfsCache,
		"rclone_path": r.rclonePath,
	}

	// Add version information
	cmd := exec.Command(r.rclonePath, "version")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 {
			info["rclone_version"] = strings.TrimSpace(lines[0])
		}
	}

	return info
}
