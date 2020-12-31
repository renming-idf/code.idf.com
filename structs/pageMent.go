package structs

type PageMent struct {
	Page  int
	Size  int
	Total int
	Data  interface{}
}

func (p *PageMent) New(page, s, t int, d interface{}) {
	p.Data = d
	p.Page = page
	p.Size = s
	p.Total = t
}
