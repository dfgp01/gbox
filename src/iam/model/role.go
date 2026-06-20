package model

import "gbox/common"

// IamRole 角色表。
//
// 字段说明：
//
//	Name        - 角色名称
//	Description - 角色描述
type IamRole struct {
	common.BaseRDB
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (IamRole) TableName() string {
	return "iam_role"
}
