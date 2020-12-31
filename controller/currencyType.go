package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/model"
	"xdf/structs"
)

type CurrencyType struct {
}

func (CurrencyType) GetCurrencyType(ctx iris.Context) {
	c := &model.CurrencyType{}
	cSlice := c.GetCurrencyTypeNotMainNet()
	_, _ = ctx.JSON(structs.NewResult(cSlice))
}

func (CurrencyType) GetCurrencyTypeList(ctx iris.Context) {
	c := &model.CurrencyType{}
	cSlice := c.GetAllCurrencyType()
	_, _ = ctx.JSON(structs.NewResult(cSlice))
}

func (CurrencyType) GetCurrencyTypeInfo(ctx iris.Context) {
	currencyTypeID := cast.ToUint(ctx.URLParam("currency_type_id"))
	if currencyTypeID < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("currency_type_id错误")))
		return
	}
	c := &model.CurrencyType{}
	cInfo := c.GetCurrencyTypeByID(currencyTypeID)
	_, _ = ctx.JSON(structs.NewResult(cInfo))

}
