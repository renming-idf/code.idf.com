package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
)

type AirPods struct {
}

func (AirPods) GetAirDrop(ctx iris.Context) {
	code := ctx.URLParam("code")
	if code == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("兑换码禁止为空")))
		return
	}
	m, ok := middleware.ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(401)
		return
	}
	uid := cast.ToUint(m["aid"])
	if uid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	ad := &model.UserAirDrop{}
	err := ad.GetAirDrop(uid, code)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("兑换成功"))
}

func (AirPods) Popups(ctx iris.Context) {
	m, ok := middleware.ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(401)
		return
	}
	uid := cast.ToUint(m["aid"])
	if uid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	ad := &model.UserAirDrop{}
	p, amount := ad.Popups(uid)
	tmpMap := make(map[string]interface{})
	tmpMap["popups"] = p
	tmpMap["amount"] = amount
	_, _ = ctx.JSON(structs.NewResult(tmpMap))
}

func (AirPods) ReceiveAirDrop(ctx iris.Context) {
	m, ok := middleware.ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(401)
		return
	}
	uid := cast.ToUint(m["aid"])
	if uid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	ad := &model.UserAirDrop{}
	err := ad.ReceiveAirDrop(uid)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("领取成功"))
}
