package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
)

type Borrow struct {
}

func (Borrow) GetBorrowCurrencyTypeID(ctx iris.Context) {
	b := &model.Borrow{}
	returnInfo := b.GetBorrowCurrencyTypeID()
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

func (Borrow) GetBorrowList(ctx iris.Context) {
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
	currencyTypeID := cast.ToUint(ctx.URLParam("currency_type_id"))
	if currencyTypeID < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("currency_type_id参数错误")))
		return
	}
	b := &model.Borrow{}
	returnInfo := b.GetBorrowList(page, size, uid, currencyTypeID)
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

func (Borrow) GetCurrencyTypeBorrowInfo(ctx iris.Context) {
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
	currencyTypeID := cast.ToUint(ctx.URLParam("currency_type_id"))
	if currencyTypeID < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("currency_type_id参数错误")))
		return
	}
	b := &model.Borrow{}
	returnInfo := b.GetCurrencyTypeBorrowInfo(uid, currencyTypeID)
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

func (Borrow) GetBorrowById(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	if id < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
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
	b := &model.Borrow{}
	returnInfo, err := b.GetBorrowById(id, uid)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

func (Borrow) GetXdfAmount(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	if id < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
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
	amount := cast.ToInt64(ctx.URLParam("amount"))
	if amount < 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("amount不能小于0")))
	}
	currencyTypeID := cast.ToUint(ctx.URLParam("currency_type_id"))
	if currencyTypeID < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("currency_type_id参数错误")))
		return
	}
	b := &model.Borrow{}
	xdfAmount, err := b.GetXdfAmount(id, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	pledgeAmount, err := b.GetPledgeAmount(id, currencyTypeID, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	tmpMap := make(map[string]int64)
	tmpMap["pledge_amount"] = pledgeAmount
	tmpMap["xdf_amount"] = xdfAmount
	uw := &model.UserWallet{}
	uw, _ = uw.GetUserWalletByCurrencyTypeID(uid, currencyTypeID)
	tmpMap["balance"] = uw.Balance
	_, _ = ctx.JSON(structs.NewResult(tmpMap))
}

func (Borrow) Borrow(ctx iris.Context) {
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
	id, _ := ctx.Params().GetUint("id")
	if id < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
	}
	amount := cast.ToInt64(ctx.URLParam("amount"))
	if amount < 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("借款数量不能小于0")))
	}
	currencyTypeID := cast.ToUint(ctx.URLParam("currency_type_id"))
	if currencyTypeID < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("currency_type_id参数错误")))
		return
	}
	b := &model.Borrow{}
	err := b.Borrow(uid, id, currencyTypeID, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("借款成功"))
}

func (Borrow) GetUserBorrowList(ctx iris.Context) {
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
	state := cast.ToInt(ctx.URLParam("state"))
	if state < 1 || state > 2 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("state参数错误")))
		return
	}
	ub := &model.UserBorrow{}
	returnInfo, total := ub.GetUserBorrowList(page, size, state, uid)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}

func (Borrow) UpdateAutomaticRenew(ctx iris.Context) {
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
	id, _ := ctx.Params().GetUint("id")
	if id < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
	}
	automaticRenew := cast.ToBool(ctx.URLParam("automatic_renew"))
	ub := &model.UserBorrow{}
	err := ub.UpdateAutomaticRenew(id, uid, automaticRenew)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("无法修改续借状态")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("修改成功"))
}

func (Borrow) Repayment(ctx iris.Context) {
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
	id, _ := ctx.Params().GetUint("id")
	if id < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
	}
	ub := &model.UserBorrow{}
	err := ub.Repayment(uid, id)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("还款成功"))
}

func (Borrow) GetExpireUserBorrow(ctx iris.Context) {
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
	ub := &model.UserBorrow{}
	ubSlice := ub.GetExpireUserBorrow(uid)
	if len(ubSlice) > 0 {
		_, _ = ctx.JSON(structs.NewResult("您有即将到期的借款，请及时处理"))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(errors.New("不需要处理")))
}
