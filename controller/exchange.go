package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
	"xdf/validates"
)

type Exchange struct {
}

func (Exchange) GetExchangeInfo(ctx iris.Context) {
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
	ei := &model.ExchangeInfo{}
	info, err := ei.GetExchangeInfo(uid)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(info))
}

func (Exchange) GetUsdtAmount(ctx iris.Context) {
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
	amount := cast.ToInt64(ctx.URLParam("amount"))
	ei := &model.ExchangeInfo{}
	_, _ = ctx.JSON(structs.NewResult(ei.GetUsdtAmount(amount, uid)))
}

func (Exchange) GetAllExchangeRecord(ctx iris.Context) {
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
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 || size > 10 {
		size = 10
	}
	er := &model.ExchangeRecord{}
	returnInfo, total := er.GetAllExchangeRecord(page, size, 0, uid)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}

// 兑换
func (Exchange) Exchange(ctx iris.Context) {
	aul := new(validates.ExchangeRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	if aul.Amount < 100000 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("兑换数量不少于10")))
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
	er := &model.ExchangeRecord{}
	err = er.Exchange(aul, uid)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("兑换成功"))
}
