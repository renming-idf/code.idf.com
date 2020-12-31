package controller

import (
	"errors"
	"github.com/kataras/iris/v12"
	gf "github.com/snowlyg/gotransformer"
	"github.com/spf13/cast"
	"github.com/tyler-smith/go-bip39"
	"strings"
	"time"
	"xdf/common"
	"xdf/common/log"
	"xdf/common/watch"
	"xdf/middleware"
	"xdf/model"
	"xdf/structs"
	"xdf/transformer"
	"xdf/validates"
)

type User struct {
}

func (User) FlashToken(ctx iris.Context) {
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
	user := model.User{}
	user.GetUserById(uid)
	var sessionID = middleware.SMgr.StartSession(ctx.ResponseWriter(), 1)
	lt := time.Now().Add(24 * 7 * time.Hour).Unix()
	tokenMap := map[string]interface{}{"aid": uid, "exp": lt, "session": sessionID, "type": 1}
	tokenString := middleware.CreateToken(tokenMap, 1)
	middleware.SMgr.SetSessionVal(sessionID, "UserInfo", user, 1)
	s := &model.SessionInfo{}
	err := s.SaveSessionInfo(uid, sessionID, tokenString, 1, lt)
	if err != nil {
		log.Println("记录登录状态失败")
	}
	_, _ = ctx.JSON(structs.NewResult(tokenString))
}

// 获取验证码
func (User) GetCode(ctx iris.Context) {
	phone := ctx.URLParam("phone")
	tp := ctx.URLParam("type")
	if phone == "" || len(phone) != 11 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("请输入正确的手机号")))
		return
	}
	// 验证该手机号是否已经注册
	u := &model.User{}
	u.GetUserByPhone(phone)
	if tp == "register" {
		if u.ID > 0 {
			_, _ = ctx.JSON(structs.NewResult(errors.New("手机号已注册")))
			return
		}
	} else {
		if u.ID == 0 {
			_, _ = ctx.JSON(structs.NewResult(errors.New("手机号未注册")))
			return
		}
	}
	err := common.GetCode(phone)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("短信发送失败，请联系管理员")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("发送成功"))
}

// 重置密码
func (User) ChangeLoginPassword(ctx iris.Context) {
	aul := new(validates.ChangeLoginPasswordRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	checkCode, ok := common.CodeCheckMap.Load(aul.Phone)
	if !ok {
		_, _ = ctx.JSON(structs.NewResult(errors.New("验证码过期")))
		return
	}
	if aul.MessageCode != cast.ToString(checkCode) {
		_, _ = ctx.JSON(structs.NewResult(errors.New("验证码输入错误")))
		return
	}
	u := &model.User{}
	err = u.ChangeLoginPassword(aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("修改成功"))
}

// 创建账户
func (User) CreateUser(ctx iris.Context) {
	ip := ctx.RemoteAddr()
	aul := new(validates.CreateAppUserRequest)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	if len(aul.Password) != 6 || !common.IsNum(aul.Password) {
		_, _ = ctx.JSON(structs.NewResult("请输入6位且仅含数字的安全密码"))
		return
	}
	if !common.CheckPwd(aul.LoginPassword) || len(aul.LoginPassword) > 16 || len(aul.LoginPassword) < 8 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("请输入8~16位，含英文数字的登录密码")))
		return
	}
	checkCode, ok := common.CodeCheckMap.Load(aul.Phone)
	if !ok {
		_, _ = ctx.JSON(structs.NewResult(errors.New("验证码过期")))
		return
	}
	if aul.MessageCode != cast.ToString(checkCode) {
		_, _ = ctx.JSON(structs.NewResult(errors.New("验证码输入错误")))
		return
	}
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	aul.Mnemonic = mnemonic
	pks, addr, e := watch.NewAccount()
	if e != nil {
		_, _ = ctx.JSON(structs.NewResult(e))
		return
	}
	u := &model.User{}
	user, err := u.CreateUser(aul, addr, pks, ip)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	// 如果为注册页注册，直接返回助记词和私钥
	if aul.Type == 1 {
		tmpMap := make(map[string]interface{})
		tmpMap["userInfo"] = userTransform(&user)
		_, _ = ctx.JSON(structs.NewResult(tmpMap))
		return
	}
	//向钱包池中插入钱包
	go watch.InjectAccountAddressToPool(user)
	var sessionID = middleware.SMgr.StartSession(ctx.ResponseWriter(), 1)
	lt := time.Now().Add(24 * 7 * time.Hour).Unix()
	tokenMap := map[string]interface{}{"aid": user.ID, "exp": lt, "session": sessionID, "type": 1}
	tokenString := middleware.CreateToken(tokenMap, 1)
	middleware.SMgr.SetSessionVal(sessionID, "UserInfo", user, 1)
	s := &model.SessionInfo{}
	err = s.SaveSessionInfo(user.ID, sessionID, tokenString, 1, lt)
	if err != nil {
		log.Println("记录登录状态失败")
	}
	tmpMap := make(map[string]interface{})
	tmpMap["token"] = tokenString
	tmpMap["userInfo"] = userTransform(&user)
	// 获取公告-最新
	a := &model.Announcement{}
	a.GetFirstAnnouncement()
	tmpMap["announcement"] = a
	c := &model.CurrencyType{}
	var mingct model.IniAllowMining
	cid := mingct.GetAllowMingCtID()
	hi := &model.HotInfo{}
	tmpMap["xdf_info"] = hi.GetAllRateByCurrencyTypeID(cid)
	ar := &model.AccountRecord{}
	tmpMap["ecological_income"] = ar.GetEcologicalIncome(user.ID)
	xdfInfo := c.GetCurrencyTypeByID(cid)
	tmpMap["ecological_income_cny"] = int(xdfInfo.CnyRate * cast.ToFloat64(tmpMap["ecological_income"]))
	v := &model.IniVersion{}
	tmpMap["version"] = v.GetVersion()
	b := &model.Banner{}
	tmpMap["banner"] = b.GetBanner()
	_, _ = ctx.JSON(structs.NewResult(tmpMap))
}

func (User) Import(ctx iris.Context) {
	aul := new(validates.ImportUSer)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	user := model.User{}
	user.GetUserByPhoneAndLoginPassword(aul.Phone, aul.LoginPassword)
	if user.ID == 0 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("请输入正确的手机号或密码")))
		return
	}
	BackManyKickPeople(int(user.ID))
	var sessionID = middleware.SMgr.StartSession(ctx.ResponseWriter(), 1)
	lt := time.Now().Add(24 * 7 * time.Hour).Unix()
	tokenMap := map[string]interface{}{"aid": user.ID, "exp": lt, "session": sessionID, "type": 1}
	tokenString := middleware.CreateToken(tokenMap, 1)
	middleware.SMgr.SetSessionVal(sessionID, "UserInfo", user, 1)
	s := &model.SessionInfo{}
	err = s.SaveSessionInfo(user.ID, sessionID, tokenString, 1, lt)
	if err != nil {
		log.Println("记录登录状态失败")
	}
	tmpMap := make(map[string]interface{})
	tmpMap["token"] = tokenString
	tmpMap["userInfo"] = userTransform(&user)
	// 获取公告-最新
	a := &model.Announcement{}
	a.GetFirstAnnouncement()
	tmpMap["announcement"] = a
	c := &model.CurrencyType{}
	var mingct model.IniAllowMining
	cid := mingct.GetAllowMingCtID()
	hi := &model.HotInfo{}
	tmpMap["xdf_info"] = hi.GetAllRateByCurrencyTypeID(cid)
	ar := &model.AccountRecord{}
	tmpMap["ecological_income"] = ar.GetEcologicalIncome(user.ID)
	xdfInfo := c.GetCurrencyTypeByID(cid)
	tmpMap["ecological_income_cny"] = int(xdfInfo.CnyRate * cast.ToFloat64(tmpMap["ecological_income"]))
	v := &model.IniVersion{}
	tmpMap["version"] = v.GetVersion()
	b := &model.Banner{}
	tmpMap["banner"] = b.GetBanner()
	_, _ = ctx.JSON(structs.NewResult(tmpMap))
}

func (User) GetPreachInfo(ctx iris.Context) {
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
	au := &model.User{}
	users := au.AllSon(uid, page, size)
	isEnd := len(users) < size
	tmpMap := make(map[string]interface{})
	ar := &model.AccountRecord{}
	tmpMap["is_end"] = isEnd
	tmpMap["preach_income"] = ar.GetPreachIncome(uid)
	ct := &model.CurrencyType{}
	iam := &model.IniAllowMining{}
	ctInfo := ct.GetCurrencyTypeByID(iam.GetAllowMingCtID())
	tmpMap["preach_income_cny"] = int(ctInfo.CnyRate * cast.ToFloat64(tmpMap["preach_income"]))
	tmpMap["info"] = users
	_, _ = ctx.JSON(structs.NewResult(tmpMap))
}

// aul.type 1 导出助记词  2 导出密匙
func (User) GetUserDetails(ctx iris.Context) {
	aul := new(validates.GetUserDetail)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
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
	u := &model.User{}
	err = u.CheckPasswordByID(uid, aul.Password)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	user := u.GetUserById(uid)
	dataMap := make(map[string]string)
	switch aul.Type {
	case 1:
		dataMap["mnemonic_words"] = user.MnemonicWords
		dataMap["public_key"] = user.PublicKey
		dataMap["private_key"] = user.PrivateKey
	case 2:
		dataMap["public_key"] = user.PublicKey
		dataMap["private_key"] = user.PrivateKey
	default:
		_, _ = ctx.JSON(structs.NewResult(errors.New("type错误")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(dataMap))
}

// 获取个人收款地址
func (User) GetUserPublicKey(ctx iris.Context) {
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
	u.GetUserPublicKey(uid)
	if u.PublicKey == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("用户钱包地址加载错误")))
		return
	}
	returnMap := make(map[string]string)
	returnMap["public_key"] = u.PublicKey
	s, err := common.UrlToBase(u.PublicKey)
	if err != nil {
		ctx.JSON(structs.NewResult(err))
		return
	}
	returnMap["public_key_code"] = s
	_, _ = ctx.JSON(structs.NewResult(returnMap))
}

// 获取充USDT地址
func (User) GetUserAccountAddress(ctx iris.Context) {
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
	u.GetUserAccountAddress(uid)
	if u.AccountAddress == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("用户充值USDT地址加载错误")))
		return
	}
	returnMap := make(map[string]string)
	returnMap["account_address"] = u.AccountAddress
	s, err := common.UrlToBase(u.AccountAddress)
	if err != nil {
		ctx.JSON(structs.NewResult(err))
		return
	}
	returnMap["account_address_code"] = s
	_, _ = ctx.JSON(structs.NewResult(returnMap))
}

// 刷新首页
func (User) GetHome(ctx iris.Context) {
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
	tmpMap := make(map[string]interface{})
	// 获取公告-最新
	a := &model.Announcement{}
	a.GetFirstAnnouncement()
	tmpMap["announcement"] = a
	var mingct model.IniAllowMining
	cid := mingct.GetAllowMingCtID()
	hi := &model.HotInfo{}
	tmpMap["xdf_info"] = hi.GetAllRateByCurrencyTypeID(cid)
	ar := &model.AccountRecord{}
	tmpMap["ecological_income"] = ar.GetEcologicalIncome(uid)
	c := &model.CurrencyType{}
	xdfInfo := c.GetCurrencyTypeByID(cid)
	tmpMap["ecological_income_cny"] = int(xdfInfo.CnyRate * cast.ToFloat64(tmpMap["ecological_income"]))
	b := &model.Banner{}
	tmpMap["banner"] = b.GetBanner()
	_, _ = ctx.JSON(structs.NewResult(tmpMap))
}

// 验证密码
func (User) CheckPassword(ctx iris.Context) {
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
	p := ctx.URLParam("password")
	u := &model.User{}
	err := u.CheckPasswordByID(uid, p)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("密码错误")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("密码正确"))
}

func (User) CheckLoginPasswordByID(ctx iris.Context) {
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
	p := ctx.URLParam("password")
	u := &model.User{}
	err := u.CheckLoginPasswordByID(uid, p)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("密码错误")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("密码正确"))
}

func (User) ChangePassword(ctx iris.Context) {
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
	aul := new(validates.ChangePassword)
	err := validates.Warp(ctx, aul)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	if aul.NewPassword == "" {
		_, _ = ctx.JSON(structs.NewResult(errors.New("密码不能为空！")))
		return
	}
	user := model.User{}
	user.GetUserById(uid)
	checkCode, ok := common.CodeCheckMap.Load(user.Phone)
	if !ok {
		_, _ = ctx.JSON(structs.NewResult(errors.New("验证码过期")))
		return
	}
	if aul.MessageCode != cast.ToString(checkCode) {
		_, _ = ctx.JSON(structs.NewResult(errors.New("验证码输入错误")))
		return
	}
	switch aul.Type {
	case 1:
		if !common.CheckPwd(aul.NewPassword) || len(aul.NewPassword) > 16 || len(aul.NewPassword) < 8 {
			_, _ = ctx.JSON(structs.NewResult(errors.New("请输入8~16位，含英文数字的登录密码")))
			return
		}
		err = user.ChangeLoginPasswordByPrivateKey(aul)
		if err != nil {
			_, _ = ctx.JSON(structs.NewResult(errors.New("修改失败！")))
			return
		}
	case 2:
		//if aul.NewPassword
		if len(aul.NewPassword) != 6 || !common.IsNum(aul.NewPassword) {
			_, _ = ctx.JSON(structs.NewResult("请输入6位且仅含数字的安全密码"))
			return
		}
		err = user.ChangePasswordByPrivateKey(aul)
		if err != nil {
			_, _ = ctx.JSON(structs.NewResult(errors.New("修改失败！")))
			return
		}
	default:
		_, _ = ctx.JSON(structs.NewResult(errors.New("请输入正确的type")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("修改密码成功！"))
}

func (User) GetReceive(ctx iris.Context) {
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
	tmp, err := u.GetReceive(uid)
	if err != nil {
		log.Error(err)
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult(tmp))

}

func (User) Receive(ctx iris.Context) {
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
	err := u.Receive(uid)
	if err != nil {
		log.Error(err)
		_, _ = ctx.JSON(structs.NewResult(err))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("领取成功"))
}

func (User) GetOnlineTime(ctx iris.Context) {
	_, _ = ctx.JSON(structs.NewResult("在线时间10:00-18:00"))
}

// 上传文件
func (User) PutFile(ctx iris.Context) {
	f, h, e := ctx.FormFile("file_name")
	if e != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("上传错误")))
		return
	}
	log.Println(h.Filename)
	log.Println(h.Size)
	if h.Size > 1024*1024*50 {
		_, _ = ctx.JSON(structs.NewResult(errors.New("文件大小不能超过50M")))
		return
	}
	log.Println("50M可以")
	fName := strings.ToLower(h.Filename)
	if !strings.HasSuffix(fName, ".mp4") && !strings.HasSuffix(fName, ".avi") && !strings.HasSuffix(fName, ".mov") && !strings.HasSuffix(fName, ".jpg") && !strings.HasSuffix(fName, ".png") && !strings.HasSuffix(fName, ".gif") && !strings.HasSuffix(fName, ".tif") && !strings.HasSuffix(fName, ".jpeg") {
		_, _ = ctx.JSON(structs.NewResult(errors.New("文件格式错误")))
		return
	}
	log.Println("文件格式可以")
	fileName, err := common.SavePictures(f, h)
	if err != nil {
		_, _ = ctx.JSON(structs.NewResult(errors.New("上传文件失败")))
		return
	}
	log.Println("文件格式上传成功")
	log.Println(fileName)
	ctx.JSON(structs.NewResult(fileName))
}

func (User) GetAllUserAccountAddress(ctx iris.Context) {
	page := cast.ToInt(ctx.URLParam("page"))
	if page < 1 {
		page = 1
	}
	size := cast.ToInt(ctx.URLParam("size"))
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	u := model.User{}
	accountAddressu := u.GetAllUserAccountAddress(page, size)
	_, _ = ctx.JSON(structs.NewResult(accountAddressu))
}

func (User) ChangeUserName(ctx iris.Context) {
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
	name := ctx.URLParam("name")
	u := &model.User{}
	err := u.ChangeUserName(uid, name)
	if err != nil {
		log.Error(err)
		_, _ = ctx.JSON(structs.NewResult(errors.New("修改名称失败！")))
		return
	}
	_, _ = ctx.JSON(structs.NewResult("修改名称成功！"))
}

func (User) GetUsdtGas(ctx iris.Context) {
	m := make(map[string]interface{})

	m["FastestPrice"] = watch.FastGasPrice
	m["SafeGasPrice"] = watch.SafeGasPrice

	_, _ = ctx.JSON(structs.NewResult(m))
}
func userTransform(user *model.User) *transformer.User {
	tu := &transformer.User{}
	g := gf.NewTransform(tu, user, time.RFC3339)
	_ = g.Transformer()
	tu.ID = user.ID
	tu.Name = user.Name
	tu.PrivateKey = user.PrivateKey
	tu.Phone = user.Phone
	tu.PublicKey = user.PublicKey
	tu.ParentID = user.ParentID
	return tu
}

//用于提出当前登录的用户
func BackManyKickPeople(userID ...int) {
	for _, v := range userID {
		token, ok := middleware.AppTokenMap.Load(uint(v))
		if ok {
			var sessionID = ""
			if token != "" {
				m, _ := middleware.ParseToken(cast.ToString(token), middleware.JwtKey)
				sessionID = m["session"]
				middleware.SMgr.EndSessionBy(sessionID, 1)
			}
		}
	}
}
