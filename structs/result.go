package structs

type Result struct {
	Status bool
	Msg    string
	Data   interface{}
}

/*
d为通用返回数据，e为可选参数，e不为空时返回失败result
*/
func NewResult(d interface{}) *Result {
	r := &Result{}

	dt, ok := d.(error)
	if ok {
		r.Msg = dt.Error()
		return r
	}

	r.Status = true
	r.Msg = "SUCCESS"
	r.Data = d
	return r
}
