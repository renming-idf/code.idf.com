package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/model"
	"xdf/structs"
)

type Announcement struct {
}

func (Announcement) GetAnnouncementList(ctx iris.Context) {
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 || size > 10 {
		size = 10
	}
	a := &model.Announcement{}
	returnInfo, total := a.GetAnnouncementList(page, size)
	p := structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}
	ctx.JSON(structs.NewResult(p))
}

func (Announcement) GetAnnouncementInfo(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	if id == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("公告ID错误")))
		return
	}
	a := &model.Announcement{}
	a, err := a.GetAnnouncementInfo(id)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(a))
}
