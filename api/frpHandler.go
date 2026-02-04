package api

import (
	"encoding/json"
	"strconv"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/logger"
	"github.com/gin-gonic/gin"
)

// GetFrpServers 获取所有FRP服务器
func (a *ApiService) GetFrpServers(c *gin.Context) {
	servers, err := a.FrpService.GetAll()
	if err != nil {
		jsonMsg(c, "获取服务器列表失败", err)
		return
	}
	jsonObj(c, servers, nil)
}

// GetFrpServer 获取单个FRP服务器详情
func (a *ApiService) GetFrpServer(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonMsg(c, "无效的ID", err)
		return
	}

	server, err := a.FrpService.Get(uint(id))
	if err != nil {
		jsonMsg(c, "获取服务器失败", err)
		return
	}
	jsonObj(c, server, nil)
}

// SaveFrpServer 保存FRP服务器配置
func (a *ApiService) SaveFrpServer(c *gin.Context, loginUser string) {
	action := c.Request.FormValue("action")
	data := c.Request.FormValue("data")

	err := a.FrpService.SaveServer(
		database.GetDB(),
		action,
		json.RawMessage(data),
	)
	if err != nil {
		jsonMsg(c, "保存失败", err)
		return
	}

	logger.Infof("用户 %s 保存了FRP服务器配置 action=%s", loginUser, action)

	// 返回更新后的服务列表
	err = a.LoadPartialData(c, []string{"services"})
	if err != nil {
		jsonMsg(c, "保存成功但获取列表失败", err)
	}
}

// StartFrpServer 启动FRP服务器
func (a *ApiService) StartFrpServer(c *gin.Context, loginUser string) {
	idStr := c.Request.FormValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonMsg(c, "无效的ID", err)
		return
	}

	err = a.FrpService.StartServer(uint(id))
	if err != nil {
		jsonMsg(c, "启动失败", err)
		return
	}

	logger.Infof("用户 %s 启动了FRP服务器 ID=%d", loginUser, id)
	jsonMsg(c, "启动成功", nil)
}

// StopFrpServer 停止FRP服务器
func (a *ApiService) StopFrpServer(c *gin.Context, loginUser string) {
	idStr := c.Request.FormValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonMsg(c, "无效的ID", err)
		return
	}

	err = a.FrpService.StopServer(uint(id))
	if err != nil {
		jsonMsg(c, "停止失败", err)
		return
	}

	logger.Infof("用户 %s 停止了FRP服务器 ID=%d", loginUser, id)
	jsonMsg(c, "停止成功", nil)
}

// RestartFrpServer 重启FRP服务器
func (a *ApiService) RestartFrpServer(c *gin.Context, loginUser string) {
	idStr := c.Request.FormValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonMsg(c, "无效的ID", err)
		return
	}

	err = a.FrpService.RestartServer(uint(id))
	if err != nil {
		jsonMsg(c, "重启失败", err)
		return
	}

	logger.Infof("用户 %s 重启了FRP服务器 ID=%d", loginUser, id)
	jsonMsg(c, "重启成功", nil)
}

// GetFrpServerStatus 获取FRP服务器状态
func (a *ApiService) GetFrpServerStatus(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonMsg(c, "无效的ID", err)
		return
	}

	status, err := a.FrpService.GetServerStatus(uint(id))
	if err != nil {
		jsonMsg(c, "获取状态失败", err)
		return
	}

	jsonObj(c, status, nil)
}

// GetFrpLogs 获取FRP日志
func (a *ApiService) GetFrpLogs(c *gin.Context) {
	idStr := c.Query("server_id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonMsg(c, "无效的ID", err)
		return
	}

	limitStr := c.Query("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	logs, err := a.FrpService.GetLogs(uint(id), limit)
	if err != nil {
		jsonMsg(c, "获取日志失败", err)
		return
	}

	jsonObj(c, logs, nil)
}

// SaveFrpProxy 保存FRP代理配置
func (a *ApiService) SaveFrpProxy(c *gin.Context, loginUser string) {
	action := c.Request.FormValue("action")
	data := c.Request.FormValue("data")

	err := a.FrpService.SaveProxy(
		database.GetDB(),
		action,
		json.RawMessage(data),
	)
	if err != nil {
		jsonMsg(c, "保存代理失败", err)
		return
	}

	logger.Infof("用户 %s 保存了FRP代理配置 action=%s", loginUser, action)
	jsonMsg(c, "保存成功", nil)
}

// DownloadFrp 下载FRP二进制文件
func (a *ApiService) DownloadFrp(c *gin.Context, loginUser string) {
	version := c.Request.FormValue("version")
	frpType := c.Request.FormValue("type")

	if version == "" {
		version = "v0.58.1" // 默认版本
	}

	err := a.FrpService.DownloadFRP(version, frpType)
	if err != nil {
		jsonMsg(c, "下载失败", err)
		return
	}

	logger.Infof("用户 %s 下载了FRP %s %s", loginUser, version, frpType)
	jsonMsg(c, "下载成功", nil)
}
