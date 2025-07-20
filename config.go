package main

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	TusdURL       string `yaml:"tusd_url"`
	WatchDir      string `yaml:"watch_dir"`
	LogLevel      string `yaml:"log_level"`
	ChunkSize     int64  `yaml:"chunk_size"`
	MaxRetries    int    `yaml:"max_retries"`
	RetryDelay    int    `yaml:"retry_delay_seconds"`
	
	// Mount configuration
	MountPoint    string `yaml:"mount_point"`
	MountType     string `yaml:"mount_type"`
	WebDAVPort    int    `yaml:"webdav_port"`
	BundleID      string `yaml:"bundle_id"`
	DisplayName   string `yaml:"display_name"`
	ProviderName  string `yaml:"provider_name"`
	SyncRootPath  string `yaml:"sync_root_path"`
	CacheDir      string `yaml:"cache_dir"`
	CacheSize     int64  `yaml:"cache_size"`
	ReadOnly      bool   `yaml:"read_only"`
	DebugMode     bool   `yaml:"debug_mode"`
	RcloneConfig  string `yaml:"rclone_config"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		TusdURL:      "http://localhost:1080/files/",
		WatchDir:     "./sync",
		LogLevel:     "info",
		ChunkSize:    1024 * 1024, // 1MB
		MaxRetries:   3,
		RetryDelay:   5,
		MountPoint:   "",
		MountType:    "auto",
		WebDAVPort:   8080,
		BundleID:     "com.mdriver.fileprovider",
		DisplayName:  "Tusd Sync",
		ProviderName: "TusdSync Provider",
		CacheSize:    1024 * 1024 * 1024, // 1GB
		ReadOnly:     false,
		DebugMode:    false,
	}

	if configPath == "" {
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return config, err
	}

	return config, nil
}