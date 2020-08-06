/*
@Author : yidun_dev
@Date : 2020-01-20
@File : videosolution_callback.go
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
	apiUrl    = "http://as.dun.163.com/v1/videosolution/callback/results"
	version   = "v1.1"            //点播音视频解决方案版本v1.1及以上语音二级细分类subLabels结构进行调整
	secretId  = "your_secret_id"  //产品密钥ID，产品标识
	secretKey = "your_secret_key" //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
)

//请求易盾接口
func check() *simplejson.Json {
	params := url.Values{}
	params["secretId"] = []string{secretId}
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
	ret := check()

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		if len(resultArray) == 0 {
			fmt.Printf("暂时没有结果需要获取, 请稍后重试!")
		} else {
			for _, result := range resultArray {
				if resultMap, ok := result.(map[string]interface{}); ok {
					taskId := resultMap["taskId"].(string)
					result, _ := resultMap["result"].(json.Number).Int64()
					fmt.Printf("taskId: %s, result: %d", taskId, result)
					if resultMap["evidences"] != nil {
						evidences, _ := resultMap["evidences"].(map[string]interface{})
						if evidences["audio"] != nil {
							audio, _ := evidences["audio"].(map[string]interface{})
							asrStatus, _ := audio["asrStatus"].(json.Number).Int64()
							if asrStatus == 4 {
								asrResult, _ := audio["asrResult"].(json.Number).Int64()
								fmt.Printf("检测失败: taskId=%s, asrResult=%d", taskId, asrResult)
							} else {
								action, _ := audio["action"].(json.Number).Int64()
								labelArray, _ := audio["labels"].([]interface{})
								if action == 0 {
									fmt.Printf("taskId=%s，结果：通过", taskId)
								} else if action == 1 || action == 2 {
									for _, labelItem := range labelArray {
										if labelItemMap, ok := labelItem.(map[string]interface{}); ok {
											_, _ = labelItemMap["label"].(json.Number).Int64()
											_, _ = labelItemMap["level"].(json.Number).Int64()
											// 注意二级细分类结构
											subLabels := labelItemMap["subLabels"].([]interface{})
											if subLabels != nil && len(subLabels) > 0 {
												for _, subLabelItem := range subLabels {
													if subLabelMap, ok := subLabelItem.(map[string]interface{}); ok {
														_, _ = subLabelMap["subLabel"].(string)
														details := subLabelMap["details"].(map[string]interface{})
														_, _ = details["hint"].([]interface{})
													}
												}
											}
										}
									}
									var printString string
									if action == 1 {
										printString = "不确定"
									} else {
										printString = "不通过"
									}
									fmt.Printf("taskId=%s, 结果: %s，证据信息如下: %s", taskId, printString, labelArray)
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
