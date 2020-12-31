package casbinControllers

import (
	"encoding/hex"
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/common/tools"
	"xdf/model"
	"xdf/structs"
	"xdf/validates"
)

type User struct {
}

// 获取所有用户——分页
func (User) GetAllUserByList(ctx iris.Context) {
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 || size > 10 {
		size = 10
	}
	isForbidden := cast.ToBool(ctx.URLParam("is_forbidden"))
	isAuthentication := cast.ToBool(ctx.URLParam("is_authentication"))
	publicKey := ctx.URLParam("public_key")
	name := ctx.URLParam("name")
	id := cast.ToUint(ctx.URLParam("id"))
	parentID := cast.ToUint(ctx.URLParam("parent_id"))
	u := &model.User{}
	returnInfo, total := u.GetAllUserByList(page, size, isForbidden, isAuthentication, publicKey, name, id, parentID)
	p := structs.PageMent{Page: page, Size: size, Total: total, Data: returnInfo}
	_, _ = ctx.JSON(structs.NewResult(p))
}

// 修改用户信息
func (User) ChangeUserInfo(ctx iris.Context) {
	uid := cast.ToUint(ctx.URLParam("user_id"))
	password := ctx.URLParam("password")
	isForbidden := cast.ToInt(ctx.URLParam("is_forbidden"))
	if isForbidden < 0 || isForbidden > 2 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("is_forbidden参数错误")))
		return
	}
	u := &model.User{}
	err := u.ChangeUserInfo(uid, password, isForbidden)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("修改成功"))
}

// 不分页获取所有用户
func (User) GetAllUser(ctx iris.Context) {
	u := &model.User{}
	returnInfo := u.GetAllUser()
	_, _ = ctx.JSON(structs.NewResult(returnInfo))
}

func (User) GetPrivateKey(ctx iris.Context) {
	aul := new(validates.ImportUserPrivateKey)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	u := model.User{}
	u = u.GetUserByAccountAddress(aul.AccountAddress)
	if u.ID == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("请填写正确的钱包地址")))
		return
	}
	prv, err := tools.KeystoreToPrivateKey("./wallets"+u.KeyStorePath, aul.Password)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	privKey := hex.EncodeToString(prv.D.Bytes())
	_, _ = ctx.JSON(structs.NewResult(privKey))
}
