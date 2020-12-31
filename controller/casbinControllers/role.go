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

type Role struct {
}

// 查看角色详情
func (Role) GetRole(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	r := &model.Role{}
	role := r.GetRoleById(id)

	rr := roleTransform(role)
	rr.Perms = permsTransform(r.RolePermissions(role.ID))
	_, _ = ctx.JSON(structs.NewResult(rr))

}

// 新建角色
func (Role) CreateRole(ctx iris.Context) {
	aul := new(validates.RoleRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	r := &model.Role{}
	u := r.CreateRole(aul, aul.PermissionsIds)
	ctx.StatusCode(iris.StatusOK)
	if u.ID == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("操作失败")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("创建成功"))
}

// 更新角色
func (Role) UpdateRole(ctx iris.Context) {
	aul := new(validates.RoleRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	r := &model.Role{}

	id, _ := ctx.Params().GetUint("id")
	//role := r.GetRoleById(id)
	//if role.Name == "admin" {
	//	_, _ = ctx.JSON(structs.NewResult(errors.New("不能编辑管理员角色")))
	//	return
	//}

	u := r.UpdateRole(aul, id, aul.PermissionsIds)
	if u.ID == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("操作失败")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("更新成功"))

}

// 删除角色
func (Role) DeleteRole(ctx iris.Context) {
	id, err := ctx.Params().GetUint("id")
	if id <= 0 || err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("id参数错误")))
		return
	}
	r := &model.Role{}
	role := r.GetRoleById(id)
	if role.Name == "admin" || role.Name == "merchant" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("不能删除管理员或商户角色")))
		return
	}
	err = r.DeleteRoleById(id)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("删除成功"))
}

// 获取所有角色
func (Role) GetAllRoles(ctx iris.Context) {
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 || size > 10 {
		size = 10
	}
	name := ctx.URLParam("name")
	r := &model.Role{}
	roleSlice, total := r.GetAllRole(page, size, name)
	p := structs.PageMent{Page: page, Size: size, Total: total, Data: rolesTransform(roleSlice)}
	_, _ = ctx.JSON(structs.NewResult(p))
}

// 获取所有角色
func (Role) GetAllRolesWithOutPage(ctx iris.Context) {
	r := &model.Role{}
	roleSlice := r.GetAllRoleWithOutPage()
	_, _ = ctx.JSON(structs.NewResult(roleSlice))
}

// 查看该角色下所有的菜单ID
func (Role) GetRoleMenu(ctx iris.Context) {
	roleID := cast.ToUint(ctx.URLParam("roleID"))
	if roleID < 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("角色错误")))
		return
	}
	m := model.Menu{}
	menuIDList, err := m.GetAllMenuByRole(roleID)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(menuIDList))
}

// 设置角色的菜单
func (Role) SetRoleMenu(ctx iris.Context) {
	type tmpStruct struct {
		RoleID      uint   `json:"role_id"`
		MenuIdSlice []uint `json:"menu_id_slice"`
	}
	tmp := new(tmpStruct)
	err := ctx.ReadJSON(&tmp)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	mr := model.RoleMenu{}
	err = mr.SetRole(tmp.RoleID, tmp.MenuIdSlice)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("操作成功"))
}

func rolesTransform(roles []*model.Role) []*transformer.Role {
	var rs []*transformer.Role
	for _, role := range roles {
		r := roleTransform(role)
		rs = append(rs, r)
	}
	return rs
}

func roleTransform(role *model.Role) *transformer.Role {
	r := &transformer.Role{}
	g := gf.NewTransform(r, role, time.RFC3339)
	_ = g.Transformer()
	return r
}
