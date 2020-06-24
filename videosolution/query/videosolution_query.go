/*
@Author : yidun_dev
@Date : 2020-06-24
@File : videosolution_query.go
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
	apiUrl    = "http://as.dun.163.com/v1/videosolution/query/task"
	version   = "v1"
	secretId  = "your_secret_id"  //产品密钥ID，产品标识
	secretKey = "your_secret_key" //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
)

//请求易盾接口
func check(params url.Values) *simplejson.Json {
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
	taskIds := []string{"fss8b041517c46b7b2fff5d5110833d5", "df3d5cc6b5474fddb92dfe4d4f1cda34"}
	jsonString, _ := json.Marshal(taskIds)
	params := url.Values{"taskIds": []string{string(jsonString)}}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		if resultArray == nil || len(resultArray) == 0 {
			fmt.Printf("暂时没有结果需要获取, 请稍后重试!")
		} else {
			for _, resultItem := range resultArray {
				if resultMap, ok := resultItem.(map[string]interface{}); ok {
					taskId := resultMap["taskId"].(string)
					status, _ := resultMap["status"].(json.Number).Int64()
					if status == 0 {
						result, _ := resultMap["result"].(json.Number).Int64()
						fmt.Printf("点播音视频, taskId:%s, 检测结果:%d", taskId, result)
						if resultMap["evidences"] != nil {
							evidences, _ := resultMap["evidences"].(map[string]interface{})
							if evidences["texts"] != nil {
								texts := evidences["texts"].([]interface{})
								for _, textItem := range texts {
									if textMap, ok := textItem.(map[string]interface{}); ok {
										dataId, _ := textMap["dataId"].(string)
										action, _ := textMap["action"].(json.Number).Int64()
										fmt.Printf("文本信息, dataId:%s, 检测结果:%d", dataId, action)
									}
								}
							} else if evidences["images"] != nil {
								images := evidences["images"].([]interface{})
								for _, imageItem := range images {
									if imageMap, ok := imageItem.(map[string]interface{}); ok {
										dataId, _ := imageMap["dataId"].(string)
										status, _ := imageMap["status"].(json.Number).Int64()
										action, _ := imageMap["action"].(json.Number).Int64()
										fmt.Printf("图片信息, dataId:%s, 检测状态:%d, 检测结果:%d", dataId, status, action)
									}
								}
							} else if evidences["audios"] != nil {
								audios := evidences["audios"].([]interface{})
								for _, audioItem := range audios {
									if audioMap, ok := audioItem.(map[string]interface{}); ok {
										dataId, _ := audioMap["dataId"].(string)
										status, _ := audioMap["asrStatus"].(json.Number).Int64()
										action, _ := audioMap["action"].(json.Number).Int64()
										fmt.Printf("语音信息, dataId:%s, 检测状态:%d, 检测结果:%d", dataId, status, action)
									}
								}
							} else if evidences["videos"] != nil {
								videos := evidences["videos"].([]interface{})
								for _, videoItem := range videos {
									if videoMap, ok := videoItem.(map[string]interface{}); ok {
										dataId, _ := videoMap["dataId"].(string)
										status, _ := videoMap["status"].(json.Number).Int64()
										level, _ := videoMap["level"].(json.Number).Int64()
										fmt.Printf("视频信息, dataId:%s, 检测状态:%d, 检测结果:%d", dataId, status, level)
									}
								}
							}
						} else if resultMap["reviewEvidences"] != nil {
							reviewEvidences, _ := resultMap["reviewEvidences"].(map[string]interface{})
							_ = reviewEvidences["reason"].(string)
							_ = reviewEvidences["detail"].(map[string]interface{})
							_ = reviewEvidences["text"].([]interface{})
							_ = reviewEvidences["image"].([]interface{})
							_ = reviewEvidences["audio"].([]interface{})
							_ = reviewEvidences["video"].([]interface{})
						}
					} else if status == 20 {
						fmt.Printf("点播音视频, taskId:%s, 数据非7天内", taskId)
					} else if status == 30 {
						fmt.Printf("点播音视频, taskId:%s, 数据不存在", taskId)
					}
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
