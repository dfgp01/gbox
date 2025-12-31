package model

import "errors"

// error部分，全数据库通用
var ErrRecordNotFound = errors.New("record not found")
