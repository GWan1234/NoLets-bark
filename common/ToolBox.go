package common

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v4"
	"github.com/sunvc/apns2"
)

type NotPushedData struct {
	ID           string          `json:"id"`
	CreateDate   time.Time       `json:"createDate"`
	LastPushDate time.Time       `json:"lastPushDate"`
	Count        int             `json:"count"`
	Params       *ParamsResult   `json:"params"`
	PushType     apns2.EPushType `json:"pushType"`
}

func BaseDir(path ...string) string {
	dataDir := LocalConfig.System.DataDir
	if len(path) == 0 {
		return dataDir
	}
	return filepath.Join(append([]string{dataDir}, path...)...)
}

func Unique[T comparable](list []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(list))

	for _, v := range list {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Contains 判断切片中是否包含指定元素
func Contains[T comparable](slice []T, val T) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func Admin(ctx *gin.Context) bool {
	admin, ok := ctx.Get("admin")
	if ok {
		auth, success := admin.(bool)
		return success && auth
	}
	return false
}

func GetClientHost(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := c.Request.Host
	return fmt.Sprintf("%s://%s", scheme, host)
}

func IsFileInDirectory(dirPath, fileName string) (bool, error) {
	// 对目录路径进行规范化处理
	dirPath = filepath.Clean(dirPath)

	// 检查目录是否存在
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // 目录不存在时，直接返回文件不在目录中
		}
		return false, fmt.Errorf("检查目录状态出错: %w", err)
	}

	// 确认路径指向的是一个目录
	if !dirInfo.IsDir() {
		return false, fmt.Errorf("路径 %q 不是一个目录", dirPath)
	}

	// 构建文件的完整路径
	filePath := filepath.Join(dirPath, fileName)

	// 检查文件是否存在
	_, err = os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // 文件不存在
		}
		return false, fmt.Errorf("检查文件状态出错: %w", err)
	}

	return true, nil
}

func FilterShortStrings(input []string, minNumber, maxNumber int) []string {
	var result []string
	for _, s := range input {
		if len(s) >= minNumber && len(s) <= maxNumber {
			result = append(result, s)
		}
	}
	return result
}

func UserID(name ...string) string {
	return shortuuid.NewWithNamespace(strings.Join(name, ""))
}

func InterfaceSliceToStringSlice(input []interface{}) []string {
	result := make([]string, 0, len(input))
	for _, v := range input {
		if str, ok := v.(string); ok {
			result = append(result, str)
		} else {
			// 如果类型不是 string，可以选择忽略或报错
			// 这里选择忽略非字符串类型
		}
	}
	return result
}

func Decrypt(signText string, key string) (string, error) {

	// Base64 URL Safe -> 标准 Base64
	signText = strings.ReplaceAll(signText, "-", "+")
	signText = strings.ReplaceAll(signText, "_", "/")
	if m := len(signText) % 4; m != 0 {
		signText += strings.Repeat("=", 4-m)
	}

	data, err := base64.StdEncoding.DecodeString(signText)
	if err != nil {
		return "", errors.New("missing signature")
	}

	nonceSize := 12
	tagSize := 16
	if len(data) <= nonceSize+tagSize {
		return "", errors.New("missing signature")
	}
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize : len(data)-tagSize]
	tag := data[len(data)-tagSize:]

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", errors.New("missing signature")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New("missing signature")
	}

	// CryptoKit 是把 tag 单独放在尾部，需要拼接到 ciphertext
	decrypted, err := aesgcm.Open(nil, nonce, append(ciphertext, tag...), nil)
	if err != nil {

		return "", errors.New("missing signature")
	}

	return string(decrypted), nil
}
