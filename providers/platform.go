package providers

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// PlatformInfo 平台信息详细结构
type PlatformInfo struct {
	OS           string // "darwin", "windows", "linux"
	Version      string // 操作系统版本
	Build        string // 构建号
	Arch         string // "amd64", "arm64", 等
	MajorVersion int    // 主版本号
	MinorVersion int    // 次版本号
}

// GetPlatformInfo 获取详细平台信息
func GetPlatformInfo() PlatformInfo {
	info := PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	switch runtime.GOOS {
	case "darwin":
		info = getMacOSInfo(info)
	case "windows":
		info = getWindowsInfo(info)
	case "linux":
		info = getLinuxInfo(info)
	}

	// 解析版本号
	parseVersion(&info)

	return info
}

// getMacOSInfo 获取 macOS 信息
func getMacOSInfo(info PlatformInfo) PlatformInfo {
	// 获取 macOS 版本
	cmd := exec.Command("sw_vers", "-productVersion")
	if output, err := cmd.Output(); err == nil {
		info.Version = strings.TrimSpace(string(output))
	}

	// 获取构建版本
	cmd = exec.Command("sw_vers", "-buildVersion")
	if output, err := cmd.Output(); err == nil {
		info.Build = strings.TrimSpace(string(output))
	}

	return info
}

// getWindowsInfo 获取 Windows 信息
func getWindowsInfo(info PlatformInfo) PlatformInfo {
	// 使用 ver 命令获取版本
	cmd := exec.Command("cmd", "/c", "ver")
	if output, err := cmd.Output(); err == nil {
		verStr := string(output)
		// 解析类似 "Microsoft Windows [Version 10.0.19044.1889]" 的输出
		if idx := strings.Index(verStr, "Version "); idx != -1 {
			start := idx + len("Version ")
			if end := strings.Index(verStr[start:], "]"); end != -1 {
				info.Version = verStr[start : start+end]
			}
		}
	}

	// 尝试使用 PowerShell 获取更详细信息
	cmd = exec.Command("powershell", "-Command", "(Get-ItemProperty 'HKLM:SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion').ReleaseId")
	if output, err := cmd.Output(); err == nil {
		info.Build = strings.TrimSpace(string(output))
	}

	return info
}

// getLinuxInfo 获取 Linux 信息
func getLinuxInfo(info PlatformInfo) PlatformInfo {
	// 尝试从 /etc/os-release 获取信息
	cmd := exec.Command("sh", "-c", "cat /etc/os-release | grep VERSION_ID | cut -d= -f2 | tr -d '\"'")
	if output, err := cmd.Output(); err == nil {
		info.Version = strings.TrimSpace(string(output))
	}

	// 获取内核版本
	cmd = exec.Command("uname", "-r")
	if output, err := cmd.Output(); err == nil {
		info.Build = strings.TrimSpace(string(output))
	}

	return info
}

// parseVersion 解析版本号
func parseVersion(info *PlatformInfo) {
	if info.Version == "" {
		return
	}

	parts := strings.Split(info.Version, ".")
	if len(parts) >= 1 {
		if major, err := strconv.Atoi(parts[0]); err == nil {
			info.MajorVersion = major
		}
	}
	if len(parts) >= 2 {
		if minor, err := strconv.Atoi(parts[1]); err == nil {
			info.MinorVersion = minor
		}
	}
}

// SupportLevel 支持级别
type SupportLevel int

const (
	SupportLevelNone SupportLevel = iota
	SupportLevelBasic
	SupportLevelNative
	SupportLevelOptimal
)

// Capability 挂载能力信息
type Capability struct {
	Name          string
	SupportLevel  SupportLevel
	RequiredOS    string
	MinVersion    string
	MaxVersion    string
	Description   string
	Prerequisites []string
}

// GetCapabilities 获取当前平台的挂载能力
func GetCapabilities(info PlatformInfo) []Capability {
	capabilities := []Capability{}

	switch info.OS {
	case "darwin":
		capabilities = append(capabilities, getMacOSCapabilities(info)...)
	case "windows":
		capabilities = append(capabilities, getWindowsCapabilities(info)...)
	case "linux":
		capabilities = append(capabilities, getLinuxCapabilities(info)...)
	}

	// 添加通用的 rclone SDK 支持
	capabilities = append(capabilities, Capability{
		Name:          "rclone-sdk",
		SupportLevel:  SupportLevelOptimal,
		RequiredOS:    "any",
		Description:   "Universal mounting solution using rclone SDK",
		Prerequisites: []string{"FUSE support (Linux/macOS)", "WinFsp (Windows)"},
	})

	return capabilities
}

// getMacOSCapabilities 获取 macOS 挂载能力
func getMacOSCapabilities(info PlatformInfo) []Capability {
	capabilities := []Capability{}

	// File Provider API (macOS 10.15+ or macOS 11+)
	// macOS versions: 10.15 = Catalina, 11+ = Big Sur and later, 15+ = Sequoia
	if (info.MajorVersion == 10 && info.MinorVersion >= 15) || info.MajorVersion >= 11 {
		capabilities = append(capabilities, Capability{
			Name:          "macos-fileprovider",
			SupportLevel:  SupportLevelNative,
			RequiredOS:    "darwin",
			MinVersion:    "10.15.0",
			Description:   "Native macOS File Provider API integration",
			Prerequisites: []string{"Xcode or Command Line Tools", "System Extensions allowed"},
		})
	}

	// macFUSE support
	capabilities = append(capabilities, Capability{
		Name:          "macfuse",
		SupportLevel:  SupportLevelOptimal,
		RequiredOS:    "darwin",
		Description:   "macFUSE-based mounting",
		Prerequisites: []string{"macFUSE installed", "FUSE for macOS"},
	})

	return capabilities
}

// getWindowsCapabilities 获取 Windows 挂载能力
func getWindowsCapabilities(info PlatformInfo) []Capability {
	capabilities := []Capability{}

	// Cloud Files API (Windows 10 1809+)
	if info.MajorVersion >= 10 {
		// 检查构建号
		buildNum := parseBuildNumber(info.Version)
		if buildNum >= 17763 { // Windows 10 1809
			capabilities = append(capabilities, Capability{
				Name:          "windows-cloudfiles",
				SupportLevel:  SupportLevelNative,
				RequiredOS:    "windows",
				MinVersion:    "10.0.17763",
				Description:   "Native Windows Cloud Files API for Smart Folders",
				Prerequisites: []string{"Windows 10 1809+", "Visual Studio C++ build tools", "Windows SDK"},
			})
		}
	}

	// WinFsp support
	capabilities = append(capabilities, Capability{
		Name:          "winfsp",
		SupportLevel:  SupportLevelOptimal,
		RequiredOS:    "windows",
		Description:   "WinFsp-based mounting",
		Prerequisites: []string{"WinFsp installed"},
	})

	// Legacy Windows support (XP/Vista/7/8/Early 10)
	capabilities = append(capabilities, Capability{
		Name:          "windows-legacy",
		SupportLevel:  SupportLevelBasic,
		RequiredOS:    "windows",
		MinVersion:    "5.1", // Windows XP
		Description:   "Legacy Windows compatibility with basic file sync",
		Prerequisites: []string{"Windows XP or later"},
	})

	return capabilities
}

// getLinuxCapabilities 获取 Linux 挂载能力
func getLinuxCapabilities(info PlatformInfo) []Capability {
	capabilities := []Capability{}

	// FUSE support
	capabilities = append(capabilities, Capability{
		Name:          "linux-fuse",
		SupportLevel:  SupportLevelOptimal,
		RequiredOS:    "linux",
		Description:   "FUSE-based mounting",
		Prerequisites: []string{"fuse package installed", "fuse kernel module loaded"},
	})

	return capabilities
}

// parseBuildNumber 从版本字符串解析构建号
func parseBuildNumber(version string) int {
	parts := strings.Split(version, ".")
	if len(parts) >= 3 {
		if build, err := strconv.Atoi(parts[2]); err == nil {
			return build
		}
	}
	return 0
}

// SelectBestProvider 选择最佳挂载提供者
func SelectBestProvider(config *Config, handler FileSystemHandler) Provider {
	info := GetPlatformInfo()
	capabilities := GetCapabilities(info)

	// 按优先级排序提供者
	var bestProvider Provider
	bestLevel := SupportLevelNone

	for _, cap := range capabilities {
		var provider Provider

		switch cap.Name {
		case "macos-fileprovider":
			if cap.SupportLevel > bestLevel {
				// macOS provider only available on macOS builds
				provider = NewRcloneSDKProvider(config, handler) // fallback
			}
		case "windows-cloudfiles":
			if cap.SupportLevel > bestLevel {
				// Windows Cloud Files provider (requires Windows build)
				provider = NewRcloneSDKProvider(config, handler) // fallback
			}
		case "windows-legacy":
			// Only use legacy provider if no better option is available
			if cap.SupportLevel > bestLevel || (bestProvider == nil && info.OS == "windows") {
				// Windows Legacy provider (requires Windows build)
				provider = NewRcloneSDKProvider(config, handler) // fallback
			}
		case "rclone-sdk", "macfuse", "linux-fuse", "winfsp":
			if cap.SupportLevel > bestLevel || bestProvider == nil {
				provider = NewRcloneSDKProvider(config, handler)
			}
		}

		if provider != nil && provider.IsSupported() && cap.SupportLevel > bestLevel {
			bestProvider = provider
			bestLevel = cap.SupportLevel
		}
	}

	// 如果没有找到合适的提供者，默认使用 rclone SDK
	if bestProvider == nil {
		bestProvider = NewRcloneSDKProvider(config, handler)
	}

	return bestProvider
}

// GetRecommendations 获取挂载建议
func GetRecommendations(config *Config) []string {
	info := GetPlatformInfo()
	capabilities := GetCapabilities(info)
	recommendations := []string{}

	switch info.OS {
	case "darwin":
		if (info.MajorVersion == 10 && info.MinorVersion >= 15) || info.MajorVersion >= 11 {
			recommendations = append(recommendations,
				"推荐使用 macOS File Provider API 获得最佳集成体验",
				"如需高性能，可使用 rclone SDK + macFUSE",
				"使用 go build -tags fileprovider 启用 File Provider API",
			)
		} else {
			recommendations = append(recommendations,
				"当前 macOS 版本过低，建议升级到 10.15+ 以支持 File Provider API",
				"可使用 rclone SDK + macFUSE 作为替代方案",
			)
		}

	case "windows":
		buildNum := parseBuildNumber(info.Version)
		if info.MajorVersion >= 10 && buildNum >= 17763 {
			recommendations = append(recommendations,
				"推荐使用 Windows Cloud Files API 创建 Smart Folder",
				"使用 go build -tags cloudfiles 启用 Cloud Files API",
				"需要 Visual Studio C++ 构建工具和 Windows SDK",
				"提供原生 Windows Explorer 集成和文件占位符功能",
			)
		} else {
			recommendations = append(recommendations,
				"当前 Windows 版本不支持 Cloud Files API",
				"建议安装 WinFsp 并使用 rclone SDK",
			)
		}

	case "linux":
		recommendations = append(recommendations,
			"Linux 平台推荐使用 rclone SDK + FUSE 方案",
			"确保已安装 fuse 包：sudo apt install fuse (Ubuntu/Debian) 或 sudo yum install fuse (RHEL/CentOS)",
		)
	}

	// rclone SDK 无需额外安装

	// 添加通用建议
	for _, cap := range capabilities {
		if len(cap.Prerequisites) > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("%s 需要满足以下前提条件：%s", cap.Name, strings.Join(cap.Prerequisites, ", ")),
			)
		}
	}

	return recommendations
}
