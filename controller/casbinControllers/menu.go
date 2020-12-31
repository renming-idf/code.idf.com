package casbinControllers

import (
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"xdf/common"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
	"xdf/validates"
)

type Menu struct {
}

// 获取页面
func (Menu) MenuInfo(ctx iris.Context) {
	m, ok := middleware.ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(401)
		return
	}
	aid := cast.ToUint(m["aid"])
	if aid < 1 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("ID错误")))
		return
	}
	// 获取角色id
	roleIds := model.GetRolesForUser(aid)
	roleID := cast.ToUint(roleIds[0])
	// 判断是否为超级管理员
	var menuData []model.Menu
	var err error
	menu := &model.Menu{}
	if roleID == common.SUPER_ADMIN_ID {
		menuData, _ = menu.GetAllMenu()
		if len(menuData) == 0 {
			menuModelTop := model.Menu{Status: 1, ParentID: 0, Name: "Top", Component: "", Key: "", Redirect: "", Sequence: 1, Icon: ""}
			if err := menu.CreateMenu(&menuModelTop); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}
			menuModelHome := model.Menu{Status: 1, ParentID: menuModelTop.ID, Name: "主页信息", Component: "Home", Key: "Home", Redirect: "/home", Sequence: 1, Icon: "lock"}
			if err := menu.CreateMenu(&menuModelHome); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}
			menuModel := model.Menu{Status: 1, ParentID: menuModelHome.ID, Name: "控制台", Component: "Controller", Key: "Controller", Redirect: "/home/controller", Sequence: 2, Icon: "lock"}
			if err := menu.CreateMenu(&menuModel); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}

			menuModelUser := model.Menu{Status: 1, ParentID: menuModelTop.ID, Name: "用户管理", Component: "User", Key: "User", Redirect: "/user", Sequence: 10, Icon: "user"}
			if err := menu.CreateMenu(&menuModelUser); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}
			menuModel = model.Menu{Status: 1, ParentID: menuModelUser.ID, Name: "用户列表", Component: "UserList", Key: "UserList", Redirect: "/user/list", Sequence: 11, Icon: "table"}
			if err := menu.CreateMenu(&menuModel); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}

			menuModelRole := model.Menu{Status: 1, ParentID: menuModelTop.ID, Name: "角色管理", Component: "Role", Key: "Role", Redirect: "/role", Sequence: 20, Icon: "user"}
			if err := menu.CreateMenu(&menuModelRole); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}
			menuModel = model.Menu{Status: 1, ParentID: menuModelRole.ID, Name: "角色列表", Component: "RoleList", Key: "RoleList", Redirect: "/role/list", Sequence: 21, Icon: "table"}
			if err := menu.CreateMenu(&menuModel); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}

			menuModelMenu := model.Menu{Status: 1, ParentID: menuModelTop.ID, Name: "页面与权限", Component: "Menu", Key: "Menu", Redirect: "/menu", Sequence: 30, Icon: "documentation"}
			if err := menu.CreateMenu(&menuModelMenu); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}
			menuModel = model.Menu{Status: 1, ParentID: menuModelMenu.ID, Name: "页面管理", Component: "MenuList", Key: "MenuList", Redirect: "/menu/list", Sequence: 31, Icon: "table"}
			if err := menu.CreateMenu(&menuModel); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}
			menuModel = model.Menu{Status: 1, ParentID: menuModelMenu.ID, Name: "权限管理", Component: "MenuPermission", Key: "MenuPermission", Redirect: "/menu/permission", Sequence: 32, Icon: "table"}
			if err := menu.CreateMenu(&menuModel); err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}

			//menuModelMerchant := model.Menu{Status: 1, ParentID: menuModelTop.ID, Name: "商户管理", Component: "Merchant", Key: "Merchant", Redirect: "/merchant", Sequence: 40, Icon: "user"}
			//if err := menu.CreateMenu(&menuModelMerchant); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//menuModel = model.Menu{Status: 1, ParentID: menuModelMerchant.ID, Name: "商户信息", Component: "MerchantInfo", Key: "MerchantInfo", Redirect: "/merchant/info", Sequence: 41, Icon: "user"}
			//if err := menu.CreateMenu(&menuModel); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//menuModel = model.Menu{Status: 1, ParentID: menuModelMerchant.ID, Name: "代理商", Component: "MerchantAgent", Key: "MerchantAgent", Redirect: "/merchant/agent", Sequence: 42, Icon: "user"}
			//if err := menu.CreateMenu(&menuModel); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//menuModel = model.Menu{Status: 1, ParentID: menuModelMerchant.ID, Name: "商户列表", Component: "MerchantList", Key: "MerchantList", Redirect: "/merchant/list", Sequence: 43, Icon: "documentation"}
			//if err := menu.CreateMenu(&menuModel); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//
			//menuModelOrder := model.Menu{Status: 1, ParentID: menuModelTop.ID, Name: "订单管理", Component: "Order", Key: "Order", Redirect: "/order", Sequence: 50, Icon: "documentation"}
			//if err := menu.CreateMenu(&menuModelOrder); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//menuModel = model.Menu{Status: 1, ParentID: menuModelOrder.ID, Name: "订单列表", Component: "OrderList", Key: "OrderList", Redirect: "/order/list", Sequence: 51, Icon: "documentation"}
			//if err := menu.CreateMenu(&menuModel); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//
			//menuModelRecord := model.Menu{Status: 1, ParentID: menuModelTop.ID, Name: "流水管理", Component: "Record", Key: "Record", Redirect: "/record", Sequence: 60, Icon: "documentation"}
			//if err := menu.CreateMenu(&menuModelRecord); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//menuModel = model.Menu{Status: 1, ParentID: menuModelRecord.ID, Name: "流水列表", Component: "RecordList", Key: "RecordList", Redirect: "/record/list", Sequence: 61, Icon: "documentation"}
			//if err := menu.CreateMenu(&menuModel); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//menuModel = model.Menu{Status: 1, ParentID: menuModelRecord.ID, Name: "提现记录表", Component: "RecordWithdrawal", Key: "RecordWithdrawal", Redirect: "/record/withdrawal", Sequence: 62, Icon: "documentation"}
			//if err := menu.CreateMenu(&menuModel); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//
			//menuModelCapitalChain := model.Menu{Status: 1, ParentID: menuModelTop.ID, Name: "财务管理", Component: "CapitalChain", Key: "CapitalChain", Redirect: "/capitalChain", Sequence: 70, Icon: "documentation"}
			//if err := menu.CreateMenu(&menuModelCapitalChain); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			//menuModel = model.Menu{Status: 1, ParentID: menuModelCapitalChain.ID, Name: "资金变动", Component: "CapitalChainList", Key: "CapitalChainList", Redirect: "/capitalChain/list", Sequence: 71, Icon: "documentation"}
			//if err := menu.CreateMenu(&menuModel); err != nil {
			//	_, _ = ctx.JSON(structs.NewResult(err))
			//	return
			//}
			menuData, err = menu.GetAllMenu()
			if err != nil {
				_, _ = ctx.JSON(structs.NewResult(err))
				return
			}
		}
	} else {
		menuData, err = menu.GetMenusByAdminsID(roleID)
		if err != nil {
			_, _ = ctx.JSON(structs.NewResult(err))
			return
		}
	}
	var menus []model.MenuModel
	if len(menuData) > 0 {
		var topMenuId = menuData[0].ParentID
		if topMenuId == 0 {
			topMenuId = menuData[0].ID
		}
		menus = menu.SetMenu(menuData, topMenuId)
	}
	_, _ = ctx.JSON(structs.NewResult(menus))
}

// 获取所有页面
func (Menu) GetMenu(ctx iris.Context) {
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 || size > 10 {
		size = 10
	}
	name := ctx.URLParam("name")
	parentId := cast.ToUint(ctx.URLParam("parentId"))
	if parentId < 0 {
		parentId = 1
	}
	m := &model.Menu{}
	menuSlice, total := m.GetAllMenuList(page, size, parentId, name)
	p := structs.PageMent{Page: page, Size: size, Total: total, Data: menuSlice}
	ctx.JSON(structs.NewResult(p))
}

func (Menu) GetMenuWithOutPage(ctx iris.Context) {
	menu := &model.Menu{}
	menuData, err := menu.GetAllMenu()
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	var menus []model.MenuModel
	if len(menuData) > 0 {
		var topMenuId = menuData[0].ParentID
		if topMenuId == 0 {
			topMenuId = menuData[0].ID
		}
		menus = menu.SetMenu(menuData, topMenuId)
	}
	_, _ = ctx.JSON(structs.NewResult(menus))
}

// 新增页面/修改页面
func (Menu) CreateMenu(ctx iris.Context) {
	aul := new(validates.CreateMenuRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	m := &model.Menu{
		Status:    1,
		Memo:      aul.Memo,
		ParentID:  aul.ParentID,
		Key:       aul.Key,
		Name:      aul.Name,
		Sequence:  aul.Sequence,
		Component: aul.Component,
		Redirect:  aul.Redirect,
		Icon:      aul.Icon,
	}
	if aul.ID == 0 {
		err = m.CreateMenu(m)
	} else {
		m.ID = aul.ID
		err = m.UpdateMenu(m)
	}
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("操作成功"))
}

// 查看详情
func (Menu) GetMenuInfo(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	m := &model.Menu{}
	m.ID = id
	m.GetMenuInfo()
	_, _ = ctx.JSON(structs.NewResult(m))
}

// 删除页面
func (Menu) DeleteMenu(ctx iris.Context) {
	var ids []uint
	err := ctx.ReadJSON(&ids)
	if err != nil || len(ids) == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("参数错误")))
		return
	}
	m := &model.Menu{}
	err = m.DeleteMenu(ids)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("删除成功"))
}
