/*
@Author : yidun_dev
@Date : 2020-01-20
@File : video_query.go
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
	apiUrl     = "http://as.dun.163.com/v1/video/query/task"
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
	taskIds := []string{"c679d93d4a8d411cbe3454214d4b1fd7", "49800dc7877f4b2a9d2e1dec92b988b6"}
	jsonString, _ := json.Marshal(taskIds)
	params := url.Values{"taskIds": []string{string(jsonString)}}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		for _, result := range resultArray {
			if resultMap, ok := result.(map[string]interface{}); ok {
				status, _ := resultMap["status"].(json.Number).Int64()
				if status != 0 {
					fmt.Printf("获取结果异常, status: %d", status)
					continue
				}
				taskId, _ := resultMap["taskId"].(string)
				callback, _ := resultMap["callback"].(string)
				videoLevel, _ := resultMap["level"].(json.Number).Int64()
				if videoLevel == 0 {
					fmt.Printf("正常, callback: %s", callback)
				} else if videoLevel == 1 || videoLevel == 2 {
					evidenceArray := resultMap["evidences"].([]interface{})
					for _, evidence := range evidenceArray {
						if evidenceMap, ok := evidence.(map[string]interface{}); ok {
							_, _ = evidenceMap["beginTime"].(json.Number).Int64()
							_, _ = evidenceMap["endTime"].(json.Number).Int64()
							_, _ = evidenceMap["type"].(json.Number).Int64()
							_, _ = evidenceMap["url"].(string)
							labelArray, _ := evidenceMap["labels"].([]interface{})
							for _, labelItem := range labelArray {
								if labelItemMap, ok := labelItem.(map[string]interface{}); ok {
									_, _ = labelItemMap["label"].(json.Number).Int64()
									_, _ = labelItemMap["level"].(json.Number).Int64()
									_, _ = labelItemMap["rate"].(json.Number).Float64()
								}
							}
							fmt.Printf("taskId: %s, %d, callback: %s, 证据信息: %s, 证据分类: %s", taskId, videoLevel, callback, evidence, labelArray)
						}
					}
				}
			}
		}

	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
