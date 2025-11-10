package common

import (
	"encoding/json"
	"errors"
	"unicode/utf8"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func SplitPayloadIfExceedsLimit(basePayload *ParamsMap) ([]*ParamsMap, error) {
	// 取出原始 body
	rawBody, ok := basePayload.Get(Body)
	if !ok {
		return nil, errors.New("no body field")
	}
	bodyStr, ok := rawBody.(string)
	if !ok {
		return nil, errors.New("body is not string")
	}

	// 复制 payload 并去掉 body 字段，计算剩余占用字节
	base := copyPayload(basePayload)
	base.Delete(Body)

	baseJson, _ := json.Marshal(orderedToMap(base))
	baseSize := len(baseJson)

	// 如果总大小不超 4096，直接返回原始 payload
	if baseSize+len([]byte(bodyStr)) <= MaxBytes {
		return []*ParamsMap{}, nil
	}

	// 需要分片
	remaining := MaxBytes - baseSize - 100 // 留余：考虑 index/count/body key等
	if remaining <= 0 {
		return nil, errors.New("base payload too large without body")
	}

	// 分片 body
	chunks := splitByUTF8Bytes(bodyStr, remaining)
	count := len(chunks)
	var results []*ParamsMap

	for i, part := range chunks {
		p := copyPayload(base)
		p.Set(Body, part)
		p.Set(CurrentIndex, i)
		p.Set(TotalCount, count)
		results = append(results, p)
	}

	return results, nil
}

func copyPayload(orig *ParamsMap) *ParamsMap {
	newMap := orderedmap.New[string, interface{}]()
	for el := orig.Oldest(); el != nil; el = el.Next() {
		newMap.Set(el.Key, el.Value)
	}
	return newMap
}

func orderedToMap(o *ParamsMap) map[string]interface{} {
	out := make(map[string]interface{})
	for el := o.Oldest(); el != nil; el = el.Next() {
		out[el.Key] = el.Value
	}
	return out
}

func splitByUTF8Bytes(s string, maxBytes int) []string {
	var result []string
	start := 0
	current := 0
	totalBytes := 0

	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			size = 1 // invalid rune fallback
		}

		if totalBytes+size > maxBytes {
			result = append(result, s[start:current])
			start = current
			totalBytes = 0
		}

		totalBytes += size
		current += size
		i += size
	}

	if start < len(s) {
		result = append(result, s[start:])
	}

	return result
}
