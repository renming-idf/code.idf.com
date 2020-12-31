package datasource

import (
	"errors"
	"fmt"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"os"
	"strings"
	"xdf/transformer"
)

var Db *gorm.DB
var err error
var dirverName string
var conn string

var c *gormadapter.Adapter
var Enforcer *casbin.Enforcer
var casbinConn string

func Register(rc *transformer.Conf) {
	if rc.App.DirverType == "Mysql" {
		dirverName = rc.Mysql.DirverName
		if isTestEnv() {
			casbinConn = rc.Mysql.Connect + rc.Mysql.CasbinName
			conn = rc.Mysql.Connect + rc.Mysql.TName + "?charset=utf8&parseTime=True&loc=Local"
		} else {
			casbinConn = rc.Mysql.Connect + rc.Mysql.CasbinName
			conn = rc.Mysql.Connect + rc.Mysql.Name + "?charset=utf8&parseTime=True&loc=Local"
		}
	}

	Db, err = gorm.Open(dirverName, conn)
	if err != nil {
		panic(err)
	}
	Db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	Db.DB().SetMaxIdleConns(20)   //最大打开的连接数
	Db.DB().SetMaxOpenConns(2000) //设置最大闲置个数
	Db.SingularTable(true)        //表生成结尾不带s
	Db.LogMode(true)

	// casbin
	c, err = gormadapter.NewAdapter(dirverName, casbinConn, true) // Your driver and data source.
	if err != nil {
		color.Red(fmt.Sprintf("NewAdapter 错误: %v", err))
	}

	Enforcer, err = casbin.NewEnforcer("./conf/rbac_model.conf", c)
	if err != nil {
		color.Red(fmt.Sprintf("NewEnforcer 错误: %v", err))
	}
	_ = Enforcer.LoadPolicy()

}

// 获取程序运行环境
// 根据程序运行路径后缀判断
// 如果是 test 就是测试环境
func isTestEnv() bool {
	files := os.Args
	for _, v := range files {
		if strings.Contains(v, "test") {
			return true
		}
	}
	return false
}

// record not found 特殊处理
func IsNotFound(err error) {
	if ok := errors.Is(err, gorm.ErrRecordNotFound); !ok && err != nil {
		color.Red(fmt.Sprintf("error :%v \n ", err))
	}
}
