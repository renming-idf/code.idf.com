package casbinControllers

import (
	"errors"
	"github.com/kataras/iris/v12"
	gf "github.com/snowlyg/gotransformer"
	"github.com/spf13/cast"
	"time"
	"xdf/model"
	"xdf/structs"
	"xdf/transformer"
	"xdf/validates"
)

type Permissions struct {
}

// 通过ID查询权限
func (Permissions) GetPermission(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	p := &model.Permission{}
	permission := p.GetPermissionById(id)
	_, _ = ctx.JSON(structs.NewResult(permission))
}

// 创建权限
func (Permissions) CreatePermission(ctx iris.Context) {
	aul := new(validates.PermissionRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	p := &model.Permission{}
	pm := p.CreatePermission(aul)
	if pm.ID == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("创建失败")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("创建成功"))
}

// 更新权限
func (Permissions) UpdatePermission(ctx iris.Context) {
	aul := new(validates.PermissionRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}

	id, err := ctx.Params().GetUint("id")
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id错误")))
		return
	}
	p := &model.Permission{}
	err = p.UpdatePermission(aul, id)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("更新失败")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("更新成功"))
}

// 删除权限
func (Permissions) DeletePermission(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	p := &model.Permission{}
	p.ID = id
	err := p.DeletePermission()
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("删除失败")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("删除成功"))
}

// 获取所有权限
func (Permissions) GetAllPermissions(ctx iris.Context) {
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 || size > 10 {
		size = 10
	}
	name := ctx.URLParam("name")
	p := &model.Permission{}
	permissionSlice, total := p.GetAllPermissions(page, size, name)
	pm := structs.PageMent{Page: page, Size: size, Total: total, Data: permsTransform(permissionSlice)}
	_, _ = ctx.JSON(structs.NewResult(pm))
}

// 获取所有权限
func (Permissions) GetAllPermissionsWithOutPage(ctx iris.Context) {
	p := &model.Permission{}
	permissionSlice := p.GetAllPermissionsWithOutPage()
	_, _ = ctx.JSON(structs.NewResult(permissionSlice))
}

func permsTransform(perms []*model.Permission) []*transformer.Permission {
	var rs []*transformer.Permission
	for _, perm := range perms {
		r := permTransform(perm)
		rs = append(rs, r)
	}
	return rs
}

func permTransform(perm *model.Permission) *transformer.Permission {
	r := &transformer.Permission{}
	g := gf.NewTransform(r, perm, time.RFC3339)
	_ = g.Transformer()
	return r
}
