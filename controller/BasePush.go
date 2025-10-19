package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunvc/NoLets/common"
	"github.com/sunvc/NoLets/database"
	"github.com/sunvc/NoLets/push"
	"github.com/sunvc/apns2"
)

// BasePush 处理基础推送请求
// 验证推送参数并执行推送操作
func BasePush(c *gin.Context) {

	result := common.NewParamsResult(c)

	if result == nil {
		c.JSON(http.StatusOK, common.Failed(http.StatusBadRequest, "Not Params"))
		return
	}

	if len(result.Tokens) <= 0 {
		for _, key := range result.Keys {
			if len(key) > 5 {
				if token, err := database.DB.DeviceTokenByKey(key); err == nil {

					result.Tokens = append(result.Tokens, token)
				}

			}
		}
	}

	if len(result.Tokens) <= 0 {
		c.JSON(http.StatusOK, common.Failed(http.StatusBadRequest, "Failed to get device token"))
		return
	}

	pushType := func() apns2.EPushType {
		// 如果 title, subtitle 和 body 都为空，设置静默推送模式
		if result.PushType == 0 {
			return apns2.PushTypeBackground
		}
		return apns2.PushTypeAlert
	}()

	if err := push.BatchPush(result, pushType); err != nil {
		c.JSON(http.StatusOK, common.Failed(http.StatusInternalServerError, "push failed: %v", err))
		return
	}

	// 如果是管理员，加入到未推送列表
	if id, ok := result.Get(common.ID).(string); common.Admin(c) && ok && len(id) > 0 {
		UpdateNotPushedData(id, result, pushType)
	}

	c.JSON(http.StatusOK, common.Success())
}
