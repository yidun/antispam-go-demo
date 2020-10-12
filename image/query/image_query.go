/*
@Author : yidun_dev
@Date : 2020-01-20
@File : image_query.go
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
	apiUrl     = "http://as.dun.163.com/v1/image/submit/task"
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

func main() {
	taskIds := []string{"093238e25fa54a3993849e72eb6040af", "8d26fded48d74263a9ad30e8199fdac7"}
	jsonString, _ := json.Marshal(taskIds)
	params := url.Values{"taskIds": []string{string(jsonString)}}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		for _, result := range resultArray {
			if resultMap, ok := result.(map[string]interface{}); ok {
				var name string
				if resultMap["name"] == nil {
					name = ""
				} else {
					name = resultMap["name"].(string)
				}
				taskId := resultMap["taskId"].(string)
				status, _ := resultMap["status"].(json.Number).Int64()
				labelArray, _ := resultMap["labels"].([]interface{})
				fmt.Printf("taskId: %s, status: %d, name: %s, labels: %s", taskId, status, name, labelArray)
				maxLevel := int64(-1)
				//产品需根据自身需求，自行解析处理，本示例只是简单判断分类级别
				for _, labelItem := range labelArray {
					if labelItemMap, ok := labelItem.(map[string]interface{}); ok {
						label, _ := labelItemMap["label"].(json.Number).Int64()
						level, _ := labelItemMap["level"].(json.Number).Int64()
						rate, _ := labelItemMap["rate"].(json.Number).Float64()
						fmt.Printf("label: %d, level: %d, rate: %d", label, level, rate)
						if level > maxLevel {
							maxLevel = level
						}
					}
				}
				if maxLevel == 0 {
					fmt.Printf("#图片查询结果: 最高等级为\"正常\"\n")
				} else if maxLevel == 1 {
					fmt.Printf("#图片查询结果: 最高等级为\"嫌疑\"\n")
				} else if maxLevel == 2 {
					fmt.Printf("#图片查询结果: 最高等级为\"确定\"\n")
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
