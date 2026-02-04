package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/frp"
	"github.com/alireza0/s-ui/logger"
	"gorm.io/gorm"
)

// FrpService FRP业务服务层
type FrpService struct {
	manager *frp.FrpManager
}

// 全局FRP管理器实例
var globalFrpManager *frp.FrpManager

// init 初始化全局FRP管理器
func init() {
	globalFrpManager = frp.NewFrpManager()
}

// NewFrpService 创建FRP服务
func NewFrpService() *FrpService {
	return &FrpService{
		manager: globalFrpManager,
	}
}

// GetAll 获取所有FRP服务器
func (s *FrpService) GetAll() ([]model.FrpServer, error) {
	db := database.GetDB()
	var servers []model.FrpServer
	err := db.Model(model.FrpServer{}).Find(&servers).Error
	return servers, err
}

// Get 获取单个FRP服务器详情（包含代理配置）
func (s *FrpService) Get(id uint) (*model.FrpServer, error) {
	db := database.GetDB()
	var server model.FrpServer
	err := db.Model(model.FrpServer{}).Preload("Proxies").First(&server, id).Error
	if err != nil {
		return nil, err
	}

	// 如果是客户端，获取所有启用的代理
	if server.Type == "client" {
		var proxies []model.FrpProxy
		err = db.Model(model.FrpProxy{}).Where("server_id = ?", id).Find(&proxies).Error
		if err == nil {
			server.Proxies = proxies
		}
	}

	return &server, err
}

// SaveServer 保存FRP服务器配置
func (s *FrpService) SaveServer(tx *gorm.DB, action string, data json.RawMessage) error {
	var err error

	switch action {
	case "new", "edit":
		var server model.FrpServer
		err = json.Unmarshal(data, &server)
		if err != nil {
			return err
		}

		// 验证必填字段
		if server.Name == "" {
			return fmt.Errorf("配置名称不能为空")
		}
		if server.Type == "" {
			return fmt.Errorf("类型不能为空")
		}

		// 设置默认版本
		if server.Version == "" {
			server.Version = frp.DefaultVersion
		}

		// 如果是编辑，先停止服务器
		if action == "edit" && server.Id > 0 {
			oldServer, err := s.Get(server.Id)
			if err == nil && oldServer.Status == "running" {
				s.manager.Stop(server.Id)
			}
		}

		err = tx.Save(&server).Error

	case "del":
		var id uint
		err = json.Unmarshal(data, &id)
		if err != nil {
			return err
		}

		// 先停止服务器
		s.manager.Stop(id)

		// 删除关联的代理配置
		tx.Where("server_id = ?", id).Delete(&model.FrpProxy{})

		// 删除服务器
		err = tx.Delete(&model.FrpServer{}, id).Error
	}

	return err
}

// StartServer 启动FRP服务器
func (s *FrpService) StartServer(id uint) error {
	db := database.GetDB()
	var server model.FrpServer
	err := db.Model(model.FrpServer{}).First(&server, id).Error
	if err != nil {
		return err
	}

	// 检查FRP是否已安装
	if !frp.IsInstalled(server.Version, server.Arch, server.Type) {
		logger.Info("FRP not installed, downloading...", server.Type, server.Version)

		// 下载FRP
		downloader := frp.NewDownloader(server.Version)
		if err := downloader.Download(server.Type); err != nil {
			return err
		}

		logger.Info("FRP downloaded successfully")
	}

	// 如果是客户端，加载代理配置
	var proxies []model.FrpProxy
	if server.Type == "client" {
		err = db.Model(model.FrpProxy{}).
			Where("server_id = ? AND enable = ?", id, true).
			Find(&proxies).Error
		if err != nil {
			return err
		}
	}

	// 更新服务器配置（包含代理）
	server.Proxies = proxies

	// 启动进程
	err = s.manager.Start(&server)
	if err != nil {
		// 更新状态为错误
		db.Model(&server).Updates(map[string]interface{}{
			"status": "error",
		})
		return err
	}

	// 更新数据库状态
	return db.Model(&server).Updates(map[string]interface{}{
		"status":      "running",
		"pid":         server.Pid,
		"last_run_at": server.LastRunAt,
	}).Error
}

// StopServer 停止FRP服务器
func (s *FrpService) StopServer(id uint) error {
	err := s.manager.Stop(id)
	if err != nil {
		return err
	}

	db := database.GetDB()
	return db.Model(model.FrpServer{}).
		Where("id = ?", id).
		Update("status", "stopped").Error
}

// RestartServer 重启FRP服务器
func (s *FrpService) RestartServer(id uint) error {
	logger.Info("Restarting FRP server:", id)

	// 先停止
	s.manager.Stop(id)

	// 等待一小段时间
	// 这里不能用time.Sleep因为要避免import cycle

	// 再启动
	return s.StartServer(id)
}

// GetServerStatus 获取服务器状态
func (s *FrpService) GetServerStatus(id uint) (map[string]interface{}, error) {
	db := database.GetDB()
	var server model.FrpServer
	err := db.Model(model.FrpServer{}).First(&server, id).Error
	if err != nil {
		return nil, err
	}

	status := s.manager.GetStatus(id)
	uptime := s.manager.GetUptime(id)

	return map[string]interface{}{
		"id":      server.Id,
		"name":    server.Name,
		"type":    server.Type,
		"status":  status,
		"pid":     server.Pid,
		"uptime":  uptime.Seconds(),
		"enable":  server.Enable,
		"version": server.Version,
		"arch":    server.Arch,
	}, nil
}

// GetAllProxies 获取所有代理配置
func (s *FrpService) GetAllProxies(serverId uint) ([]model.FrpProxy, error) {
	db := database.GetDB()
	var proxies []model.FrpProxy
	err := db.Model(model.FrpProxy{}).Where("server_id = ?", serverId).Find(&proxies).Error
	return proxies, err
}

// SaveProxy 保存代理配置
func (s *FrpService) SaveProxy(tx *gorm.DB, action string, data json.RawMessage) error {
	var err error

	switch action {
	case "new", "edit":
		var proxy model.FrpProxy
		err = json.Unmarshal(data, &proxy)
		if err != nil {
			return err
		}

		err = tx.Save(&proxy).Error

		// 如果关联的服务器正在运行，重启它
		if proxy.ServerId > 0 {
			var server model.FrpServer
			tx.Model(model.FrpServer{}).First(&server, proxy.ServerId)
			if server.Status == "running" {
				go func(id uint) {
					s.RestartServer(id)
				}(proxy.ServerId)
			}
		}

	case "del":
		var id uint
		err = json.Unmarshal(data, &id)
		if err != nil {
			return err
		}

		var proxy model.FrpProxy
		tx.Model(model.FrpProxy{}).First(&proxy, id)

		err = tx.Delete(&model.FrpProxy{}, id).Error

		// 如果关联的服务器正在运行，重启它
		if proxy.ServerId > 0 {
			var server model.FrpServer
			tx.Model(model.FrpServer{}).First(&server, proxy.ServerId)
			if server.Status == "running" {
				go func(id uint) {
					s.RestartServer(id)
				}(proxy.ServerId)
			}
		}
	}

	return err
}

// DownloadFRP 下载FRP二进制文件
func (s *FrpService) DownloadFRP(version string, frpType string) error {
	downloader := frp.NewDownloader(version)
	return downloader.Download(frpType)
}

// GetLogs 获取FRP日志（从文件读取）
func (s *FrpService) GetLogs(serverId uint, limit int) ([]model.FrpLog, error) {
	db := database.GetDB()
	var server model.FrpServer
	if err := db.Model(model.FrpServer{}).First(&server, serverId).Error; err != nil {
		return nil, err
	}

	logPath := filepath.Join(frp.GetFrpLogDir(), fmt.Sprintf("%s.log", server.Name))

	file, err := os.Open(logPath)
	if err != nil {
		// 文件不存在视为无日志，返回空列表而不是错误
		return []model.FrpLog{}, nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// 取最后 limit 行
	total := len(lines)
	start := total - limit
	if start < 0 {
		start = 0
	}

	// 构造 FrpLog 对象
	var logs []model.FrpLog
	// 倒序返回，最新的在前面
	for i := total - 1; i >= start; i-- {
		logs = append(logs, model.FrpLog{
			Id:        uint(i + 1),
			ServerId:  serverId,
			Message:   lines[i],
			CreatedAt: time.Now(), // 这里的相关时间暂时无法准确获取
			Level:     "info",
		})
	}

	return logs, nil
}

// GetLogPath 获取日志文件路径
func (s *FrpService) GetLogPath(serverId uint) string {
	return s.manager.GetLogPath(serverId)
}

// GetRuntimeStatus 获取运行时状态（不查询数据库）
func (s *FrpService) GetRuntimeStatus(serverId uint) string {
	return s.manager.GetStatus(serverId)
}
