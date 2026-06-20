package common

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"
)

type ISerializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, ptr interface{}) error
}

var (
	_jsonSerializer  = jsonSerializer{}
	_protoSerializer = protoSerializer{}
)

type jsonSerializer struct{}

func (s *jsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *jsonSerializer) Unmarshal(data []byte, ptr interface{}) error {
	return json.Unmarshal(data, ptr)
}

// errSentinel is an unexported string type for read-only error constants.
// External packages cannot create values of this type, preventing reassignment.
type errSentinel string

func (e errSentinel) Error() string { return string(e) }

const (
	ErrProtobuf    = errSentinel("value is not a protobuf message")
	ErrProtobufPtr = errSentinel("dest is not a protobuf message")
)

type protoSerializer struct{}

func (s *protoSerializer) Marshal(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, ErrProtobuf
	}
	return proto.Marshal(msg)
}

func (s *protoSerializer) Unmarshal(data []byte, ptr interface{}) error {
	msg, ok := ptr.(proto.Message)
	if !ok {
		return ErrProtobufPtr
	}
	return proto.Unmarshal(data, msg)
}

var defaultSerializer ISerializer

func init() {
	// 默认使用 JSON 序列化，用户可以通过 SetDefaultProtoSerializer 切换到 Protobuf。
	defaultSerializer = &_jsonSerializer
}

func SetDefaultProtoSerializer() {
	defaultSerializer = &_protoSerializer
}

func Marshal(v interface{}) ([]byte, error) {
	return defaultSerializer.Marshal(v)
}

func Unmarshal(data []byte, ptr interface{}) error {
	return defaultSerializer.Unmarshal(data, ptr)
}

func JsonMarshal(v interface{}) ([]byte, error) {
	return _jsonSerializer.Marshal(v)
}

func JsonUnmarshal(data []byte, ptr interface{}) error {
	return _jsonSerializer.Unmarshal(data, ptr)
}

func ProtoMarshal(v interface{}) ([]byte, error) {
	return _protoSerializer.Marshal(v)
}

func ProtoUnmarshal(data []byte, ptr interface{}) error {
	return _protoSerializer.Unmarshal(data, ptr)
}
