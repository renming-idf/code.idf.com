package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/model"
	"xdf/structs"
)

type ServiceCharge struct {
}

func (ServiceCharge) GetServiceCharge(ctx iris.Context) {
	currencyTypeID := cast.ToUint(ctx.URLParam("currency_type_id"))
	if currencyTypeID < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("currencyTypeID错误")))
		return
	}
	name := ctx.URLParam("name")
	if name == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("请输入正确的参数！")))
		return
	}
	s := &model.IniServiceCharge{}
	s, err := s.GetServiceCharge(currencyTypeID, name)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("获取手续费失败")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(s))
}
