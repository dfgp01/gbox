package env

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// 生成随机16进制字符串，返回的字符串长度是len*2
func RandomHex(len int) string {
	return hex.EncodeToString(RandomBytes(len))
}

// 生成随机bytes数组，len指定长度
func RandomBytes(len int) []byte {
	b := make([]byte, len)
	_, _ = crand.Read(b)
	return b
}

const (
	defaultStr = "0123456789abcdefghijklmnopqrstuvwxyz"
)

// 生成随机字符串（小写），length=指定长度
func RanStrings(length int) string {
	b := make([]byte, length)
	rn := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = defaultStr[rn.Intn(len(defaultStr))]
	}
	return string(b)
}

// camelToSnakeCase 将驼峰式字符串转换为小写下划线形式
func CamelToSnakeCase(camelCase string) string {
	// 先将字符串的首字母转为小写
	snakeCase := strings.ToLower(string(camelCase[0])) + camelCase[1:]

	// 使用正则表达式找到所有大写字母并替换为下划线加小写形式
	re := regexp.MustCompile("([a-z])([A-Z])")
	snakeCase = re.ReplaceAllString(snakeCase, "${1}_${2}")

	return strings.ToLower(snakeCase)
}
