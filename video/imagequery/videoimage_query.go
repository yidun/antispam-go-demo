/*
@Author : yidun_dev
@Date : 2020-07-15
@File : videoimage_query.go
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
	apiUrl     = "http://as.dun.163.com/v1/video/query/image"
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
	params := url.Values{
		"taskId":    []string{"4bc345f4bdc74a92b64543b35412d678"},
		"levels":    []string{"[0,1,2]"},
		"pageNum":   []string{"1"},
		"pageSize":  []string{"20"},
		"orderType": []string{"3"}, // 详情查看官网VideoDataOderType
	}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		result := ret.Get("result")
		status, _ := result.Get("status").Int()
		if status == 0 {
			images := result.Get("images")
			count, _ := images.Get("count").Int()
			rows, _ := images.Get("rows").Array()
			for _, row := range rows {
				if rowMap, ok := row.(map[string]interface{}); ok {
					picUrl, _ := rowMap["url"].(string)
					label, _ := rowMap["label"].(json.Number).Int64()
					labelLevel, _ := rowMap["labelLevel"].(json.Number).Int64()
					beginTime, _ := rowMap["beginTime"].(json.Number).Int64()
					endTime, _ := rowMap["endTime"].(json.Number).Int64()
					fmt.Printf("成功, count: %d, url: %s, label: %d, labelLevel: %d, 开始时间: %d, 结束时间: %d",
						count, picUrl, label, labelLevel, beginTime, endTime)
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
