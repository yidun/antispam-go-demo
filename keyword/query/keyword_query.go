/*
@Author : yidun_dev
@Date : 2020-06-08
@File : keyword_query.go
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
	apiUrl     = "http://as.dun.163.com/v1/keyword/query"
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
	params := url.Values{
		"id":        []string{"163"},
		"category":  []string{"100"},
		"keyword":   []string{"色情敏感词"},
		"orderType": []string{"1"},
		"pageNum":   []string{"1"},
		"pageSize":  []string{"20"},
	}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		result, _ := ret.Get("result").Map()
		words := result["words"].(map[string]interface{})
		count, _ := words["count"].(json.Number).Int64()
		rows, _ := words["rows"].([]interface{})
		for _, row := range rows {
			if rowMap, ok := row.(map[string]interface{}); ok {
				id, _ := rowMap["id"].(json.Number).Int64()
				word := rowMap["word"].(string)
				category, _ := rowMap["category"].(json.Number).Int64()
				status, _ := rowMap["status"].(json.Number).Int64()
				updateTime, _ := rowMap["updateTime"].(json.Number).Int64()
				fmt.Printf("敏感词查询成功，count: %d, id: %d，keyword: %s，category: %d，status: %d，updateTime: %d \n",
					count, id, word, category, status, updateTime)
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
