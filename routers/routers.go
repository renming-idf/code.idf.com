package routers

import (
	"errors"
	"github.com/iris-contrib/middleware/cors"
	json "github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/pprof"
	"github.com/kataras/iris/v12/websocket"
	gf "github.com/snowlyg/gotransformer"
	"strings"
	"time"
	"xdf/common"
	"xdf/common/log"
	"xdf/controller"
	"xdf/controller/casbinControllers"
	"xdf/datasource"
	"xdf/middleware"
	"xdf/middleware/sockets/gorilla"
	"xdf/model"
	"xdf/transformer"
	"xdf/validates"
)

// 获取路由信息
func getRoutes(api *iris.Application) []*validates.PermissionRequest {
	rs := api.APIBuilder.GetRoutes()
	var rrs []*validates.PermissionRequest
	for _, s := range rs {
		if strings.Contains(s.Path, "v2") {
			if !isPermRoute(s) {
				path := strings.Replace(s.Path, ":id", "*", 1)
				rr := &validates.PermissionRequest{Name: path, DisplayName: s.Name, Description: s.Name, Act: s.Method}
				rrs = append(rrs, rr)
			}
		}
	}
	return rrs
}

// 过滤非必要权限
func isPermRoute(s *router.Route) bool {
	exceptRouteName := []string{"OPTIONS", "GET", "POST", "HEAD", "PUT", "PATCH"}
	for _, er := range exceptRouteName {
		if strings.Contains(s.Name, er) {
			return true
		}
	}
	return false
}

// 获取配置信息
func getSysConf(Sc iris.Configuration) *transformer.Conf {
	app := transformer.App{}
	g := gf.NewTransform(&app, Sc.Other["App"], time.RFC3339)
	_ = g.Transformer()
	db := transformer.Mysql{}
	g.OutputObj = &db
	g.InsertObj = Sc.Other["Mysql"]
	_ = g.Transformer()

	testData := transformer.TestData{}
	g.OutputObj = &testData
	g.InsertObj = Sc.Other["TestData"]
	_ = g.Transformer()

	data := transformer.Data{}
	g.OutputObj = &data
	g.InsertObj = Sc.Other["Data"]
	_ = g.Transformer()
	cf := &transformer.Conf{
		App:      app,
		Mysql:    db,
		TestData: testData,
		Data:     data,
	}
	return cf
}

/**
*初始化系统 账号 权限 角色
 */
func CreateSystemData(rc *transformer.Conf, perms []*validates.PermissionRequest) {
	if rc.App.CreateSysData {
		p := &model.Permission{}
		permIds := p.CreateSystemAdminPermission(perms) //初始化权限
		r := &model.Role{}
		role := r.CreateSystemAdminRole(permIds) //初始化角色
		if role.ID != 0 {
			au := &model.AdminUser{}
			au.CreateSystemAdmin(role.ID, rc) //初始化管理员
		}
	}
}

type Server struct {
	//db     *gorm.DB
	config *transformer.Conf
	api    *iris.Application
	sc     iris.Configuration
}

func New() Server {

	server := Server{}
	server.sc = iris.TOML("./conf/conf.tml") // 加载配置文件
	requestLogger := logger.New(logger.Config{
		// Status displays status code
		Status: true,
		// IP displays request's remote address
		IP: true,
		// Method displays the http method
		Method: true,
		// Path displays the request path
		Path: true,
		// Query appends the url query to the Path.
		Query: true,
		// if !empty then its contents derives from `ctx.Values().Get("logger_message")
		// will be added to the logs.
		MessageContextKeys: []string{"logger_message"},
		// if !empty then its contents derives from `ctx.GetHeader("User-Agent")
		MessageHeaderKeys: []string{"User-Agent"},
	})
	server.config = getSysConf(server.sc) //格式化配置文件 other 数据
	api := iris.New()
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"PUT", "PATCH", "GET", "POST", "OPTIONS", "DELETE"},
	})
	api.Use(crs)
	api.AllowMethods(iris.MethodOptions)
	api.Logger().SetLevel(server.config.App.LoggerLevel)

	datasource.Register(server.config) // 数据库初始化
	model.CreateTable()                // 自动创建数据库
	iris.RegisterOnInterrupt(func() {
		_ = datasource.Db.Close()
	})

	server.Register(api)     //注册路由
	middleware.Register(api) // 中间件注册
	//apiRoutes := getRoutes(api)                // 获取路由数据
	//CreateSystemData(server.config, apiRoutes) // 初始化系统数据 管理员账号，角色，权限
	api.Use(requestLogger)
	api.HandleDir("/uploads", iris.Dir(server.config.Data.Uploads))
	server.api = api
	server.initFunc(server.config)
	if err := common.DirExists(server.config.Data.Uploads); err != nil {
		panic(err)
	}
	return server
}

func (s Server) initFunc(t *transformer.Conf) {
	//初始化验证码创建者
	common.InvitationCodeCreator = common.Code{
		Base:    "123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		Decimal: 35,
		Pad:     "0",
		Len:     6,
	}
	invitationCode := common.InvitationCodeCreator.IdToCode(uint64(1))
	log.Println(invitationCode)
	//watch.InitPool(t)
	middleware.InitSession()
	go controller.TimeTaskInit()
}

//func (s Server) injectData(ctx iris.Context) {
//	values := ctx.Values()
//	values.Set("conf", s.config)
//	ctx.Next()
//}

func (s Server) Run() error {
	if err := s.api.Run(iris.TLS(s.config.App.URl+s.config.App.Port, s.config.App.CertFile, s.config.App.KeyFile), iris.WithoutServerError(iris.ErrServerClosed), iris.WithOptimizations); err != nil {
		log.Error(err)
		return s.api.Run(iris.Addr(s.config.App.Port), iris.WithConfiguration(s.sc), iris.WithoutServerError(iris.ErrServerClosed), iris.WithOptimizations)
	}
	return nil
}

func (s Server) Register(api *iris.Application) {
	websocketServer := websocket.New(
		gorilla.DefaultUpgrader, websocket.Events{
			websocket.OnNativeMessage: func(nsConn *websocket.NSConn, msg websocket.Message) error {
				rmt := &model.ChatRecord{}
				if string(msg.Body) == "ping" {
					return nil
				}
				err := json.Unmarshal(msg.Body, rmt)
				if err != nil {
					log.Error(err)
					return errors.New("消息内容错误")
				}
				// 存储到数据库
				err = rmt.SaveChatRecord(rmt)
				if err != nil {
					log.Error(err)
					return errors.New("存入数据库失败")
				}
				//  用户发送的
				if rmt.Type == 2 {
					err := gorilla.GetOneAdminClient(rmt)
					if err != nil {
						log.Error(err)
						return err
					}
				} else if rmt.Type == 1 {
					con, ok := gorilla.Clients.Load(rmt.To)
					if ok {
						c, ok := con.(*gorilla.Socket)
						if ok {
							err = c.UnderlyingConn.WriteMessage(gorilla.TextMessage, msg.Body)
							if err != nil {
								gorilla.Clients.Delete(rmt.To)
								return err
							}
						}
					}
				}
				return nil
			},
		})
	main := api.Party("/")
	api.Get("/get_account_address", controller.User{}.GetAllUserAccountAddress)
	api.Get("/get_gas", controller.User{}.GetUsdtGas)
	//pprof
	// 记载主路由
	api.Any("/debug/pprof", pprof.New())
	// 加载子路由
	api.Any("/debug/pprof/{action:path}", pprof.New())
	//main.Use(s.injectData)
	{
		//APP
		v1 := main.Party("/v1")
		{
			v1.PartyFunc("/", func(noToken iris.Party) {
				noToken.Use(middleware.FilterIpRequest)
				noToken.Get("/app/version/share", controller.Version{}.Share).Name = "获取下载地址"
				noToken.Get("/app/version/download_address", controller.Version{}.GetDownloadAddress).Name = "获取下载地址"
				noToken.Get("/official/download_address", controller.Version{}.GetOfficialAddress).Name = "获取官网下载地址"
				noToken.Get("/echo", websocket.Handler(websocketServer))
				noToken.Post("/app/user/create_user", controller.User{}.CreateUser).Name = "创建地址"
				noToken.Post("/app/user/import", controller.User{}.Import).Name = "导入账户"
				noToken.Get("/app/version/info", controller.Version{}.GetVersion).Name = "获取当前版本号"
			})
			v1.PartyFunc("/app", func(app iris.Party) {
				app.Use(middleware.ReqInterception, middleware.CheckSession, middleware.CheckUserForbidden)
				app.PartyFunc("/user", func(users iris.Party) {
					users.Get("/update_token", controller.User{}.FlashToken).Name = "刷新token"
					users.Post("/put_file", controller.User{}.PutFile).Name = "聊天上传文件"
					users.Post("/user_details", controller.User{}.GetUserDetails).Name = "账号管理中用户详情信息"
					users.Get("/preach_info", controller.User{}.GetPreachInfo).Name = "获取布道中一级矿工的用户信息"
					users.Get("/public_key", controller.User{}.GetUserPublicKey).Name = "获取用户钱包地址"
					users.Get("/account_address", controller.User{}.GetUserAccountAddress).Name = "获取用户USDT充值地址"
					users.Get("/home", controller.User{}.GetHome).Name = "获取首页"
					users.Post("/change_password", controller.User{}.ChangePassword).Name = "修改密码"
					users.Get("/check_password", controller.User{}.CheckPassword).Name = "验证密码"
					users.Get("/check_login_password", controller.User{}.CheckLoginPasswordByID).Name = "验证登录密码"
					users.Get("/change_name", controller.User{}.ChangeUserName).Name = "修改名称"
					users.Get("/online_time", controller.User{}.GetOnlineTime).Name = "获取在线时间"
				})

			})
		}
	}
}
