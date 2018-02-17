package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
)

var (
	Log       *logrus.Logger
	Viper     *viper.Viper
	KitLogger log.Logger
)

func init() {
	Viper = viper.New()

	Log = logrus.New()
	// 以json格式显示.
	// logrus.SetFormatter(&logrus.JSONFormatter{})
	// Log.Formatter = &logrus.JSONFormatter{}
	viper.SetDefault("ContentDir", "content")

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// logrus.SetOutput(os.Stdout)
	Log.Out = os.Stdout

	// file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY, 0666)
	// if err == nil {
	// 	Log.Out = file
	// } else {
	// 	Log.Info("日志文件打开失败，使用默认输出")
	// }

	// Only log the warning severity or above.
	// logrus.SetLevel(logrus.WarnLevel)
	Log.Level = logrus.ErrorLevel

	//gokitLog

}

//判断一个数是否是2的N次方
func Is2N(n int) bool {
	return n > 0 && ((n & (n - 1)) == 0)
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func GetGuid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

//判断文件或文件夹是否存在 error为nil 则存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
