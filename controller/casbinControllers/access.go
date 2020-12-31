package casbinControllers

import (
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"time"
	"xdf/common/log"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
	"xdf/validates"
)

type Access struct {
}

//用于提出当前登录的用户
func BackManyKickPeople(userID ...int) {
	for _, v := range userID {
		token, ok := middleware.BackendTokenMap.Load(uint(v))
		if ok {
			var sessionID = ""
			if token != "" {
				m, _ := middleware.ParseToken(cast.ToString(token), middleware.JwtKey)
				sessionID = m["session"]
				middleware.SMgr.EndSessionBy(sessionID, 2)
			}
		}
	}
}

// 用户登陆
func (Access) UserLogin(ctx iris.Context) {
	aul := new(validates.LoginRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	ctx.Application().Logger().Infof("%s 登录系统", aul.Username)
	au := &model.AdminUser{}
	admin, err := au.CheckLogin(aul.Username, aul.Password)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	j := &model.Journal{}
	ip := ctx.RemoteAddr()
	err = j.CreateJournal(admin.ID, ip)
	if err != nil {
		log.Println("记录登录状态失败")
	}
	//先删除之前用户的session
	BackManyKickPeople(int(admin.ID))
	var sessionID = middleware.SMgr.StartSession(ctx.ResponseWriter(), 2)
	lt := time.Now().Add(24 * 7 * time.Hour).Unix()
	tokenMap := map[string]interface{}{"aid": admin.ID, "exp": lt, "session": sessionID, "type": 2}
	tokenString := middleware.CreateToken(tokenMap, 2)
	middleware.SMgr.SetSessionVal(sessionID, "UserInfo", admin, 2)
	s := &model.SessionInfo{}
	err = s.SaveSessionInfo(admin.ID, sessionID, tokenString, 2, lt)
	if err != nil {
		log.Println("记录登录状态失败")
	}
	m := make(map[string]interface{})
	m["token"] = tokenString
	m["userInfo"] = userTransform(admin)
	_, _ = ctx.JSON(structs.NewResult(m))
}
