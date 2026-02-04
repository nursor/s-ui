package model

import (
	"encoding/json"
	"time"
)

// FrpServer FRP服务器配置表
type FrpServer struct {
	Id   uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" form:"name" gorm:"size:255;uniqueIndex"`
	Type string `json:"type" form:"type" gorm:"index"` // "server" 或 "client"

	// 进程管理
	Enable bool   `json:"enable" form:"enable" gorm:"index"`
	Status string `json:"status" form:"status" gorm:"default:'stopped'"` // running, stopped, error
	Pid    int    `json:"pid" form:"pid" gorm:"default:0"`

	// FRP配置
	Version string `json:"version" form:"version"` // 如 "v0.58.1"
	Arch    string `json:"arch" form:"arch"`       // 如 "linux_amd64"
	Port    int    `json:"port" form:"port"`       // 监听端口

	// 配置内容（TOML格式的JSON表示）
	Config json.RawMessage `json:"config,omitempty" form:"config" gorm:"type:json"`

	// 文件路径
	ConfigPath string `json:"config_path" form:"config_path" gorm:"column:config_path"`
	BinaryPath string `json:"binary_path" form:"binary_path" gorm:"column:binary_path"`
	LogPath    string `json:"log_path" form:"log_path" gorm:"column:log_path"`

	// 元数据
	CreatedAt time.Time  `json:"created_at" form:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" form:"updated_at"`
	LastRunAt *time.Time `json:"last_run_at,omitempty" form:"last_run_at" gorm:"default:null"`

	// 关联
	Proxies []FrpProxy `json:"proxies,omitempty" form:"proxies" gorm:"foreignKey:ServerId"`
}

// FrpProxy FRP代理配置表（用于frpc）
type FrpProxy struct {
	Id       uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	ServerId uint   `json:"server_id" form:"server_id" gorm:"index"`
	Name     string `json:"name" form:"name" gorm:"size:255;uniqueIndex"`
	Type     string `json:"type" form:"type"` // tcp, udp, http, https, stcp, xtcp
	Enable   bool   `json:"enable" form:"enable"`

	// 代理配置
	LocalIp       string   `json:"local_ip" form:"local_ip"`
	LocalPort     int      `json:"local_port" form:"local_port"`
	RemotePort    int      `json:"remote_port" form:"remote_port"`
	CustomDomains []string `json:"custom_domains" form:"custom_domains" gorm:"serializer:json"`

	// 扩展配置（JSON格式）
	Options json.RawMessage `json:"options,omitempty" form:"options" gorm:"type:json"`

	// 关联
	Server    *FrpServer `json:"server,omitempty" form:"server" gorm:"foreignKey:ServerId"`
	CreatedAt time.Time  `json:"created_at" form:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" form:"updated_at"`
}

// FrpLog FRP日志表
type FrpLog struct {
	Id        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ServerId  uint      `json:"server_id" form:"server_id" gorm:"index"`
	Level     string    `json:"level" form:"level" gorm:"index"` // info, warning, error
	Message   string    `json:"message" form:"message" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" form:"created_at" gorm:"index"`
}

// TableName 指定表名
func (FrpServer) TableName() string {
	return "sui_frp_servers"
}

func (FrpProxy) TableName() string {
	return "sui_frp_proxies"
}

func (FrpLog) TableName() string {
	return "sui_frp_logs"
}
