package model

import "errors"

type Page struct {
	No    int `json:"no"`    //當前頁
	Size  int `json:"size"`  //每頁數量
	Count int `json:"count"` //總條數
}

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

var ErrRecordNotFound = errors.New("record not found")
