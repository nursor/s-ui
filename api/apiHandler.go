package api

import (
	"strings"

	"github.com/alireza0/s-ui/service"
	"github.com/alireza0/s-ui/util/common"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	ApiService
	apiv2 *APIv2Handler
}

func NewAPIHandler(g *gin.RouterGroup, a2 *APIv2Handler) {
	a := &APIHandler{
		apiv2: a2,
		ApiService: ApiService{
			FrpService: *service.NewFrpService(),
		},
	}
	a.initRouter(g)
}

func (a *APIHandler) initRouter(g *gin.RouterGroup) {
	g.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "login") && !strings.HasSuffix(path, "logout") {
			checkLogin(c)
		}
	})

	// FRP专用路由组（必须在通用路由之前）
	frp := g.Group("/frp")
	{
		// GET 路由
		frp.GET("/servers", a.ApiService.GetFrpServers)
		frp.GET("/server", a.ApiService.GetFrpServer)
		frp.GET("/status", a.ApiService.GetFrpServerStatus)
		frp.GET("/logs", a.ApiService.GetFrpLogs)

		// POST 路由
		frp.POST("/save", func(c *gin.Context) {
			a.ApiService.SaveFrpServer(c, GetLoginUser(c))
		})
		frp.POST("/start", func(c *gin.Context) {
			a.ApiService.StartFrpServer(c, GetLoginUser(c))
		})
		frp.POST("/stop", func(c *gin.Context) {
			a.ApiService.StopFrpServer(c, GetLoginUser(c))
		})
		frp.POST("/restart", func(c *gin.Context) {
			a.ApiService.RestartFrpServer(c, GetLoginUser(c))
		})
	}

	// 通用路由
	g.POST("/:postAction", a.postHandler)
	g.GET("/:getAction", a.getHandler)
}

func (a *APIHandler) postHandler(c *gin.Context) {
	loginUser := GetLoginUser(c)
	action := c.Param("postAction")

	switch action {
	case "login":
		a.ApiService.Login(c)
	case "changePass":
		a.ApiService.ChangePass(c)
	case "save":
		a.ApiService.Save(c, loginUser)
	case "restartApp":
		a.ApiService.RestartApp(c)
	case "restartSb":
		a.ApiService.RestartSb(c)
	case "linkConvert":
		a.ApiService.LinkConvert(c)
	case "importdb":
		a.ApiService.ImportDb(c)
	case "addToken":
		a.ApiService.AddToken(c)
		a.apiv2.ReloadTokens()
	case "deleteToken":
		a.ApiService.DeleteToken(c)
		a.apiv2.ReloadTokens()
	// FRP相关（下划线格式，保持向后兼容）
	case "save_frp_server":
		a.ApiService.SaveFrpServer(c, loginUser)
	case "start_frp":
		a.ApiService.StartFrpServer(c, loginUser)
	case "stop_frp":
		a.ApiService.StopFrpServer(c, loginUser)
	case "restart_frp":
		a.ApiService.RestartFrpServer(c, loginUser)
	case "save_frp_proxy":
		a.ApiService.SaveFrpProxy(c, loginUser)
	case "download_frp":
		a.ApiService.DownloadFrp(c, loginUser)
	default:
		jsonMsg(c, "failed", common.NewError("unknown action: ", action))
	}
}

func (a *APIHandler) getHandler(c *gin.Context) {
	action := c.Param("getAction")

	switch action {
	case "logout":
		a.ApiService.Logout(c)
	case "load":
		a.ApiService.LoadData(c)
	case "inbounds", "outbounds", "endpoints", "services", "tls", "clients", "config":
		err := a.ApiService.LoadPartialData(c, []string{action})
		if err != nil {
			jsonMsg(c, action, err)
		}
		return
	case "users":
		a.ApiService.GetUsers(c)
	case "settings":
		a.ApiService.GetSettings(c)
	case "stats":
		a.ApiService.GetStats(c)
	case "status":
		a.ApiService.GetStatus(c)
	case "onlines":
		a.ApiService.GetOnlines(c)
	case "logs":
		a.ApiService.GetLogs(c)
	case "changes":
		a.ApiService.CheckChanges(c)
	case "keypairs":
		a.ApiService.GetKeypairs(c)
	case "getdb":
		a.ApiService.GetDb(c)
	case "tokens":
		a.ApiService.GetTokens(c)
	// FRP相关（下划线格式，保持向后兼容）
	case "frp_servers":
		a.ApiService.GetFrpServers(c)
	case "frp_server":
		a.ApiService.GetFrpServer(c)
	case "frp_server_status":
		a.ApiService.GetFrpServerStatus(c)
	case "frp_status":
		a.ApiService.GetFrpServerStatus(c)
	case "frp_logs":
		a.ApiService.GetFrpLogs(c)
	default:
		jsonMsg(c, "failed", common.NewError("unknown action: ", action))
	}
}
