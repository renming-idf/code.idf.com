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

type Order struct {
}

func (Order) CreateOrder(ctx iris.Context) {
	aul := new(validates.CreateOrderRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	if aul.Type != 1 && aul.Type != 2 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("交易类型选择错误")))
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
	o := &model.Order{}
	err = o.CreateBuyOrder(aul, uid)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("发起交易成功"))
}

func (Order) ServicePopups(ctx iris.Context) {
	amount := cast.ToInt64(ctx.URLParam("amount"))
	if amount <= 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("金额错误")))
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
	o := &model.Order{}
	returnBool, limitAmount, err := o.ServicePopups(uid, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	tmpMap := make(map[string]interface{})
	tmpMap["popups"] = returnBool
	tmpMap["limit"] = limitAmount
	_, _ = ctx.JSON(structs.NewResult(tmpMap))
}

func (Order) CancelOrder(ctx iris.Context) {
	oid := cast.ToUint(ctx.URLParam("order_id"))
	if oid == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("订单id错误")))
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
	o := &model.Order{}
	err := o.CancelOrder(oid, uid)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("取消订单成功"))
}

func (Order) GetHangUpOrder(ctx iris.Context) {
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
	tp := cast.ToInt(ctx.URLParam("type"))
	selfOrder := cast.ToBool(ctx.URLParam("self_order"))
	o := &model.Order{}
	returnInfo, total := o.GetHangUpOrder(uid, page, size, tp, selfOrder)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}

func (Order) Buy(ctx iris.Context) {
	aul := new(validates.BuyAndSellRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
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
	o := &model.Order{}
	err = o.Buy(uid, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("交易成功"))
}

func (Order) GetAllSelfOrder(ctx iris.Context) {
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
	tp := cast.ToInt(ctx.URLParam("type"))
	o := &model.Order{}
	returnInfo, total := o.GetAllOrder(page, size, tp, uid)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}
