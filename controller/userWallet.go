package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"time"
	"xdf/common/log"
	"xdf/common/watch"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
	"xdf/validates"
)

type UserWallet struct {
}

// 获取用户资产
func (UserWallet) GetUserWallet(ctx iris.Context) {
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
	uw := &model.UserWallet{}
	cny, returnInfo := uw.GetUserWallet(uid)
	tmpMap := make(map[string]interface{})
	tmpMap["wallet"] = returnInfo
	tmpMap["total"] = cny
	_, _ = ctx.JSON(structs.NewResult(tmpMap))
}

// 获取用户该币种下可用余额
func (UserWallet) GetUserWalletByCurrencyTypeID(ctx iris.Context) {
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
		_, _ = ctx.JSON(structs.NewResult(errors.New("currency_type_id错误")))
		return
	}
	uw := &model.UserWallet{}
	uw, err := uw.GetUserWalletByCurrencyTypeID(uid, currencyTypeID)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult("该用户钱包存在问题，请联系管理员！"))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(uw))
}

func (UserWallet) TransferAccount(ctx iris.Context) {
	aul := new(validates.TransferAccountPasswordRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	if aul.Amount < 200000 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("数量不能小于20")))
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
	u := &model.User{}
	err = u.CheckUserPassword(uid, aul.PassWord)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("安全密码错误")))
		return
	}
	uw := &model.UserWallet{}
	err = uw.TransferAccount(uid, aul.CurrencyTypeID, aul.PublicKey, aul.Amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("转账成功"))
}

// 获取激活矿工需要的费用
func (UserWallet) GetActivatePrice(ctx iris.Context) {
	iam := model.IniAllowMining{}
	cid := iam.GetAllowMingCtID()
	if cid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("没有获取到激活费用，请联系管理员")))
		return
	}
	s := &model.IniServiceCharge{}
	s, err := s.GetServiceCharge(cid, "activate")
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("没有获取到激活费用，请联系管理员")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(s))

}

//提币 必须是主网络上的 且 需要从主钱包传入
func (UserWallet) WithdrawMoney(ctx iris.Context) {
	nowTime := time.Now()
	muteEndTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 10, 0, 0, 0, time.Local)
	muteStartTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 21, 0, 0, 0, time.Local)
	if nowTime.Before(muteEndTime) || nowTime.After(muteStartTime) {
		_, _ = ctx.JSON(structs.NewResult(errors.New("提现时间为：10:00—21:00")))
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
	aul := new(validates.TransferAccountRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	var iwl model.IniWithdrawLimit
	count := iwl.GetLimitByCurrencyTypeID(aul.CurrencyTypeID)
	if aul.Amount < count {
		_, _ = ctx.JSON(structs.NewResult(errors.New("数量不能小于" + cast.ToString(count))))
		return
	}
	if !watch.IsETHAddress(aul.PublicKey) {
		_, _ = ctx.JSON(structs.NewResult(errors.New("请输入正确的地址")))
		return
	}
	// 获取手续费信息
	s := &model.IniServiceCharge{}
	s, err = s.GetServiceCharge(aul.CurrencyTypeID, "withdraw")
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("获取手续费失败")))
		return
	}
	ct := model.CurrencyType{}
	tokenAddressHex := ct.GetCurrencyTypeByID(aul.CurrencyTypeID).ContractAddress
	if tokenAddressHex == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("币种暂不能提币！")))
		return
	}
	u := &model.User{}
	// 记录账单
	uw := model.UserWallet{}
	ok = watch.IsInAccountAddressPool(aul.PublicKey)
	from := u.GetUserById(uid).AccountAddress
	to := aul.PublicKey
	_, err = uw.WithdrawMoney(uid, aul.CurrencyTypeID, float64(aul.Amount), s.Amount, ok, from, to)
	if err != nil {
		log.Error(err)
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("已提交提币请求"))
}

// 获取提币、充币记录
func (UserWallet) GetCapitalTrendsRecord(ctx iris.Context) {
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
		_, _ = ctx.JSON(structs.NewResult(errors.New("currencyTypeID错误")))
		return
	}
	tp := cast.ToInt(ctx.URLParam("type"))
	if tp != 1 && tp != 2 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("type错误")))
		return
	}
	c := &model.CapitalTrendsRecord{}
	returnInfo, total := c.GetCapitalTrendsRecord(uid, currencyTypeID, tp, page, size)
	p := structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}
	ctx.JSON(structs.NewResult(p))
}
