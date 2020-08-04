/*
@Author : yidun_dev
@Date : 2020-01-20
@File : liveaudio_callback.go
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
	apiUrl     = "http://as-liveaudio.dun.163.com/v2/liveaudio/callback/results"
	version    = "v2"
	secretId   = "your_secret_id"   //产品密钥ID，产品标识
	secretKey  = "your_secret_key"  //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
	businessId = "your_business_id" //业务ID，易盾根据产品业务特点分配
)

//请求易盾接口
func check() *simplejson.Json {
	params := url.Values{}
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

// 机审信息
func parseMachine(evidences map[string]interface{}, taskId string) {
	fmt.Printf("=== 机审信息 ===")
	asrStatus, _ := evidences["asrStatus"].(json.Number).Int64()
	startTime, _ := evidences["startTime"].(json.Number).Int64()
	endTime, _ := evidences["endTime"].(json.Number).Int64()
	if asrStatus == 4 {
		asrResult, _ := evidences["asrResult"].(json.Number).Int64()
		fmt.Printf("检测失败: taskId=%s, asrResult=%d", taskId, asrResult)
	} else {
		action, _ := evidences["action"].(json.Number).Int64()
		segmentArray, _ := evidences["segments"].([]interface{})
		if action == 0 {
			fmt.Printf("taskId=%s，结果：通过，时间区间【%d-%d】，证据信息如下：%s", taskId, startTime, endTime, segmentArray)
		} else if action == 1 || action == 2 {
			for _, segment := range segmentArray {
				if segmentMap, ok := segment.(map[string]interface{}); ok {
					_, _ = segmentMap["label"].(json.Number).Int64()
					_, _ = segmentMap["level"].(json.Number).Int64()
					_, _ = segmentMap["evidence"].(string)
					var printString string
					if action == 1 {
						printString = "不确定"
					} else {
						printString = "不通过"
					}
					fmt.Printf("taskId=%s，结果：%s，时间区间【%d-%d】，证据信息如下：%s", taskId, printString, startTime, endTime, segmentArray)
				}
			}
		}
	}
	fmt.Printf("================")
}

// 人审信息
func parseHuman(reviewEvidences map[string]interface{}, taskId string) {
	fmt.Printf("=== 人审信息 ===")
	action, _ := reviewEvidences["action"].(json.Number).Int64()
	_, _ = reviewEvidences["actionTime"].(json.Number).Int64()
	_, _ = reviewEvidences["spamType"].(json.Number).Int64()
	spamDetail, _ := reviewEvidences["spamDetail"].(string)
	warnCount, _ := reviewEvidences["warnCount"].(json.Number).Int64()
	prompCount, _ := reviewEvidences["prompCount"].(json.Number).Int64()
	segments, _ := reviewEvidences["segments"].([]interface{})
	status, _ := reviewEvidences["status"].(json.Number).Int64()
	statusStr := "未知"

	if status == 2 {
		statusStr = "检测中"
	} else if status == 3 {
		statusStr = "检测完成"
	}

	if action == 2 {
		fmt.Printf("警告, taskId:%s, 检测状态:%s, 警告次数:%d, 违规详情:%s, 证据信息:%s", taskId, statusStr, warnCount, spamDetail, segments)
	} else if action == 3 {
		fmt.Printf("断流, taskId:%s, 检测状态:%s, 警告次数:%d, 违规详情:%s, 证据信息:%s", taskId, statusStr, warnCount, spamDetail, segments)
	} else if action == 4 {
		fmt.Printf("断流, taskId:%s, 检测状态:%s, 警告次数:%d, 违规详情:%s, 证据信息:%s", taskId, statusStr, prompCount, spamDetail, segments)
	} else {
		fmt.Printf("人审信息：%s", reviewEvidences)
	}

	fmt.Printf("================")
}

func main() {
	ret := check()

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		if resultArray == nil || len(resultArray) == 0 {
			fmt.Printf("暂时没有结果需要获取, 请稍后重试!")
		} else {
			for _, result := range resultArray {
				if resultMap, ok := result.(map[string]interface{}); ok {
					taskId := resultMap["taskId"].(string)
					callback := resultMap["callback"].(string)
					dataId := resultMap["dataId"].(string)
					fmt.Printf("taskId:%s, callback:%s, dataId:%s", taskId, callback, dataId)

					evidences, _ := resultMap["evidences"].(map[string]interface{})
					reviewEvidences, _ := resultMap["reviewEvidences"].(map[string]interface{})
					if evidences != nil {
						parseMachine(evidences, taskId)
					} else if reviewEvidences != nil {
						parseHuman(reviewEvidences, taskId)
					}
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
