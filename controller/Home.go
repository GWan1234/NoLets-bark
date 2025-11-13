package controller

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

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
		if !common.LocalConfig.System.ProxyDownload {
			c.String(http.StatusBadRequest, "missing")
			return
		}
		DownloadProject(c)
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
		"ICP":           common.LocalConfig.System.ICPInfo,
		"URL":           template.URL(url),
		"LOGORAW":       template.HTML(common.LOGORAW),
		"BACKGROUNDSVG": template.URL(common.LogoSvgImage("ff00000f", false)),
		"DOCS":          "https://wiki.wzs.app",
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
	client := &http.Client{}

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

func DownloadProject(c *gin.Context) {
	goos := strings.ToLower(c.GetHeader("os"))
	var name string
	if strings.Contains(goos, "win") {
		name = "windows_amd64"
	} else if strings.Contains(goos, "mac") || strings.Contains(goos, "darwin") {
		if strings.Contains(goos, "arm") {
			name = "darwin_arm64"
		} else {
			name = "darwin_amd64"
		}
	} else {
		if strings.Contains(goos, "arm") {
			name = "linux_arm64"
		} else {
			name = "linux_amd64"
		}
	}

	// 发起请求获取远程文件
	// 2. 创建带超时的HTTP客户端
	client := &http.Client{}

	url := fmt.Sprintf("https://github.com/sunvc/NoLets/releases/download/%s/NoLets_%s.tar.gz", common.LocalConfig.System.Version, name)
	log.Println(url)
	resp, err := client.Get(url)

	defer resp.Body.Close()

	if err != nil {
		c.String(http.StatusBadGateway, "fetch error: %v", err)
		return
	}

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
