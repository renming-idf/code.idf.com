package validates

type CreateMenuRequest struct {
	ID        uint   `json:"id"  comment:"id"`
	Memo      string `json:"memo" validate:"required,gte=1,lte=64" comment:"备注"`
	ParentID  uint   `json:"parent_id" comment:"上级菜单ID"`
	Component string `json:"component" validate:"required,gte=1,lte=72" comment:"component"`
	Key       string `json:"key" validate:"required,gte=1,lte=72" comment:"key"`
	Redirect  string `json:"redirect" validate:"required,gte=1,lte=72" comment:"redirect"`
	Name      string `json:"name" validate:"required,gte=1,lte=32" comment:"菜单名称"`
	Sequence  int    `json:"sequence" comment:"排序值"`
	Icon      string `json:"icon" validate:"required,gte=1,lte=32" comment:"icon"`
}
