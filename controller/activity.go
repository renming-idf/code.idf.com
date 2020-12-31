package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
)

type Activity struct {
}

func (Activity) GetActivity(ctx iris.Context) {
	id := cast.ToUint(ctx.URLParam("id"))
	if id == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("ID错误")))
		return
	}
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 || size > 10 {
		size = 10
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
	a := &model.Activity{}
	returnInfo, err := a.GetActivity(id, uid, page, size)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

func (Activity) GetActivityPageInfo(ctx iris.Context) {
	id := cast.ToUint(ctx.URLParam("id"))
	if id == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("ID错误")))
		return
	}
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 || size > 10 {
		size = 10
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
	a := &model.Activity{}
	returnInfo, total := a.GetActivityPageInfo(id, uid, page, size)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}

func (Activity) Buy(ctx iris.Context) {
	id := cast.ToUint(ctx.URLParam("id"))
	if id == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("ID错误")))
		return
	}
	amount := cast.ToInt64(ctx.URLParam("amount"))
	if amount <= 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("amount错误")))
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
	a := &model.Activity{}
	err := a.Buy(id, uid, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("认购成功"))
}
