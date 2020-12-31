package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"xdf/common"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
)

type Version struct {
}

func (Version) GetVersion(ctx iris.Context) {
	platform := ctx.URLParam("platform")
	if platform == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("platform 不能为空！")))
		return
	}
	v := &model.IniVersion{}
	v.GetVersionByPlatform(platform)
	_, _ = ctx.JSON(structs.NewResult(v))
}

func (Version) Share(ctx iris.Context) {
	platform := ctx.URLParam("platform")
	if platform == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("platform 不能为空！")))
		return
	}
	v := &model.IniVersion{}
	v.GetDownloadAddressByPlatform(platform)
	returnMap := make(map[string]string)
	returnMap["download_address"] = v.DownloadAddress
	s, err := common.UrlToBase(v.DownloadAddress)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	returnMap["download_address_code"] = s
	_, _ = ctx.JSON(structs.NewResult(returnMap))
}

func (Version) GetShareAddress(ctx iris.Context) {
	platform := ctx.URLParam("platform")
	if platform == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("platform 不能为空！")))
		return
	}
	v := &model.IniVersion{}
	v.GetDownloadAddressByPlatform(platform)
	returnMap := make(map[string]string)
	returnMap["share_address"] = v.ShareAddress
	s, err := common.UrlToBase(v.ShareAddress)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	returnMap["share_address_code"] = s
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
	returnMap["invite_code"] = u.GetUserInvitationCode(uid)
	_, _ = ctx.JSON(structs.NewResult(returnMap))
}

func (Version) GetDownloadAddress(ctx iris.Context) {
	v := &model.IniVersion{}
	vSlice := v.GetVersion()
	_, _ = ctx.JSON(structs.NewResult(vSlice))
}

func (Version) GetOfficialAddress(ctx iris.Context) {
	v := &model.IniVersion{}
	returnInfo := v.GetOfficialVersion()
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}
