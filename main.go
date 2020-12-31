package main

import (
	"fmt"
	"github.com/Rhymond/go-money"
	"xdf/files"
	"xdf/routers"
	"xdf/validates"

	"github.com/fatih/color"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"

	"io"
	"log"
	"os"
	"strings"
	"time"
)

var Sc iris.Configuration

// 获取路由信息
func getRoutes(api *iris.Application) []*validates.PermissionRequest {
	rs := api.APIBuilder.GetRoutes()
	var rrs []*validates.PermissionRequest
	for _, s := range rs {
		if strings.Contains(s.Path, "v1") {
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

// 创建打开文件
func newLogFile() *os.File {

	uploadsPath := "./uploads/"
	_ = files.CreateFile(uploadsPath)

	path := "./logs/"
	_ = files.CreateFile(path)
	loc, _ := time.LoadLocation("Asia/Shanghai") //重要：获取时区
	filename := path + time.Now().In(loc).Format("2006-01-02") + ".log"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		color.Red(fmt.Sprintf("日志记录出错: %v", err))
	}
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw) // 将文件设置为log输出的文件
	log.SetPrefix("[NewFour]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	return f
}

func main() {
	c := money.GetCurrency("CNY")
	c.Grapheme = ""
	c.Template = "1"
	c.Thousand = ""

	//f := newLogFile()
	//defer f.Close()
	s := routers.New()
	//api.Logger().SetOutput(io.MultiWriter(f, os.Stdout)) //记录日志
	err := s.Run()
	if err != nil {
		color.Yellow(fmt.Sprintf("项目运行结束: %v", err))
	}
}
