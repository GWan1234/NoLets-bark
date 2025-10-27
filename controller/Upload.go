package controller

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sunvc/NoLets/common"
)

// Upload 处理图片上传请求
// 支持GET和POST两种请求方式:
// GET: 返回上传页面
// POST: 上传图片并保存
func Upload(c *gin.Context) {
	// 验证管理员权限
	admin := Verification(c)

	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "upload.html", gin.H{})
		return
	}

	if !admin {
		log.Println("Unauthorized upload attempt")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 获取文件名
	fileName := c.PostForm("filename")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	// 创建上传目录
	uploadDir := "./images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}
	// 验证文件是否在uploads目录下
	if isTrue, _ := common.IsFileInDirectory(fileName, uploadDir); isTrue {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found in uploads directory"})
		return

	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file upload failed: " + err.Error()})
		return
	}

	// 验证文件类型
	imageType := file.Header.Get(common.HeaderContentType)
	if imageType != common.MIMEImagePng && imageType != common.MIMEImageJpeg {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only image files are allowed"})
		return
	}

	// 生成安全的文件名
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = filepath.Ext(file.Filename)
	}
	if ext == "" {
		ext = ".png" // 默认扩展名
	}

	// 使用时间戳和随机字符串生成唯一文件名
	safeFileName := fmt.Sprintf("%s%s",
		strings.TrimSuffix(fileName, ext),
		strings.ToLower(ext))

	// 完整的文件路径
	filePath := filepath.Join(uploadDir, safeFileName)

	// 保存文件
	if err = c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file: " + err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":  "file uploaded successfully",
		"filename": safeFileName,
		"path":     filePath,
		"size":     file.Size,
		"type":     file.Header.Get("Content-Type"),
	})
}
