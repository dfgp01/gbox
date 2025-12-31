package model

// `Pager` ，用於分頁查詢，不受具體數據庫約束
type Page struct {
	No    int `json:"no"`    //當前頁
	Size  int `json:"size"`  //每頁數量
	Count int `json:"count"` //總數據量
}

// 默认处理
func (p *Page) Default() {
	if p.No <= 0 {
		p.No = 1
	}
	if p.Size <= 0 {
		p.Size = 10
	}
}

func (p *Page) GetTotalPage() int {
	if p.Size == 0 {
		return 0
	}
	return (p.Count + p.Size - 1) / p.Size
}

type PageResult[T any] struct {
	Page
	Items []T `json:"items"`
}
