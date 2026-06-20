package common

import "time"

// BaseRDB 数据表基础结构体，所有表 VO 均应内嵌此结构体以统一管理公共字段。
//
// 字段说明：
//
//	ID         - 主键 ID
//	CreateAt - 创建时间
//	UpdateAt - 更新时间
//	DeleteAt - 删除时间（软删除用，nil 表示未删除）
type BaseRDB struct {
	ID       uint       `json:"id"`
	CreateAt time.Time  `json:"create_at"`
	UpdateAt time.Time  `json:"update_at"`
	DeleteAt *time.Time `json:"delete_at"`
}
