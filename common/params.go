package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wk8/go-ordered-map/v2"
)

const MaxBytes = 4096

type ParamsMap = orderedmap.OrderedMap[string, interface{}]

// ParamsResult 结构体用于存储和管理请求参数
// 使用有序映射存储参数，保证参数的处理顺序
type ParamsResult struct {
	Params *ParamsMap
	//Results  []*ParamsMap
	Tokens   []string
	Keys     []string
	PushType int
}

// NewParamsResult 创建新的参数结果对象
// 参数:
//   - c: gin上下文对象，用于获取请求参数
//
// 返回:
//   - *ParamsResult: 初始化后的参数结果对象
func NewParamsResult(c *gin.Context) *ParamsResult {
	main := &ParamsResult{
		Params: orderedmap.New[string, interface{}](),
		//Results: []*ParamsMap{},
		Keys:   []string{},
		Tokens: []string{},
	}
	main.HandlerParamsToMapOrder(c)
	main.PushType = ParamsNanAndDefault(main)

	if main.PushType == -1 {
		return nil
	}

	// MARK: 分页推送支持
	//results, err := SplitPayloadIfExceedsLimit(main.Params)
	//if err == nil {
	//	main.Results = results
	//}

	var resultKeys []string

	if keys, ok := main.Params.Get(DeviceKeys); ok {
		if vals, oka := keys.([]interface{}); oka {
			resultKeys = InterfaceSliceToStringSlice(vals)
		}
	}

	if key, ok := main.Params.Get(DeviceKey); ok {
		if val, oka := key.(string); oka {
			resultKeys = append(resultKeys, val)
		}

	}

	resultKeys = FilterShortStrings(resultKeys, 5, 64)
	main.Keys = Unique[string](resultKeys)

	if len(main.Keys) > LocalConfig.System.MaxDeviceKeyArrLength {
		main.Keys = main.Keys[:LocalConfig.System.MaxDeviceKeyArrLength]
	}

	var tokens []string
	if token, ok := main.Params.Get(DeviceToken); ok {
		if val, oka := token.(string); oka && len(val) > 10 {
			tokens = append(tokens, val)
		}
	}

	tokens = FilterShortStrings(tokens, 60, 65)

	main.Tokens = tokens

	return main
}

// NormalizeKey 规范化参数键名
// 主要功能:
// 1. 去除首尾空格
// 2. 移除连字符
// 3. 转换为小写
// 参数:
//   - s: 需要规范化的键名字符串
//
// 返回:
//   - string: 规范化后的键名
func (p *ParamsResult) NormalizeKey(s string) string {
	result := strings.TrimSpace(s)
	result = strings.ReplaceAll(result, "-", "")
	result = strings.ReplaceAll(result, "_", "")
	return strings.ToLower(result)
}

// Get 获取参数值
// 参数:
//   - key: 参数键名
//
// 返回:
//   - interface{}: 参数值，如果不存在则返回空字符串
func (p *ParamsResult) Get(key string) interface{} {
	if value, ok := p.Params.Get(key); ok {
		return value
	}
	return ""
}

func PMGet(params *ParamsMap, key string) string {
	if value, ok := params.Get(key); ok {
		return fmt.Sprint(value)
	}
	return ""
}

// HandlerParamsToMapOrder 处理请求参数并转换为有序映射
// 主要功能：
// 1. 从URL路径参数中提取设备密钥、标题、副标题和内容
// 2. 从URL查询参数中获取额外参数
// 3. 处理POST请求的表单数据和JSON数据
// 4. 对参数进行便捷处理
// 5. 将处理后的参数保存到有序映射中
func (p *ParamsResult) HandlerParamsToMapOrder(c *gin.Context) {
	result := orderedmap.New[string, interface{}]()

	// 判断是否是管理员
	host := GetClientHost(c)
	if Admin(c) {
		result.Set(Host, host)
	}
	// 兼容旧版本
	result.Set(Callback, host)

	getDeviceKey := func(value string) {
		deviceKeys := strings.Split(value, ",")
		if len(deviceKeys) > 1 {
			result.Set(DeviceKeys, deviceKeys)
		} else {
			result.Set(DeviceKey, value)
		}
	}

	switch len(c.Params) {
	case 1:
		getDeviceKey(c.Params[0].Value)
	case 2:
		getDeviceKey(c.Params[0].Value)
		result.Set(Body, c.Params[1].Value)
	case 3:
		getDeviceKey(c.Params[0].Value)
		result.Set(Title, c.Params[1].Value)
		result.Set(Body, c.Params[2].Value)
	case 4:
		getDeviceKey(c.Params[0].Value)
		result.Set(Title, c.Params[1].Value)
		result.Set(Subtitle, c.Params[2].Value)
		result.Set(Body, c.Params[3].Value)
	}

	// parse query args (medium priority)
	{
		var keys []string
		var params = c.Request.URL.Query()
		for key, values := range params {
			lowKey := p.NormalizeKey(key)
			if len(values) > 0 {
				if lowKey == DeviceKey {
					keys = append(keys, values...)
				} else {
					result.Set(lowKey, values[0])
				}
			}

		}

		if keysNum := len(keys); keysNum > 0 {
			if keysNum == 1 {
				result.Set(DeviceKey, keys[0])
			} else {
				result.Set(DeviceKeys, keys)
			}
		}
	}

	// POST Body
	if c.Request.Method == http.MethodPost {

		contentType := c.Request.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, MIMEApplicationJSON) {
			var jsonData map[string]interface{}
			err := c.ShouldBindBodyWithJSON(&jsonData)
			if err == nil {
				for k, v := range jsonData {
					result.Set(p.NormalizeKey(k), v)
				}
			}
		} else {
			err := c.Request.ParseForm()
			if err == nil {
				for k, v := range c.Request.PostForm {
					result.Set(p.NormalizeKey(k), v)
				}
			}
		}
	}

	convenientProcessor(result)

	// 写入 ParamsResult.Params
	for pair := result.Oldest(); pair != nil; pair = pair.Next() {
		p.Params.Set(p.NormalizeKey(pair.Key), pair.Value)
	}
}

// convenientProcessor 处理推送参数的便捷转换
// 主要功能：
// 1. 将 data/content/message/text 字段统一转换为 body
// 2. 处理 markdown 相关字段，设置对应的 category
// 3. 规范化 category 字段的值
// 4. 处理声音文件后缀
func convenientProcessor(params *ParamsMap) {
	// 如果没有 body 字段，尝试从其他字段转换
	if _, ok := params.Get(Body); !ok {
		if data, dataOk := params.Get(Data); dataOk {
			params.Set(Body, data)
			params.Delete(Data)
		} else if content, contentOk := params.Get(Content); contentOk {
			params.Set(Body, content)
			params.Delete(Content)
		} else if message, messageOk := params.Get(Message); messageOk {
			params.Set(Body, message)
			params.Delete(Message)
		} else if text, textOk := params.Get(Text); textOk {
			params.Set(Body, text)
			params.Delete(Text)
		}
	}

	// 处理 markdown 字段
	// 如果存在 markdown 字段，将其转换为 body 并设置 category 为 markdown
	if v, ok := params.Get(Markdown); ok {
		params.Set(Body, v)
		params.Set(Category, CategoryMarkdown)
		params.Delete(Markdown)

	}
	// 如果存在 md 字段，将其转换为 body 并设置 category 为 markdown
	if v, ok := params.Get(MD); ok {
		params.Set(Body, v)
		params.Set(Category, CategoryMarkdown)
		params.Delete(MD)
	}

	// 规范化 category 字段
	// 如果 category 不是默认值或 markdown，则设置为默认值
	if v, ok := params.Get(Category); ok {
		if v != CategoryDefault && v != CategoryMarkdown {
			params.Set(Category, CategoryDefault)
		}
	}

	// 处理声音文件后缀
	// 如果声音文件没有 .caf 后缀，则添加后缀
	if val, ok := params.Get(Sound); ok {
		if sound, oka := val.(string); oka {
			if !strings.HasSuffix(sound, ".caf") {
				params.Set(Sound, sound+".caf")
			}
		}
	}
}

func ParamsNanAndDefault(paramsResult *ParamsResult) (resultType int) {
	isEmpty := func(v interface{}) bool {
		s, ok := v.(string)
		if !ok {
			return true
		}
		return len(strings.TrimSpace(s)) == 0
	}

	title, titleOk := paramsResult.Params.Get(Title)
	subTitle, subTitleOk := paramsResult.Params.Get(Subtitle)
	body, bodyOk := paramsResult.Params.Get(Body)
	cipherText, cipherTextOk := paramsResult.Params.Get(CipherText)
	id, idOK := paramsResult.Params.Get(ID)

	titleNan := !titleOk || isEmpty(title)
	subTitleNan := !subTitleOk || isEmpty(subTitle)
	bodyNan := !bodyOk || isEmpty(body)
	cipherNan := !cipherTextOk || isEmpty(cipherText)
	idNan := !idOK || isEmpty(id)

	if titleNan && subTitleNan && bodyNan && cipherNan && idNan {
		resultType = -1
		return
	}

	if (titleNan && subTitleNan && bodyNan && cipherNan) && !idNan {
		resultType = 0
	}

	if !(titleNan && subTitleNan && bodyNan && cipherNan) {
		resultType = 1
	}

	if !cipherNan {
		paramsResult.Params.Set(Body, "-")
	}

	// setDefault 内部函数用于设置默认值
	// params: 参数映射
	// key: 需要设置默认值的键
	// other: 设置默认值的回调函数
	setDefault := func(params *ParamsMap, key string, other func(string, interface{})) {
		newKey := paramsResult.NormalizeKey(key)
		if value, ok := params.Get(newKey); !ok || value == nil || len(fmt.Sprint(value)) == 0 {
			other(newKey, value)
		}
	}

	// 处理默认值
	// 设置自动复制功能的默认值
	setDefault(paramsResult.Params, AutoCopy, func(key string, value interface{}) {
		paramsResult.Params.Set(key, AutoCopyDefault)
	})

	// 设置消息级别的默认值
	setDefault(paramsResult.Params, Level, func(key string, value interface{}) {
		paramsResult.Params.Set(key, LevelDefault)
	})

	// 设置消息分类的默认值
	setDefault(paramsResult.Params, Category, func(key string, value interface{}) {
		paramsResult.Params.Set(key, CategoryDefault)
	})

	// 设置消息ID的默认值（使用UUID）
	setDefault(paramsResult.Params, ID, func(key string, value interface{}) {
		messageID, _ := uuid.NewUUID()
		paramsResult.Params.Set(key, messageID.String())
	})
	return
}
