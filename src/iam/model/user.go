package model

import "gbox/common"

// IamUser 用户表，用户源在外部系统，IAM 仅记录来源与映射关系。
//
// 字段说明：
//
//	AppID - 用户来源的应用标识
//	Sub   - 用户在外部系统中的唯一标识（subject）
type IamUser struct {
	common.BaseRDB
	AppID  int32 `json:"app_id"`
	UserID int32 `json:"user_id"`
}

func (IamUser) TableName() string {
	return "iam_user"
}
