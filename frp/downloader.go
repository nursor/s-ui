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
	DefaultVersion = "v0.67.0"
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
	// 规范化架构名称
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

	// 移除版本号中的 'v' 前缀
	versionNoV := strings.TrimPrefix(d.version, "v")
	return fmt.Sprintf("frp_%s_%s_%s.tar.gz", versionNoV, d.osType, arch)
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

	// 确定二进制文件名 (frps 或 frpc)
	actualBinaryName := "frps"
	if frpType == "client" {
		actualBinaryName = "frpc"
	}

	// 构造解压后的目录名 (frp_version_os_arch)
	tarballName := d.getTarballName(frpType)
	folderName := strings.TrimSuffix(tarballName, ".tar.gz")

	// 源文件路径: extractDir/frp_0.xx.x_linux_amd64/frps
	sourcePath := filepath.Join(extractDir, folderName, actualBinaryName)

	// 目标文件路径
	// 直接放到 binDir/frps
	destPath := filepath.Join(binDir, actualBinaryName)

	// 强制覆盖旧文件（如果有）
	if _, err := os.Stat(destPath); err == nil {
		os.Remove(destPath)
	}

	// 检查源文件是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// 尝试其他可能的路径 (兜底逻辑)
		logger.Warningf("Binary not found at standard path: %s, searching...", sourcePath)
		files, _ := os.ReadDir(extractDir)
		for _, f := range files {
			if f.IsDir() && strings.HasPrefix(f.Name(), "frp_") {
				possiblePath := filepath.Join(extractDir, f.Name(), actualBinaryName)
				if _, err := os.Stat(possiblePath); !os.IsNotExist(err) {
					sourcePath = possiblePath
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
	// 通过执行命令获取版本
	// frps --version
	binaryPath := GetBinaryPath("", "", frpType) // 版本架构参数不影响路径生成

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no installed %s found", frpType)
	}

	cmd := exec.Command(binaryPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %v", err)
	}

	// 输出通常是 "0.58.1" 或 "frps version 0.58.1"
	// 简单的处理：去除空白字符
	version := strings.TrimSpace(string(out))
	return version, nil
}

// runCommand 执行命令
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
