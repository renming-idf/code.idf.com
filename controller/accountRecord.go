package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"time"
	"xdf/common"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
)

type AccountRecord struct {
}

func (AccountRecord) GetAccountRecordList(ctx iris.Context) {
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
	if tp < 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("type错误")))
		return
	}
	currencyTypeID := cast.ToUint(ctx.URLParam("currency_type_id"))
	// 通过时间查询
	bt := cast.ToString(ctx.URLParam("beginTime"))
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		ctx.JSON(structs.NewResult(errors.New("时区获取错误")))
		return
	}
	var beginTime, endTime time.Time
	if bt != "" {
		beginTime, err = time.ParseInLocation(common.MONTH_DATE_FORMAT, bt, loc) //使用模板在对应时区转化为time.time类型
		if err != nil {
			ctx.JSON(structs.NewResult(errors.New("起始日期输入有误")))
			return
		}
		endTime = beginTime.AddDate(0, 1, 0)
	}
	ar := &model.AccountRecord{}
	tmpMap := ar.GetAccountRecordList(uid, page, size, tp, currencyTypeID, beginTime, endTime)
	ctx.JSON(structs.NewResult(tmpMap))
}

func (AccountRecord) GetAccountRecordInfo(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	if id < 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("账单id错误")))
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
	ar := &model.AccountRecord{}
	ar, err := ar.GetAccountRecordInfo(uid, id)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(ar))
}

// 获取挖矿收益详情
func (AccountRecord) GetEcologicalAccountRecord(ctx iris.Context) {
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
	var mingct model.IniAllowMining
	currencyTypeID := mingct.GetAllowMingCtID()
	// 通过时间查询
	bt := cast.ToString(ctx.URLParam("beginTime"))
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		ctx.JSON(structs.NewResult(errors.New("时区获取错误")))
		return
	}
	var beginTime, endTime time.Time
	if bt != "" {
		beginTime, err = time.ParseInLocation(common.SECOND_DATE_FORMAT, bt, loc) //使用模板在对应时区转化为time.time类型
		if err != nil {
			ctx.JSON(structs.NewResult(errors.New("起始日期输入有误")))
			return
		}
	}
	endTime = beginTime.AddDate(0, 1, 0)
	ar := &model.AccountRecord{}
	returnInfo := ar.GetEcologicalAccountRecord(uid, currencyTypeID, page, size, beginTime, endTime)
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

// 获取挖矿收益详情下的列表
func (AccountRecord) GetEcologicalAccountRecordList(ctx iris.Context) {
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
	var mingct model.IniAllowMining
	currencyTypeID := mingct.GetAllowMingCtID()
	// 通过时间查询
	bt := cast.ToString(ctx.URLParam("beginTime"))
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		ctx.JSON(structs.NewResult(errors.New("时区获取错误")))
		return
	}
	var beginTime, endTime time.Time
	if bt != "" {
		beginTime, err = time.ParseInLocation(common.SECOND_DATE_FORMAT, bt, loc) //使用模板在对应时区转化为time.time类型
		if err != nil {
			ctx.JSON(structs.NewResult(errors.New("起始日期输入有误")))
			return
		}
	}
	endTime = beginTime.AddDate(0, 1, 0)
	ar := &model.AccountRecord{}
	returnInfo := ar.GetEcologicalIncomeList(uid, currencyTypeID, page, size, beginTime, endTime)
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

// 获取布道收益
func (AccountRecord) GetPreacherIncome(ctx iris.Context) {
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
	var mingct model.IniAllowMining
	currencyTypeID := mingct.GetAllowMingCtID()
	// 通过时间查询
	bt := cast.ToString(ctx.URLParam("beginTime"))
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		ctx.JSON(structs.NewResult(errors.New("时区获取错误")))
		return
	}
	var beginTime, endTime time.Time
	if bt != "" {
		beginTime, err = time.ParseInLocation(common.SECOND_DATE_FORMAT, bt, loc) //使用模板在对应时区转化为time.time类型
		if err != nil {
			ctx.JSON(structs.NewResult(errors.New("起始日期输入有误")))
			return
		}
	}
	endTime = beginTime.AddDate(0, 1, 0)
	ar := &model.AccountRecord{}
	returnInfo := ar.GetPreacherIncome(uid, currencyTypeID, page, size, beginTime, endTime)
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

// 获取布道收益的列表
func (AccountRecord) GetPreacherIncomeList(ctx iris.Context) {
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
	var mingct model.IniAllowMining
	currencyTypeID := mingct.GetAllowMingCtID()
	// 通过时间查询
	bt := cast.ToString(ctx.URLParam("beginTime"))
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		ctx.JSON(structs.NewResult(errors.New("时区获取错误")))
		return
	}
	var beginTime, endTime time.Time
	if bt != "" {
		beginTime, err = time.ParseInLocation(common.SECOND_DATE_FORMAT, bt, loc) //使用模板在对应时区转化为time.time类型
		if err != nil {
			ctx.JSON(structs.NewResult(errors.New("起始日期输入有误")))
			return
		}
	}
	endTime = beginTime.AddDate(0, 1, 0)
	ar := &model.AccountRecord{}
	returnInfo := ar.GetPreacherIncomeList(uid, currencyTypeID, page, size, beginTime, endTime)
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

// 获取团队收益
func (AccountRecord) GetTeamIncome(ctx iris.Context) {
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
	ar := &model.AccountRecord{}
	returnInfo := ar.GetTeamIncome(uid, page, size)
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

// 获取团队收益的列表
func (AccountRecord) GetTeamList(ctx iris.Context) {
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
	ar := &model.AccountRecord{}
	returnInfo, total := ar.GetTeamList(uid, page, size)
	_, _ = ctx.JSON(structs.NewResult(structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}))
}
