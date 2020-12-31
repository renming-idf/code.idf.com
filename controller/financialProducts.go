package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
)

type FinancialProducts struct {
}

// 刚刚进理财页面获取列表、资产以及收益
func (FinancialProducts) GetFinancialProductsInfo(ctx iris.Context) {
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
	returnMap := make(map[string]interface{})
	fp := &model.FinancialProducts{}
	returnMap["fp_slice"] = fp.GetAllFinancialProductsInfo()
	returnMap["profit"] = fp.GetFinancialProductsInfo(uid)
	_, _ = ctx.JSON(structs.NewResult(returnMap))
}

//// 获取理财页面的列表
//func (FinancialProducts) GetFinancialProductsList(ctx iris.Context) {
//	page := cast.ToInt(ctx.URLParam("page"))
//	if page < 1 {
//		page = 1
//	}
//	size := cast.ToInt(ctx.URLParam("size"))
//	if size < 1 || size > 10 {
//		size = 10
//	}
//	fp := &model.FinancialProducts{}
//	returnInfo, total := fp.GetAllFinancialProducts(page, size)
//	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
//}

// 转入
func (FinancialProducts) TransferIn(ctx iris.Context) {
	m, ok := middleware.ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(401)
		return
	}
	uid := cast.ToUint(m["aid"])
	if uid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("uid错误")))
		return
	}
	id := cast.ToUint(ctx.URLParam("id"))
	if id < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	amount := cast.ToInt64(ctx.URLParam("amount"))
	if amount <= 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("转入数量不能小于0")))
		return
	}
	uf := &model.UserFinancialProducts{}
	err := uf.TransferIn(uid, id, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("转入成功"))
}

// 转出
func (FinancialProducts) TransferOut(ctx iris.Context) {
	m, ok := middleware.ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(401)
		return
	}
	uid := cast.ToUint(m["aid"])
	if uid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("uid错误")))
		return
	}
	id := cast.ToUint(ctx.URLParam("id"))
	if id < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	amount := cast.ToInt64(ctx.URLParam("amount"))
	if amount <= 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("转出金额不能小于0")))
		return
	}
	uf := &model.UserFinancialProducts{}
	err := uf.TransferOut(uid, id, amount)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("转出成功"))
}

// 查看用户的理财产品
func (FinancialProducts) GetUserFinancialProducts(ctx iris.Context) {
	m, ok := middleware.ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(401)
		return
	}
	uid := cast.ToUint(m["aid"])
	if uid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("uid错误")))
		return
	}
	id := cast.ToUint(ctx.URLParam("id"))
	if id < 1 {
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
	uf := &model.UserFinancialProducts{}
	returnInfo, err := uf.GetUserFinancialProducts(uid, id, page, size)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

// 查看用户的理财产品下的收益详情
func (FinancialProducts) GetUserFinancialProductsList(ctx iris.Context) {
	m, ok := middleware.ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(401)
		return
	}
	uid := cast.ToUint(m["aid"])
	if uid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("uid错误")))
		return
	}
	id := cast.ToUint(ctx.URLParam("id"))
	if id < 1 {
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
	uf := &model.UserFinancialProducts{}
	returnInfo, total := uf.GetUserFinancialProductsList(uid, id, page, size)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}

func (FinancialProducts) GetFinancialIcon(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	if id < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	f := model.FinancialProducts{}
	_, _ = ctx.JSON(structs.NewResult(f.GetIcon(id)))
}
