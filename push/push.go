package push

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sunvc/NoLets/common"
	"github.com/sunvc/apns2"
	"github.com/sunvc/apns2/payload"
)

// Push message to APNs server
func Push(params *common.ParamsMap, pushType apns2.EPushType, token string) error {
	pl := payload.NewPayload().MutableContent()

	if pushType == apns2.PushTypeBackground {
		pl = pl.ContentAvailable()
	} else {
		pl = pl.AlertTitle(common.PMGet(params, common.Title)).
			AlertSubtitle(common.PMGet(params, common.Subtitle)).
			AlertBody(common.PMGet(params, common.Body)).
			Sound(common.PMGet(params, common.Sound)).
			TargetContentID(common.PMGet(params, common.ID)).
			ThreadID(common.PMGet(params, common.Group)).
			Category(common.PMGet(params, common.Category))
	}

	// 添加自定义参数
	skipKeys := map[string]struct{}{
		common.DeviceKey:   {},
		common.DeviceKeys:  {},
		common.DeviceToken: {},
		common.Title:       {},
		common.Body:        {},
		common.Sound:       {},
		common.Category:    {},
	}

	for pair := params.Oldest(); pair != nil; pair = pair.Next() {
		if _, skip := skipKeys[pair.Key]; skip {
			continue
		}
		pl.Custom(pair.Key, pair.Value)
	}

	CLI := <-CLIENTS // 从池中获取一个客户端
	CLIENTS <- CLI   // 将客户端放回池中

	// 创建并发送通知
	resp, err := CLI.Push(&apns2.Notification{
		DeviceToken: token,
		CollapseID:  fmt.Sprint(params.Value(common.ID)),
		Topic:       common.LocalConfig.Apple.Topic,
		Payload:     pl,
		Expiration:  common.DateNow().Add(24 * time.Hour),
		PushType:    pushType,
	})

	// 错误处理
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("APNs push failed: %s", resp.Reason)
	}
	return nil

}

func BatchPush(params *common.ParamsResult, pushType apns2.EPushType) error {

	var (
		errors []error
		mu     sync.Mutex
		wg     sync.WaitGroup
	)

	for _, token := range params.Tokens {
		// 分页推送
		//if len(params.Results) > 0 {
		//	for _, param := range params.Results {
		//		wg.Add(1)
		//		go func(p *common.ParamsMap) {
		//			defer wg.Done()
		//			if err := Push(p, pushType, token); err != nil {
		//				log.Println(err.Error())
		//				mu.Lock()
		//				errors = append(errors, err)
		//				mu.Unlock()
		//			}
		//		}(param)
		//	}
		//} else {
		wg.Add(1)
		go func(p *common.ParamsMap) {
			defer wg.Done()
			if err := Push(params.Params, pushType, token); err != nil {
				log.Println(err.Error())
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(params.Params)
		//}
	}

	wg.Wait()
	if len(errors) > 0 {
		return fmt.Errorf("APNs push failed: %v", errors)
	}

	return nil
}
