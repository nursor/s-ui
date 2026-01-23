package frp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alireza0/s-ui/config"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
)

// FrpManager FRP进程管理器
type FrpManager struct {
	servers map[uint]*FrpInstance
	ctx     context.Context
	cancel  context.CancelFunc
}

// FrpInstance FRP实例
type FrpInstance struct {
	server     *model.FrpServer
	cmd        *exec.Cmd
	isRunning  bool
	configPath string
	logPath    string
	startTime  time.Time
}

// NewFrpManager 创建FRP管理器
func NewFrpManager() *FrpManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &FrpManager{
		servers: make(map[uint]*FrpInstance),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 启动FRP服务器
func (m *FrpManager) Start(server *model.FrpServer) error {
	// 检查是否已在运行
	if instance, exists := m.servers[server.Id]; exists && instance.isRunning {
		return fmt.Errorf("FRP server %s is already running", server.Name)
	}

	// 准备日志文件
	logPath := filepath.Join(GetFrpLogDir(), fmt.Sprintf("frp_%s_%d.log", server.Type, server.Id))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// 生成配置文件
	configPath, err := GenerateConfigFile(server)
	if err != nil {
		logFile.Close()
		return fmt.Errorf("failed to generate config: %v", err)
	}

	// 获取FRP二进制路径
	binaryPath := GetBinaryPath(server.Version, server.Arch, server.Type)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		logFile.Close()
		return fmt.Errorf("FRP binary not found: %s (please download first)", binaryPath)
	}

	// 创建命令
	args := []string{"-c", configPath}
	cmd := exec.CommandContext(m.ctx, binaryPath, args...)

	// 设置输出
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// 启动进程
	logger.Info("Starting FRP", server.Type, "server:", server.Name)
	logger.Debug("Command:", binaryPath, args)
	err = cmd.Start()
	if err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start FRP: %v", err)
	}

	// 创建实例
	instance := &FrpInstance{
		server:     server,
		cmd:        cmd,
		isRunning:  true,
		configPath: configPath,
		logPath:    logPath,
		startTime:  time.Now(),
	}

	m.servers[server.Id] = instance

	// 更新服务器信息
	server.Pid = cmd.Process.Pid
	server.Status = "running"
	server.ConfigPath = configPath
	server.LogPath = logPath
	server.LastRunAt = time.Now()

	logger.Infof("FRP %s server started: %s (PID: %d)", server.Type, server.Name, server.Pid)
	return nil
}

// Stop 停止FRP服务器
func (m *FrpManager) Stop(serverId uint) error {
	instance, exists := m.servers[serverId]
	if !exists {
		return fmt.Errorf("FRP server %d not found", serverId)
	}

	if !instance.isRunning {
		return nil
	}

	logger.Info("Stopping FRP", instance.server.Type, "server:", instance.server.Name)

	// 发送SIGTERM信号
	if instance.cmd.Process != nil {
		if err := instance.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			logger.Warning("Failed to send SIGTERM:", err)
		}
	}

	// 等待进程退出（最多5秒）
	done := make(chan error, 1)
	go func() {
		done <- instance.cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		// 超时，强制杀死
		if instance.cmd.Process != nil {
			logger.Warning("FRP server timeout, killing...")
			instance.cmd.Process.Kill()
		}
	case err := <-done:
		if err != nil {
			logger.Warning("FRP server exit error:", err)
		}
	}

	instance.isRunning = false
	delete(m.servers, serverId)

	logger.Info("FRP server stopped:", instance.server.Name)
	return nil
}

// Restart 重启FRP服务器
func (m *FrpManager) Restart(server *model.FrpServer) error {
	logger.Info("Restarting FRP", server.Type, "server:", server.Name)

	// 先停止
	if err := m.Stop(server.Id); err != nil {
		logger.Warning("Failed to stop FRP server:", err)
	}

	// 等待一小段时间
	time.Sleep(500 * time.Millisecond)

	// 再启动
	return m.Start(server)
}

// GetStatus 获取FRP服务器状态
func (m *FrpManager) GetStatus(serverId uint) string {
	instance, exists := m.servers[serverId]
	if !exists {
		return "stopped"
	}

	if !instance.isRunning {
		return "stopped"
	}

	// 检查进程是否还存活
	if instance.cmd.Process == nil {
		return "error"
	}

	if err := instance.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		instance.isRunning = false
		return "error"
	}

	return "running"
}

// GetUptime 获取运行时间
func (m *FrpManager) GetUptime(serverId uint) time.Duration {
	instance, exists := m.servers[serverId]
	if !exists || !instance.isRunning {
		return 0
	}
	return time.Since(instance.startTime)
}

// GetLogPath 获取日志文件路径
func (m *FrpManager) GetLogPath(serverId uint) string {
	instance, exists := m.servers[serverId]
	if !exists {
		return ""
	}
	return instance.logPath
}

// StopAll 停止所有FRP服务器
func (m *FrpManager) StopAll() {
	logger.Info("Stopping all FRP servers...")
	for id := range m.servers {
		if err := m.Stop(id); err != nil {
			logger.Warning("Failed to stop FRP server", id, ":", err)
		}
	}
}

// Shutdown 关闭管理器
func (m *FrpManager) Shutdown() {
	logger.Info("Shutting down FRP manager...")
	m.cancel()
	m.StopAll()
}

// GetFrpBinDir 获取FRP二进制文件目录
func GetFrpBinDir() string {
	dir := config.GetFrpBinDir()
	if dir == "" {
		dir = filepath.Join(config.GetDBDir(), "frp")
	}
	return dir
}

// GetFrpConfigDir 获取FRP配置文件目录
func GetFrpConfigDir() string {
	dir := config.GetFrpConfigDir()
	if dir == "" {
		dir = filepath.Join(config.GetDBDir(), "frp")
	}
	return dir
}

// GetFrpLogDir 获取FRP日志目录
func GetFrpLogDir() string {
	dir := config.GetFrpLogDir()
	if dir == "" {
		dir = filepath.Join(config.GetDBDir(), "frp")
	}
	return dir
}

// GetBinaryPath 获取FRP二进制文件路径
func GetBinaryPath(version, arch, frpType string) string {
	binDir := GetFrpBinDir()
	binaryName := "frps"
	if frpType == "client" {
		binaryName = "frpc"
	}
	return filepath.Join(binDir, fmt.Sprintf("%s_%s", binaryName, version))
}
