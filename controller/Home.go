package controller

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunvc/NoLets/common"
)

// Home 处理首页请求
// 支持两种功能:
// 1. 通过id参数移除未推送数据
// 2. 生成二维码图片
func Home(c *gin.Context) {

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
