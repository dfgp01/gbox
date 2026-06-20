package common

import (
	"fmt"
)

// Type 表示哈希值的基础类型枚举。
type Type int8

const (
	TypeInt32  Type = iota // int32 类型
	TypeInt64              // int64 类型
	TypeString             // 字符串类型
	TypeBytes              // 字节切片类型
)

// HashVarable 是可变哈希值类型的接口，统一处理不同 ID 类型的转换。
type HashVarable interface {
	// Types 返回当前值的基础类型。
	Types() Type
	// Int32 以 int32 形式返回。
	Int32() int32
	// Int64 以 int64 形式返回。
	Int64() int64
	// String 以字符串形式返回。
	String() string
	// Bytes 以 []byte 形式返回。
	Bytes() []byte
}

// Int32ID 基于 int32 的哈希 ID。
type Int32ID int32

func (h Int32ID) Types() Type    { return TypeInt32 }
func (h Int32ID) Int32() int32   { return int32(h) }
func (h Int32ID) Int64() int64   { return int64(h) }
func (h Int32ID) String() string { return fmt.Sprintf(`"%d"`, int32(h)) }
func (h Int32ID) Bytes() []byte  { return []byte(fmt.Sprintf(`"%d"`, int32(h))) }

// Int64ID 基于 int64 的哈希 ID。
type Int64ID int64

func (h Int64ID) Types() Type    { return TypeInt64 }
func (h Int64ID) Int32() int32   { return int32(h) }
func (h Int64ID) Int64() int64   { return int64(h) }
func (h Int64ID) String() string { return fmt.Sprintf(`"%d"`, int64(h)) }
func (h Int64ID) Bytes() []byte  { return []byte(fmt.Sprintf(`"%d"`, int64(h))) }

// StringID 基于字符串的哈希 ID。
type StringID string

func (h StringID) Types() Type    { return TypeString }
func (h StringID) Int32() int32   { return 0 }
func (h StringID) Int64() int64   { return 0 }
func (h StringID) String() string { return string(h) }
func (h StringID) Bytes() []byte  { return []byte(string(h)) }

// BytesID 基于 []byte 的哈希 ID。
type BytesID []byte

func (h BytesID) Types() Type    { return TypeBytes }
func (h BytesID) Int32() int32   { return 0 }
func (h BytesID) Int64() int64   { return 0 }
func (h BytesID) String() string { return string(h) }
func (h BytesID) Bytes() []byte  { return append([]byte{}, h...) }
