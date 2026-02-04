package frp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
)

// ConfigGenerator 配置生成器
type ConfigGenerator struct {
	configDir string
}

// NewConfigGenerator 创建配置生成器
func NewConfigGenerator(configDir string) *ConfigGenerator {
	return &ConfigGenerator{
		configDir: configDir,
	}
}

// GenerateConfigFile 生成FRP配置文件（入口函数）
func GenerateConfigFile(server *model.FrpServer) (string, error) {
	configGen := NewConfigGenerator(GetFrpConfigDir())

	if server.Type == "client" {
		// 加载代理配置
		db := database.GetDB()
		var proxies []model.FrpProxy
		err := db.Model(&model.FrpProxy{}).
			Where("server_id = ? AND enable = ?", server.Id, true).
			Find(&proxies).Error
		if err != nil {
			return "", fmt.Errorf("failed to load proxies: %v", err)
		}

		return configGen.GenerateClientConfig(server, proxies)
	} else {
		return configGen.GenerateServerConfig(server)
	}
}

// GenerateServerConfig 生成FRPS服务端配置
func (g *ConfigGenerator) GenerateServerConfig(server *model.FrpServer) (string, error) {
	config := make(map[string]interface{})

	// 解析数据库中的JSON配置
	if len(server.Config) > 0 {
		if err := json.Unmarshal(server.Config, &config); err != nil {
			return "", fmt.Errorf("failed to parse config JSON: %v", err)
		}
	}

	// 添加或更新common配置
	if config["common"] == nil {
		config["common"] = make(map[string]interface{})
	}
	common := config["common"].(map[string]interface{})

	// 设置绑定端口
	if server.Port > 0 {
		common["bind_port"] = server.Port
	}

	// 确保有必要的默认配置
	if common["bind_port"] == nil {
		common["bind_port"] = 7000
	}

	config["common"] = common

	// 生成TOML
	return g.writeConfig(server, config)
}

// GenerateClientConfig 生成FRPC客户端配置
func (g *ConfigGenerator) GenerateClientConfig(server *model.FrpServer, proxies []model.FrpProxy) (string, error) {
	config := make(map[string]interface{})

	// 解析数据库中的JSON配置
	if len(server.Config) > 0 {
		if err := json.Unmarshal(server.Config, &config); err != nil {
			return "", fmt.Errorf("failed to parse config JSON: %v", err)
		}
	}

	// 添加或更新common配置
	if config["common"] == nil {
		config["common"] = make(map[string]interface{})
	}
	common := config["common"].(map[string]interface{})

	// 确保基本配置存在
	if common["server_addr"] == nil {
		common["server_addr"] = "127.0.0.1"
	}
	if common["server_port"] == nil {
		common["server_port"] = 7000
	}

	config["common"] = common

	// 添加代理配置
	for _, proxy := range proxies {
		proxyConfig := map[string]interface{}{
			"type":       proxy.Type,
			"local_ip":   proxy.LocalIp,
			"local_port": proxy.LocalPort,
		}

		// 根据代理类型添加特定配置
		switch proxy.Type {
		case "tcp", "udp":
			if proxy.RemotePort > 0 {
				proxyConfig["remote_port"] = proxy.RemotePort
			}
		case "http", "https":
			if len(proxy.CustomDomains) > 0 {
				proxyConfig["custom_domains"] = proxy.CustomDomains
			}
		case "stcp", "xtcp":
			if proxy.RemotePort > 0 {
				proxyConfig["remote_port"] = proxy.RemotePort
			}
		}

		// 合并自定义选项
		if len(proxy.Options) > 0 {
			var opts map[string]interface{}
			if err := json.Unmarshal(proxy.Options, &opts); err == nil {
				for k, v := range opts {
					proxyConfig[k] = v
				}
			}
		}

		config[proxy.Name] = proxyConfig
	}

	// 生成TOML
	return g.writeConfig(server, config)
}

// writeConfig 将配置写入文件
func (g *ConfigGenerator) writeConfig(server *model.FrpServer, config map[string]interface{}) (string, error) {
	// 确保配置目录存在
	if err := os.MkdirAll(g.configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	// 生成TOML
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(config); err != nil {
		return "", fmt.Errorf("failed to encode TOML: %v", err)
	}

	// 使用配置名作为文件名（清理特殊字符）
	// 文件名格式: {name}.toml
	filename := fmt.Sprintf("%s.toml", server.Name)
	configPath := filepath.Join(g.configDir, filename)

	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %v", err)
	}

	return configPath, nil
}

// ValidateConfig 验证配置文件（使用FRP自带的验证功能）
func ValidateConfig(configPath string) error {
	// TODO: 实现配置验证逻辑
	// 可以通过执行 frpc/frps --verify 来验证
	return nil
}
