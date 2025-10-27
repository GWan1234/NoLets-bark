package common

import (
	"fmt"
	"time"
)

type BaseResp struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// for the fast return Success result

func Success(data ...interface{}) BaseResp {
	var result interface{}

	if len(data) > 0 {
		result = data[0]
	}
	return BaseResp{
		Code:      200,
		Message:   "success",
		Data:      result,
		Timestamp: DateNow().Unix(),
	}
}

// for the fast return Failed result

func Failed(code int, message string, args ...interface{}) BaseResp {
	return BaseResp{
		Code:      code,
		Message:   fmt.Sprintf(message, args...),
		Timestamp: DateNow().Unix(),
	}
}

// for the fast return result with custom data

func BaseRes(code int, message string, data ...interface{}) BaseResp {
	var result interface{}

	if len(data) > 0 {
		result = data[0]
	}

	return BaseResp{
		Code:      code,
		Message:   message,
		Data:      result,
		Timestamp: DateNow().Unix(),
	}
}

type DeviceInfo struct {
	Key   string `json:"key"`
	Token string `json:"token"`
	Group string `json:"group,omitempty"`
}

func DateNow() time.Time {
	return time.Now().UTC()
}
