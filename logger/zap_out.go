package logger

import "go.uber.org/zap"

// fields 根据键值对生成 zap 字段
func fields(kv ...interface{}) []zap.Field {
	var fs []zap.Field
	if len(kv) == 0 || len(kv)%2 != 0 {
		return fs
	}

	for i := 0; i < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		value := kv[i+1]

		switch v := value.(type) {
		case string:
			fs = append(fs, zap.String(key, v))
		case int:
			fs = append(fs, zap.Int(key, v))
		case int8:
			fs = append(fs, zap.Int8(key, v))
		case int16:
			fs = append(fs, zap.Int16(key, v))
		case int32:
			fs = append(fs, zap.Int32(key, v))
		case int64:
			fs = append(fs, zap.Int64(key, v))
		case uint:
			fs = append(fs, zap.Uint(key, v))
		case uint8:
			fs = append(fs, zap.Uint8(key, v))
		case uint16:
			fs = append(fs, zap.Uint16(key, v))
		case uint32:
			fs = append(fs, zap.Uint32(key, v))
		case uint64:
			fs = append(fs, zap.Uint64(key, v))
		case float32:
			fs = append(fs, zap.Float32(key, v))
		case float64:
			fs = append(fs, zap.Float64(key, v))
		case bool:
			fs = append(fs, zap.Bool(key, v))
		case []byte:
			fs = append(fs, zap.ByteString(key, v))
		case []string:
			fs = append(fs, zap.Strings(key, v))
		case []int:
			fs = append(fs, zap.Ints(key, v))
		case []int8:
			fs = append(fs, zap.Int8s(key, v))
		case []int16:
			fs = append(fs, zap.Int16s(key, v))
		case []int32:
			fs = append(fs, zap.Int32s(key, v))
		case []int64:
			fs = append(fs, zap.Int64s(key, v))
		case []uint:
			fs = append(fs, zap.Uints(key, v))
		// case []uint8:
		// 	fs = append(fs, zap.Uint8s(key, v))
		case []uint16:
			fs = append(fs, zap.Uint16s(key, v))
		case []uint32:
			fs = append(fs, zap.Uint32s(key, v))
		case []uint64:
			fs = append(fs, zap.Uint64s(key, v))
		case []float32:
			fs = append(fs, zap.Float32s(key, v))
		case []float64:
			fs = append(fs, zap.Float64s(key, v))
		case []bool:
			fs = append(fs, zap.Bools(key, v))
		default:
			// 对于其他类型，使用 zap.Any
			fs = append(fs, zap.Any(key, v))
		}
	}
	return fs
}

func (z *ZapLogger) Info(msg string, kv ...interface{}) {
	if z.useSugar {
		z.su.Infow(msg, kv...)
	} else {
		z.lg.Info(msg, fields(kv...)...)
	}
}

func (z *ZapLogger) Infof(template string, args ...interface{}) {
	if z.useSugar {
		z.su.Infof(template, args...)
	}
}

//定義公共的格式
// 2025.02.19 20:44:03:445
// t:2025-02-19 20:44:03:445
// ts=2025-02-19 20:44:03:445
// [2025-02-19 20:44:03:445] {2025-02-19 20:44:03:445}
// [t]2025-02-19_20:44:03:445
// key-style, value-style, formatter-style
// [2006-03-04 15:14:02:936]
