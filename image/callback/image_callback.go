/*
@Author : yidun_dev
@Date : 2020-01-20
@File : image_callback.go
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
	apiUrl     = "http://as.dun.163yun.com/v4/image/callback/results"
	version    = "v4"
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

func main() {
	ret := check()

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		antispamArray, _ := ret.Get("antispam").Array()
		if len(antispamArray) == 0 {
			fmt.Printf("暂时没有人工复审结果需要获取, 请稍后重试!")
		}
		for _, antispamResult := range antispamArray {
			if antispamMap, ok := antispamResult.(map[string]interface{}); ok {
				name := antispamMap["name"].(string)
				taskId := antispamMap["taskId"].(string)
				action, _ := antispamMap["action"].(json.Number).Int64()
				labelArray := antispamMap["labels"].([]interface{})
				fmt.Printf("taskId: %s, name: %s, action: %d", taskId, name, action)
				for _, labelItem := range labelArray {
					if labelItemMap, ok := labelItem.(map[string]interface{}); ok {
						label, _ := labelItemMap["label"].(json.Number).Int64()
						level, _ := labelItemMap["level"].(json.Number).Int64()
						rate, _ := labelItemMap["rate"].(json.Number).Float64()
						fmt.Printf("label: %d, level: %d, rate: %f", label, level, rate)
					}
				}
				if action == 0 {
					fmt.Printf("#图片人工复审结果: 最高等级为\"正常\"\n")
				} else if action == 2 {
					fmt.Printf("#图片人工复审结果: 最高等级为\"确定\"\n")
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
