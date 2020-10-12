/*
@Author : yidun_dev
@Date : 2020-01-20
@File : livewall_callback.go
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
	"github.com/tjfoc/gmsm/sm3"
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
	apiUrl     = "http://as.dun.163.com/v2/livewall/callback/results"
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
	// params["signatureMethod"] = []string{"SM3"} // 签名方法支持国密SM3，默认MD5
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
	if params["signatureMethod"] != nil && params["signatureMethod"][0] == "SM3" {
		sm3Reader := sm3.New()
		sm3Reader.Write([]byte(paramStr))
		return hex.EncodeToString(sm3Reader.Sum(nil))
	} else {
		md5Reader := md5.New()
		md5Reader.Write([]byte(paramStr))
		return hex.EncodeToString(md5Reader.Sum(nil))
	}
}

// 机审信息
func parseMachine(evidences map[string]interface{}, taskId string) {
	fmt.Printf("=== 视频机审信息 ===")
	evidence, _ := evidences["evidence"].(map[string]interface{})
	labels, _ := evidences["labels"].([]interface{})
	_, _ = evidence["type"].(json.Number).Int64()
	_, _ = evidence["url"].(string)
	_, _ = evidence["beginTime"].(json.Number).Int64()
	_, _ = evidence["endTime"].(json.Number).Int64()

	for _, labelItem := range labels {
		if labelItemMap, ok := labelItem.(map[string]interface{}); ok {
			_, _ = labelItemMap["label"].(json.Number).Int64()
			_, _ = labelItemMap["level"].(json.Number).Int64()
			_, _ = labelItemMap["rate"].(json.Number).Float64()
			_ = labelItemMap["subLabels"].([]interface{})
		}
	}
	fmt.Printf("Machine Evidence: %s", evidence)
	fmt.Printf("Machine Labels: %s", labels)
	fmt.Printf("================")
}

// 人审信息
func parseHuman(reviewEvidences map[string]interface{}, taskId string) {
	fmt.Printf("=== 人审信息 ===")
	action, _ := reviewEvidences["action"].(json.Number).Int64()
	_, _ = reviewEvidences["actionTime"].(json.Number).Int64()
	_, _ = reviewEvidences["label"].(json.Number).Int64()
	_, _ = reviewEvidences["detail"].(string)
	warnCount, _ := reviewEvidences["warnCount"].(json.Number).Int64()
	evidence, _ := reviewEvidences["evidence"].([]interface{})

	if action == 2 {
		fmt.Printf("警告, taskId:%s, 警告次数:%d, 证据信息:%s", taskId, warnCount, evidence)
	} else if action == 3 {
		fmt.Printf("断流, taskId:%s, 警告次数:%d, 证据信息:%s", taskId, warnCount, evidence)
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
			fmt.Printf("暂时没有结果需要获取，请稍后重试!")
		} else {
			for _, result := range resultArray {
				if resultMap, ok := result.(map[string]interface{}); ok {
					taskId := resultMap["taskId"].(string)
					callback := resultMap["callback"].(string)
					dataId := resultMap["dataId"].(string)
					status, _ := resultMap["status"].(json.Number).Int64()
					fmt.Printf("taskId:%s, dataId:%s, callback:%s, status:%d", taskId, dataId, callback, status)

					evidences, _ := resultMap["evidences"].(map[string]interface{})
					reviewEvidences, _ := resultMap["reviewEvidences"].(map[string]interface{}) //status, _ := resultMap["status"].(json.Number).Int64()
					if evidences != nil {
						parseMachine(evidences, taskId)
					} else if reviewEvidences != nil {
						parseHuman(reviewEvidences, taskId)
					} else {
						fmt.Printf("Invalid Result: %s", result)
					}
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
