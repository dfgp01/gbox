package msg

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

type (
	HashAvailable interface {
		ToNumber() int64
		ToString() string
		ToBytes() []byte
	}
	ByteHash   []byte
	StringHash string
	NumberHash int64
)

func (s ByteHash) ToNumber() int64 {
	si := s
	if len(s) > 8 {
		si = s[:8]
	}
	return int64(binary.BigEndian.Uint32(si))
}

func (s ByteHash) ToString() string {
	return string(s)
}

func (s ByteHash) ToBytes() []byte {
	return s
}

func (s StringHash) ToNumber() int64 {
	return int64(crc32.ChecksumIEEE([]byte(s)))
}

func (s StringHash) ToString() string {
	return string(s)
}

func (s StringHash) ToBytes() []byte {
	return []byte(s)
}

func (s NumberHash) ToNumber() int64 {
	return int64(s)
}

func (s NumberHash) ToString() string {
	return fmt.Sprint(s)
}

func (s NumberHash) ToBytes() []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(s))
	return b
}
