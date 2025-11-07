package controller

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunvc/NoLets/common"
)

// Home 处理首页请求
// 支持两种功能:
// 1. 通过id参数移除未推送数据
// 2. 生成二维码图片
func Home(c *gin.Context) {

	ua := strings.ToLower(c.GetHeader("User-Agent"))
	if strings.Contains(ua, "curl") || strings.Contains(ua, "wget") {
		c.String(http.StatusOK, "Hello World")
		return
	}

	if data := c.GetHeader("X-DATA"); len(data) > 10 {
		ProxyDownload(c, data)
		return
	}

	if id := c.Query("id"); id != "" {
		NotPushedDataList.Delete(id)
		c.Status(http.StatusOK)
		return
	}

	url := common.GetClientHost(c)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"ICP":     common.LocalConfig.System.ICPInfo,
		"URL":     template.URL(url),
		"LOGORAW": template.HTML(common.LOGORAW),
		"LOGOSVG": template.URL(common.LogoSvgImage("ff00000f", false)),
		"DOCS":    "https://wiki.wzs.app",
	})
}

func ProxyDownload(c *gin.Context, data string) {
	if !common.LocalConfig.System.ProxyDownload {
		c.String(http.StatusBadRequest, "missing")
		return
	}
	// 获取用户请求的下载地址

	targetURL, err := common.Decrypt(data, common.LocalConfig.System.SignKey)

	if targetURL == "" || err != nil {
		c.String(http.StatusBadRequest, "missing X-DATA header")
		return
	}

	// 发起请求获取远程文件
	// 2. 创建带超时的HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 3. 使用Head请求先获取文件信息（可选）
	headResp, err := client.Head(targetURL)
	if err == nil {
		// 检查文件大小
		if contentLength := headResp.Header.Get("Content-Length"); contentLength != "" {
			if size, _ := strconv.ParseInt(contentLength, 10, 64); size > 20*1024*1024 { // 限制100MB
				c.String(http.StatusBadRequest, "file too large")
				return
			}
		}
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		c.String(http.StatusBadGateway, "fetch error: %v", err)
		return
	}
	defer resp.Body.Close()

	// 设置返回头，保持文件名或类型
	for k, v := range resp.Header {
		if len(v) > 0 {
			c.Writer.Header().Set(k, v[0])
		}
	}

	// 直接流式复制响应体给用户
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)

}
