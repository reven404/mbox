package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	
	"github.com/sirupsen/logrus"
	"mdriver/providers"
)

// Version information set by GoReleaser
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	var configPath = flag.String("config", "config.yaml", "Path to configuration file")
	var showVersion = flag.Bool("version", false, "Show version information")
	var mountMode = flag.Bool("mount", false, "Enable mount mode")
	var showPlatformInfo = flag.Bool("platform-info", false, "Show platform information and mount capabilities")
	flag.Parse()

	if *showVersion {
		fmt.Printf("mdriver %s\n", version)
		fmt.Printf("  Commit:     %s\n", commit)
		fmt.Printf("  Built:      %s by %s\n", date, builtBy)
		fmt.Printf("  Go version: %s\n", runtime.Version())
		fmt.Printf("  Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	}

	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(config.LogLevel)

	// 显示平台信息
	if *showPlatformInfo {
		showPlatformCapabilities(logger)
		return
	}

	// 确定运行模式
	runMountMode := *mountMode || config.MountPoint != ""
	
	if runMountMode {
		logger.Info("Starting mdriver in mount mode")
		runMount(config, logger)
	} else {
		logger.Info("Starting mdriver in sync mode")
		runSync(config, logger)
	}
}

func runSync(config *Config, logger *logrus.Logger) {
	logger.Infof("Starting sync mode: tusd_url=%s, watch_dir=%s", 
		config.TusdURL, config.WatchDir)

	syncEngine, err := NewSyncEngine(config, logger)
	if err != nil {
		logger.Fatalf("Failed to create sync engine: %v", err)
	}

	if err := syncEngine.Start(); err != nil {
		logger.Fatalf("Failed to start sync engine: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Sync mode is running. Press Ctrl+C to stop.")
	
	<-sigChan
	logger.Info("Shutdown signal received")

	syncEngine.Stop()
	logger.Info("Sync stopped gracefully")
}

func runMount(config *Config, logger *logrus.Logger) {
	if config.MountPoint == "" {
		logger.Fatal("Mount point not specified in config")
	}

	logger.Infof("Starting mount mode: tusd_url=%s, mount_point=%s, mount_type=%s", 
		config.TusdURL, config.MountPoint, config.MountType)

	// 创建 tusd 客户端
	tusdClient := NewTusdClient(
		config.TusdURL,
		config.ChunkSize,
		config.MaxRetries,
		time.Duration(config.RetryDelay)*time.Second,
		logger,
	)

	// 创建文件系统处理器
	adaptedClient := NewTusdClientAdapter(tusdClient)
	handler := providers.NewTusdFileSystemHandler(adaptedClient, config.CacheDir, logger)

	// 创建挂载配置
	mountConfig := &providers.Config{
		MountPoint:   config.MountPoint,
		MountType:    config.MountType,
		ReadOnly:     config.ReadOnly,
		CacheDir:     config.CacheDir,
		CacheSize:    config.CacheSize,
		DebugMode:    config.DebugMode,
		BundleID:     config.BundleID,
		DisplayName:  config.DisplayName,
		ProviderName: config.ProviderName,
		SyncRootPath: config.SyncRootPath,
		RcloneConfig: config.RcloneConfig,
	}

	// 选择挂载提供者
	var provider providers.Provider
	if config.MountType == "auto" {
		provider = providers.SelectBestProvider(mountConfig, handler)
		logger.Infof("Auto-selected mount provider: %s", provider.GetName())
	} else {
		provider = createProvider(config.MountType, mountConfig, handler)
		if provider == nil {
			logger.Fatalf("Unsupported mount type: %s", config.MountType)
		}
	}

	// 检查支持性
	if !provider.IsSupported() {
		logger.Fatalf("Mount provider %s is not supported on this platform", provider.GetName())
	}

	// 启动挂载
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := provider.Mount(ctx, config.MountPoint); err != nil {
		logger.Fatalf("Failed to mount: %v", err)
	}

	logger.Infof("Successfully mounted %s at %s using %s", 
		config.TusdURL, config.MountPoint, provider.GetName())

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Mount is active. Press Ctrl+C to unmount.")
	
	<-sigChan
	logger.Info("Shutdown signal received, unmounting...")

	cancel()
	if err := provider.Unmount(); err != nil {
		logger.Errorf("Failed to unmount: %v", err)
		os.Exit(1)
	}

	logger.Info("Unmounted successfully")
}

func createProvider(mountType string, config *providers.Config, handler providers.FileSystemHandler) providers.Provider {
	switch strings.ToLower(mountType) {
	case "macos-fileprovider":
		// Use native macOS File Provider
		return providers.NewMacOSFileProvider(config, handler)
	case "windows-cloudfiles":
		// Use native Windows Cloud Files API
		return providers.NewWindowsCloudFiles(config, handler)
	case "windows-legacy":
		// Use Windows legacy provider for older versions
		return providers.NewWindowsLegacyProvider(config, handler)
	case "rclone", "rclone-sdk":
		return providers.NewRcloneSDKProvider(config, handler)
	case "webdav":
		// WebDAV 需要特殊处理，因为它不是真正的挂载
		fmt.Println("Warning: WebDAV mode is not implemented in mount mode, use sync mode with WebDAV server")
		return nil
	default:
		return nil
	}
}

func showPlatformCapabilities(logger *logrus.Logger) {
	info := providers.GetPlatformInfo()
	capabilities := providers.GetCapabilities(info)
	recommendations := providers.GetRecommendations(&providers.Config{})

	fmt.Printf("Platform Information:\n")
	fmt.Printf("  OS: %s %s (%s)\n", info.OS, info.Version, info.Arch)
	fmt.Printf("  Build: %s\n", info.Build)
	fmt.Printf("\n")

	fmt.Printf("Available Mount Capabilities:\n")
	for _, cap := range capabilities {
		levelStr := []string{"None", "Basic", "Native", "Optimal"}[cap.SupportLevel]
		fmt.Printf("  %-20s: %s support\n", cap.Name, levelStr)
		if cap.Description != "" {
			fmt.Printf("    Description: %s\n", cap.Description)
		}
		if len(cap.Prerequisites) > 0 {
			fmt.Printf("    Prerequisites: %s\n", strings.Join(cap.Prerequisites, ", "))
		}
		fmt.Printf("\n")
	}

	if len(recommendations) > 0 {
		fmt.Printf("Recommendations:\n")
		for _, rec := range recommendations {
			fmt.Printf("  - %s\n", rec)
		}
	}
}