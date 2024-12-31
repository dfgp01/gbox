package env

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"math/rand"
	"time"
)

// sha1加密字符串
func GenSha1(str string) (hexSha1str string) {
	sha1Ctx := sha1.New()
	sha1Ctx.Write([]byte(str))
	cipherStr := sha1Ctx.Sum(nil)
	hexSha1str = hex.EncodeToString(cipherStr)
	return
}

// md5加密字符串
func GenMd5(str string) (hexMd5str string) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(str))
	cipherStr := md5Ctx.Sum(nil)
	hexMd5str = hex.EncodeToString(cipherStr)
	return
}

// 根据密码明文生成密文+盐值，供数据库存放
func GenPassword(passWd string) (sha1Passwd string, salt string) {
	ret, _ := userPasswdSeed(passWd)
	salt = fmt.Sprintf("%s%d%d", RanStrings(10), time.Now().UnixNano(), rand.Intn(1000000))
	salt = GenMd5(salt)
	sha1Passwd = GenSha1(fmt.Sprintf("%d%s", ret, salt))
	return
}

// 拿别人的，具体原理我也不懂
func userPasswdSeed(passwd string) (ret uint64, err error) {
	c := crc32.NewIEEE()
	if _, err = c.Write([]byte(passwd)); err != nil {
		return
	}
	ret = uint64(c.Sum32())

	ret <<= 32
	ret >>= 12
	ret >>= 5
	ret <<= 10
	ret ^= 0xFF01CFF4FB0AF00A
	return
}

// 随机token串，目前没什么用
func GenToken(uid int64) (token string) {
	randomStr := RanStrings(10)
	tk := fmt.Sprintf("%d%d%s", uid, time.Now().UnixNano(), randomStr)
	token = GenSha1(tk)
	return
}
