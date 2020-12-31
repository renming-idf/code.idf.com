package validates

type CreateAppUserRequest struct {
	Phone          string `json:"phone" validate:"required,gte=1,lte=13" comment:"手机号"`
	Name           string `json:"name" validate:"required,gte=1,lte=13" comment:"昵称"`
	LoginPassword  string `json:"login_password" validate:"required,gte=6,lte=50" comment:"登录密码"`
	Password       string `json:"password" validate:"required" comment:"安全密码"`
	Mnemonic       string `json:"mnemonic" comment:"助记词"`
	Type           int    `json:"type" validate:"required" comment:"注册类型"`
	InvitationCode string `json:"invitation_code" validate:"required" comment:"邀请码"`
	MessageCode    string `json:"message_code" validate:"required" comment:"短信验证码"`
}

type ChangeLoginPasswordRequest struct {
	Phone         string `json:"phone" validate:"required,gte=1,lte=13" comment:"手机号"`
	LoginPassword string `json:"login_password" validate:"required,gte=6,lte=50" comment:"登录密码"`
	MessageCode   string `json:"message_code" validate:"required" comment:"短信验证码"`
}

type GetUserDetail struct {
	Type     uint   `json:"type" validate:"required" comment:"type"` // 1 导出助记词  2 导出密匙
	Password string `json:"password" validate:"required"  comment:"密码"`
}

type ImportUSer struct {
	Phone         string `json:"phone" validate:"required" comment:"手机号"`
	LoginPassword string `json:"login_password" comment:"密码"`
}

type ChangePassword struct {
	Type        uint   `json:"type" validate:"required" comment:"type"` //1 登录密码 2、交易密码
	MessageCode string `json:"message_code" validate:"required" comment:"验证码"`
	NewPassword string `json:"new_password" validate:"required"  comment:"修改密码不能为空"`
}
