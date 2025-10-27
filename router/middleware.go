package router

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunvc/NoLets/common"
	"github.com/sunvc/NoLets/controller"
)

func Verification() gin.HandlerFunc {

	return func(c *gin.Context) {

		// 先查看是否是管理员身份
		authHeader := c.GetHeader("Authorization")
		if common.Contains[string](common.LocalConfig.System.Auths, authHeader) && authHeader != "" {
			c.Set("admin", true)
			return
		}

		localUser := common.LocalConfig.System.User
		localPassword := common.LocalConfig.System.Password
		// 配置了账号密码，进行身份校验
		if localUser != "" && localPassword != "" {
			// 优先使用 Basic Auth
			user, pass, hasAuth := c.Request.BasicAuth()
			if !hasAuth {
				// 如果没有 Basic Auth，则尝试从查询参数中获取
				user = c.Query(common.UserName)
				pass = c.Query(common.Password)

				if c.Request.Method == http.MethodPost {
					if user == "" {
						user = c.PostForm(common.UserName)
					}
					if pass == "" {
						pass = c.PostForm(common.Password)
					}
				}
			}

			if user == localUser && pass == localPassword {
				c.Set("admin", true)
				return
			}

		}

		// 如果没有身份验证信息
		c.Set("admin", false)
		c.Next()
	}
}

// CheckDotParamMiddleware 检查 GET 请求第一个 path 参数是否包含 '.'
func CheckDotParamMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if value := c.Param("deviceKey"); strings.Contains(value, ".") {
			controller.GetImage(c)
			c.Abort()
			return
		}
		// 放行请求
		c.Next()
	}
}

func GCMDecryptMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 512)

		userAgent := c.GetHeader(common.HeaderUserAgent)
		if !strings.HasPrefix(strings.ToLower(userAgent), strings.ToLower(common.APPNAME)) {
			c.AbortWithStatusJSON(http.StatusOK, common.Failed(http.StatusUnauthorized, "SB"))
			return
		}

		key := []byte(common.LocalConfig.System.SignKey)
		if len(key) == 0 {
			c.Next()
			return
		}
		header := c.GetHeader("X-Signature")

		if header == "" {
			c.AbortWithStatusJSON(http.StatusOK, common.Failed(
				http.StatusUnauthorized,
				"missing signature",
			))
			return
		}

		// Base64 URL Safe -> 标准 Base64
		header = strings.ReplaceAll(header, "-", "+")
		header = strings.ReplaceAll(header, "_", "/")
		if m := len(header) % 4; m != 0 {
			header += strings.Repeat("=", 4-m)
		}

		data, err := base64.StdEncoding.DecodeString(header)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, common.Failed(
				http.StatusUnauthorized,
				"missing signature",
			))
			return
		}

		nonceSize := 12
		tagSize := 16
		if len(data) <= nonceSize+tagSize {

			c.AbortWithStatusJSON(http.StatusOK, common.Failed(
				http.StatusUnauthorized,
				"missing signature",
			))
			return
		}

		nonce := data[:nonceSize]
		ciphertext := data[nonceSize : len(data)-tagSize]
		tag := data[len(data)-tagSize:]

		block, err := aes.NewCipher(key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, common.Failed(
				http.StatusUnauthorized,
				"missing signature",
			))
			return
		}

		aesgcm, err := cipher.NewGCM(block)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, common.Failed(
				http.StatusUnauthorized,
				"missing signature",
			))
			return
		}

		// CryptoKit 是把 tag 单独放在尾部，需要拼接到 ciphertext
		decrypted, err := aesgcm.Open(nil, nonce, append(ciphertext, tag...), nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, common.Failed(
				http.StatusUnauthorized,
				"missing signature",
			))
			log.Println("Signature verification failed！err1:", err)
			return
		}

		timestampStr := string(decrypted)

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, common.Failed(
				http.StatusUnauthorized,
				"missing signature",
			))
			log.Println("Signature verification failed！err2:", err)
			return
		}

		now := time.Now().Unix()
		if now-timestamp > 10 || now < timestamp {
			c.AbortWithStatusJSON(http.StatusOK, common.Failed(
				http.StatusUnauthorized,
				"missing signature",
			))
			log.Println("Signature verification failed！timestamp:", timestampStr)
			return
		}

		log.Println("Signature verification successful！")
		// 解密成功，存入 context
		c.Set("decrypted", decrypted)
		c.Next()

	}
}
