package frp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
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
	mu      sync.RWMutex
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
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已在运行
	if instance, exists := m.servers[server.Id]; exists && instance.isRunning {
		return fmt.Errorf("FRP server %s is already running", server.Name)
	}

	// 准备日志文件（使用配置名）
	logPath := filepath.Join(GetFrpLogDir(), fmt.Sprintf("%s.log", server.Name))

	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

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
	now := time.Now()
	server.LastRunAt = &now

	logger.Infof("FRP %s server started: %s (PID: %d)", server.Type, server.Name, server.Pid)
	return nil
}

// Stop 停止FRP服务器
func (m *FrpManager) Stop(serverId uint) error {
	m.mu.Lock()
	instance, exists := m.servers[serverId]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("FRP server %d not found", serverId)
	}

	if !instance.isRunning {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock() // 解锁以防止 Stop 阻塞其他操作，但要注意并发问题

	// 更好的做法是：先标记为停止，然后由外部逻辑处理，或者持有锁直到进程真正发送信号
	// 鉴于 Stop 可能耗时（KillWait），我们可以先从 map 中移除或标记，但这里简单起见，我们持有锁进行状态修改，释放锁进行等待
	// 为了安全，我们重新加锁去获取 instance 和状态

	m.mu.Lock()
	instance, exists = m.servers[serverId] // 再次检查
	if !exists || !instance.isRunning {
		m.mu.Unlock()
		return nil
	}

	logger.Info("Stopping FRP", instance.server.Type, "server:", instance.server.Name)

	// 发送SIGTERM信号
	if instance.cmd.Process != nil {
		if err := instance.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			logger.Warning("Failed to send SIGTERM:", err)
		}
	}

	// 这里有个问题，Wait 可能会阻塞，持有锁会导致死锁风险如果其他地方也在等锁
	// 所以我们应该在这里拷贝必要信息，然后解锁去等待，最后再加锁清理 map
	cmd := instance.cmd
	m.mu.Unlock()

	// 等待进程退出（最多5秒）
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		// 超时，强制杀死
		if cmd.Process != nil {
			logger.Warning("FRP server timeout, killing...")
			cmd.Process.Kill()
		}
	case err := <-done:
		if err != nil {
			logger.Warning("FRP server exit error:", err)
		}
	}

	m.mu.Lock()
	// 再次检查实例是否还存在（可能已经被再次启动了？）
	// 简单逻辑：只要 ID 匹配就清理
	if currentInstance, ok := m.servers[serverId]; ok && currentInstance == instance {
		currentInstance.isRunning = false
		delete(m.servers, serverId)
	}
	m.mu.Unlock()

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
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.servers[serverId]
	if !exists {
		// logger.Debugf("GetStatus: server %d not found in map", serverId)
		return "stopped"
	}

	if !instance.isRunning {
		// logger.Debugf("GetStatus: server %d isRunning=false", serverId)
		return "stopped"
	}

	// 检查进程是否还存活
	if instance.cmd.Process == nil {
		logger.Warningf("GetStatus: server %d process is nil", serverId)
		return "error"
	}

	// 这个 Signal(0) 检查在 Linux/Mac 上有效
	if err := instance.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		// 进程不存在或无权限?
		if err.Error() == "os: process already finished" || fmt.Sprint(err) == "os: process already finished" {
			// Go exec 包如果知道进程结束了会报这个
			// logger.Debugf("GetStatus: process finished: %v", err)
		} else {
			logger.Warningf("GetStatus: server %d Signal(0) failed: %v", serverId, err)
		}

		return "error"
	}

	return "running"
}

// GetUptime 获取运行时间
func (m *FrpManager) GetUptime(serverId uint) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.servers[serverId]
	if !exists || !instance.isRunning {
		return 0
	}
	return time.Since(instance.startTime)
}

// GetLogPath 获取日志文件路径
func (m *FrpManager) GetLogPath(serverId uint) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.servers[serverId]
	if !exists {
		return ""
	}
	return instance.logPath
}

// StopAll 停止所有FRP服务器
func (m *FrpManager) StopAll() {
	logger.Info("Stopping all FRP servers...")

	// 先获取所有 ID，避免在遍历时修改 map
	m.mu.RLock()
	var ids []uint
	for id := range m.servers {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	for _, id := range ids {
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
	// 直接使用 frps/frpc 文件名，不带版本号
	return filepath.Join(binDir, binaryName)
}
