/*
@Author : yidun_dev
@Date : 2020-01-20
@File : audio_query.go
@Version : 1.0
@Golang : 1.13.5
@Doc : http://dun.163.com/api.html
*/
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	apiUrl     = "http://as.dun.163yun.com/v1/audio/submit/task"
	version    = "v1"
	secretId   = "your_secret_id"   //产品密钥ID，产品标识
	secretKey  = "your_secret_key"  //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
	businessId = "your_business_id" //业务ID，易盾根据产品业务特点分配
)

//请求易盾接口
func check(params url.Values) *simplejson.Json {
	params["secretId"] = []string{secretId}
	params["businessId"] = []string{businessId}
	params["version"] = []string{version}
	params["timestamp"] = []string{strconv.FormatInt(time.Now().UnixNano()/1000000, 10)}
	params["nonce"] = []string{strconv.FormatInt(rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(10000000000), 10)}
	params["signature"] = []string{genSignature(params)}

	resp, err := http.Post(apiUrl, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))

	if err != nil {
		fmt.Println("调用API接口失败:", err)
		return nil
	}

	defer resp.Body.Close()

	contents, _ := ioutil.ReadAll(resp.Body)
	result, _ := simplejson.NewJson(contents)
	return result
}

//生成签名信息
func genSignature(params url.Values) string {
	var paramStr string
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		paramStr += key + params[key][0]
	}
	paramStr += secretKey
	md5Reader := md5.New()
	md5Reader.Write([]byte(paramStr))
	return hex.EncodeToString(md5Reader.Sum(nil))
}

func main() {
	taskIds := []string{"202b1d65f5854cecadcb24382b681c1a", "0f0345933b05489c9b60635b0c8cc721"}
	params := url.Values{"taskIds": taskIds}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		antispamArray, _ := ret.Get("antispam").Array()
		if antispamArray == nil || len(antispamArray) == 0 {
			fmt.Printf("暂无审核回调数据")
		} else {
			for _, result := range antispamArray {
				if resultMap, ok := result.(map[string]interface{}); ok {
					status, _ := resultMap["status"].(json.Number).Int64()
					taskId := resultMap["taskId"].(string)
					if status == 30 {
						fmt.Printf("antispam callback taskId=%s，结果：数据不存在", taskId)
					} else {
						action, _ := resultMap["action"].(json.Number).Int64()
						labelArray, _ := resultMap["labels"].([]interface{})
						if action == 0 {
							fmt.Printf("callback taskId=%s，结果：通过", taskId)
						} else if action == 2 {
							for _, labelItem := range labelArray {
								if labelItemMap, ok := labelItem.(map[string]interface{}); ok {
									_, _ = labelItemMap["label"].(json.Number).Int64()
									_, _ = labelItemMap["level"].(json.Number).Int64()
									details := labelItemMap["details"].(map[string]interface{})
									_ = details["hint"].([]interface{})
									_ = labelItemMap["subLabels"].([]interface{})
									fmt.Printf("callback=%s，结果：不通过，分类信息如下：%s", taskId, labelArray)
								}
							}
						}
					}
				}
			}
		}
		languageArray, _ := ret.Get("language").Array()
		if languageArray == nil || len(languageArray) == 0 {
			fmt.Printf("暂无语种检测数据")
		} else {
			for _, result := range languageArray {
				if resultMap, ok := result.(map[string]interface{}); ok {
					status, _ := resultMap["status"].(json.Number).Int64()
					taskId := resultMap["taskId"].(string)
					if status == 30 {
						fmt.Printf("language callback taskId=%s，结果：数据不存在", taskId)
					} else {
						detailsArray, _ := resultMap["details"].([]interface{})
						if detailsArray != nil && len(detailsArray) > 0 {
							for _, language := range detailsArray {
								if languageMap, ok := language.(map[string]interface{}); ok {
									typeLan, _ := languageMap["type"].(string)
									segmentsArray, _ := languageMap["segments"].([]interface{})
									if segmentsArray != nil && len(segmentsArray) > 0 {
										for _, segment := range segmentsArray {
											if segmentMap, ok := segment.(map[string]interface{}); ok {
												startTime, _ := segmentMap["startTime"].(json.Number).Int64()
												endTime, _ := segmentMap["endTime"].(json.Number).Int64()
												fmt.Printf("taskId=%s，语种类型=%s，开始时间=%d秒，结束时间=%d秒", taskId, typeLan, startTime, endTime)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		asrArray, _ := ret.Get("asr").Array()
		if asrArray == nil || len(asrArray) == 0 {
			fmt.Printf("暂无语音翻译数据")
		} else {
			for _, result := range asrArray {
				if resultMap, ok := result.(map[string]interface{}); ok {
					status, _ := resultMap["status"].(json.Number).Int64()
					taskId := resultMap["taskId"].(string)
					if status == 30 {
						fmt.Printf("asr callback taskId=%s，结果：数据不存在", taskId)
					} else {
						detailsArray, _ := resultMap["details"].([]interface{})
						if detailsArray != nil && len(detailsArray) > 0 {
							for _, asr := range detailsArray {
								if asrMap, ok := asr.(map[string]interface{}); ok {
									startTime, _ := asrMap["startTime"].(json.Number).Int64()
									endTime, _ := asrMap["endTime"].(json.Number).Int64()
									content, _ := asrMap["content"].(string)
									fmt.Printf("taskId=%s，文字翻译结果=%s，开始时间=%d秒，结束时间=%d秒", taskId, content, startTime, endTime)
								}
							}
						}
					}
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
