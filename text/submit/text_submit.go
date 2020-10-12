/*
@Author : yidun_dev
@Date : 2020-01-20
@File : text_submit.go
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
	sm3 "github.com/tjfoc/gmsm/sm3"
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
	apiUrl     = "http://as.dun.163.com/v1/text/submit"
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
	params := make(url.Values)
	var texts []map[string]string

	text1 := map[string]string{
		"dataId":  "ebfcad1c-dba1-490c-b4de-e784c2691768",
		"content": "易盾测试内容！v1接口!",
		"action":  "0"}
	text2 := map[string]string{
		"dataId":  "ebfcad1c-dba1-490c-b4de-e784c2691767",
		"content": "批量提交内容！v1接口!",
		"action":  "1"}
	texts = append(texts, text1, text2)
	jsonString, _ := json.Marshal(texts)
	params["texts"] = []string{string(jsonString)}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		for _, result := range resultArray {
			if resultMap, ok := result.(map[string]interface{}); ok {
				dataId := resultMap["dataId"].(string)
				taskId := resultMap["taskId"].(string)
				fmt.Printf("dataId: %s, 文本提交返回taskId: %s\n", dataId, taskId)
			}

		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
