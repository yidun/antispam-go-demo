/*
@Author : yidun_dev
@Date : 2020-01-20
@File : image_submit.go
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
	apiUrl     = "http://as.dun.163.com/v1/image/submit"
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
	var images []map[string]string
	image1 := map[string]string{
		"name":  "image1",
		"data":  "https://nos.netease.com/yidun/2-0-0-a6133509763d4d6eac881a58f1791976.jpg",
		"level": "2",
		//"ip": "123.115.77.137",
		//"account": "golang@163.com",
		//"deviceId": "deviceId",
		//"callbackUrl": "http://***",  //主动回调地址url,如果设置了则走主动回调逻辑
	}
	image2 := map[string]string{
		"name":  "image2",
		"data":  "http://dun.163.com/public/res/web/case/sexy_normal_2.jpg?dda0e793c500818028fc14f20f6b492a",
		"level": "0",
		//"ip": "123.115.77.137",
		//"account": "golang@163.com",
		//"deviceId": "deviceId",
		//"callbackUrl": "http://***",  //主动回调地址url,如果设置了则走主动回调逻辑
	}
	images = append(images, image1, image2)
	jsonString, _ := json.Marshal(images)
	params := url.Values{"images": []string{string(jsonString)}}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		for _, result := range resultArray {
			if resultMap, ok := result.(map[string]interface{}); ok {
				name := resultMap["name"]
				taskId := resultMap["taskId"].(string)
				fmt.Printf("图片提交返回, name: %s, taskId: %s", name, taskId)
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
