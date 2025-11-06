package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunvc/NoLets/common"
	"github.com/sunvc/NoLets/database"
)

// GetDeviceToken 获取设备的推送token
// 通过deviceKey查询对应的推送token
func GetDeviceToken(c *gin.Context) {
	deviceKey := c.Param("deviceKey")
	fmt.Println(deviceKey)
	token, err := database.DB.DeviceTokenByKey(deviceKey)
	if err != nil {
		c.JSON(http.StatusOK, common.Failed(http.StatusInternalServerError, "failed to get device token: %v", err))
		return
	}
	c.JSON(http.StatusOK, common.Success(token))
}
