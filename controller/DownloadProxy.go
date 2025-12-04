package controller

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func StreamDownload(c *gin.Context) {
	filePath := "path/to/large/file.zip"
	file, err := os.Open(filePath)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}
	defer file.Close()

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=download.zip")

	buffer := make([]byte, 1024*1024) // 1MB缓冲区
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			c.Writer.Write(buffer[:n])
			c.Writer.Flush()
		}
		if err == io.EOF {
			break
		}
	}
}
