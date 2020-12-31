package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
)

type EcologicalPool struct {
}

//获取所有矿机
func (EcologicalPool) GetEcologicalPool(ctx iris.Context) {
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
	ep := &model.EcologicalPool{}
	returnInfo, total := ep.GetEcologicalPool(page, size)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}

func (EcologicalPool) GetEcologicalPoolProfit(ctx iris.Context) {
	epID := cast.ToUint(ctx.URLParam("id"))
	if epID == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	amount := cast.ToInt64(ctx.URLParam("amount"))
	if amount == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("金额禁止为0")))
		return
	}
	ep := &model.EcologicalPool{}
	returnInfo, err := ep.GetEcologicalPoolProfit(epID, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(returnInfo))

}

func (EcologicalPool) OpenEcological(ctx iris.Context) {
	epID := cast.ToUint(ctx.URLParam("id"))
	if epID == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	amount := cast.ToInt64(ctx.URLParam("amount"))
	if amount == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("金额禁止为0")))
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
	ep := &model.UserEcologicalPool{}
	err := ep.OpenEcologicalPool(epID, uid, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("投入成功"))
}

func (EcologicalPool) GetLadder(ctx iris.Context) {
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
	ep := &model.UserEcologicalPool{}
	returnInfo := ep.GetLadder(uid, page, size)
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

func (EcologicalPool) GetLadderPageInfo(ctx iris.Context) {
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
	ep := &model.UserEcologicalPool{}
	returnInfo, total := ep.GetLadderPageInfo(uid, page, size)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}
