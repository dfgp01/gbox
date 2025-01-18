package msg

import (
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	ErrProtobuf      = errors.New("can not convert protobuf message")
	ErrProtobufSlice = errors.New("can not convert protobuf slice")
	ErrDecoder       = errors.New("dest must be pointer")
	ErrEncoder       = errors.New("obj is nil")

	ErrNotNumberSlice = errors.New("not number or number slice")
	ErrNotStringSlice = errors.New("not string or string slice")
	ErrNotStructSlice = errors.New("not struct or struct slice")
)

var (
	JsonSerializer  = &jsonSerializer{}
	ProtoSerializer = &protoSerializer{}
)

type (
	ISerializer interface {
		Marshal(v interface{}) ([]byte, error)
		UnMarshal(data []byte, ptr interface{}) error
	}
	jsonSerializer  struct{}
	protoSerializer struct{}
)

func (s *jsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *jsonSerializer) UnMarshal(data []byte, ptr interface{}) error {
	return json.Unmarshal(data, ptr)
}

func (s *protoSerializer) Marshal(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, ErrProtobuf
	}
	return proto.Marshal(msg)
}

func (s *protoSerializer) UnMarshal(data []byte, ptr interface{}) error {
	msg, ok := ptr.(proto.Message)
	if !ok {
		return ErrProtobuf
	}
	return proto.Unmarshal(data, msg)
}
