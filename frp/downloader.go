package frp

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alireza0/s-ui/logger"
)

const (
	// FRPGitHubURL FRP GitHub下载地址
	FRPGitHubURL = "https://github.com/fatedier/frp/releases/download"
	// DefaultVersion 默认FRP版本
	DefaultVersion = "v0.58.1"
)

// Downloader FRP下载器
type Downloader struct {
	version string
	arch    string
	osType  string
}

// NewDownloader 创建下载器
func NewDownloader(version string) *Downloader {
	if version == "" {
		version = DefaultVersion
	}

	return &Downloader{
		version: version,
		arch:    runtime.GOARCH,
		osType:  runtime.GOOS,
	}
}

// Download 下载FRP二进制文件
func (d *Downloader) Download(frpType string) error {
	// 构造下载URL
	tarball := d.getTarballName(frpType)
	url := fmt.Sprintf("%s/%s/%s", FRPGitHubURL, d.version, tarball)

	logger.Infof("Downloading FRP %s from: %s", frpType, url)

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download FRP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status code: %d", resp.StatusCode)
	}

	// 创建临时文件
	tempPath := filepath.Join(os.TempDir(), tarball)
	outFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer outFile.Close()

	// 写入下载内容
	logger.Info("Downloading to:", tempPath)
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	// 解压文件
	logger.Info("Extracting...")
	if err := d.extract(tempPath, frpType); err != nil {
		return fmt.Errorf("failed to extract: %v", err)
	}

	// 清理临时文件
	os.Remove(tempPath)

	logger.Infof("FRP %s downloaded successfully", frpType)
	return nil
}

// getTarballName 获取压缩包文件名
func (d *Downloader) getTarballName(frpType string) string {
	// 规范化架构名���
	arch := d.arch
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	} else if arch == "386" {
		arch = "386"
	} else if arch == "arm" {
		arch = "arm"
	}

	return fmt.Sprintf("frp_%s_%s_%s.tar.gz", d.version, d.osType, arch)
}

// extract 解压tar.gz文件
func (d *Downloader) extract(tarPath, frpType string) error {
	// 确保目标目录存在
	binDir := GetFrpBinDir()
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %v", err)
	}

	// 使用tar命令解压（更可靠）
	extractDir := filepath.Join(os.TempDir(), fmt.Sprintf("frp_extract_%s", frpType))
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %v", err)
	}
	defer os.RemoveAll(extractDir)

	// 解压命令
	cmd := fmt.Sprintf("cd %s && tar -xzf %s", extractDir, tarPath)
	if err := runCommand("sh", "-c", cmd); err != nil {
		return fmt.Errorf("failed to extract archive: %v", err)
	}

	// 查找并复制可执行文件
	binaryName := fmt.Sprintf("frp_%s", frpType)
	sourcePath := filepath.Join(extractDir, binaryName, frpType)
	destPath := filepath.Join(binDir, fmt.Sprintf("%s_%s", frpType, d.version))

	// 检查源文件是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// 尝试其他可能的路径
		files, _ := os.ReadDir(extractDir)
		for _, f := range files {
			if f.IsDir() && strings.HasPrefix(f.Name(), "frp_") {
				sourcePath = filepath.Join(extractDir, f.Name(), frpType)
				if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
					break
				}
			}
		}
	}

	// 复制文件
	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy binary: %v", err)
	}

	// 设置可执行权限
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permission: %v", err)
	}

	logger.Infof("FRP binary installed to: %s", destPath)
	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// IsInstalled 检查FRP是否已安装
func IsInstalled(version, arch, frpType string) bool {
	binaryPath := GetBinaryPath(version, arch, frpType)
	_, err := os.Stat(binaryPath)
	return !os.IsNotExist(err)
}

// GetInstalledVersion 获取已安装的FRP版本
func GetInstalledVersion(frpType string) (string, error) {
	binDir := GetFrpBinDir()
	prefix := fmt.Sprintf("%s_", frpType)

	files, err := os.ReadDir(binDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) {
			// 提取版本号
			version := strings.TrimPrefix(file.Name(), prefix)
			return version, nil
		}
	}

	return "", fmt.Errorf("no installed %s found", frpType)
}

// runCommand 执行命令
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
