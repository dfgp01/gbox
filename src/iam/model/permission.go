package model

import "gbox/common"

// IamPermission 权限表。
//
// 字段说明：
//
//	Resource    - 资源标识，例如 "user:read"、"role:write"
//	Description - 权限描述
type IamPermission struct {
	common.BaseRDB
	Resource    string `json:"resource"`
	Description string `json:"description"`
}

func (IamPermission) TableName() string {
	return "iam_permission"
}
