/*
@Author : yidun_dev
@Date : 2020-10-29
@File : imagelist_query.go
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
	apiUrl     = "http://as.dun.163.com/v1/image/list/pageQuery"
	version    = "v1.0"
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
		"pageNum":   []string{"1"},
		"pageSize":  []string{"20"},
		"startTime": []string{"1598951727666"},
		"endTime":   []string{"1598961727666"},
		"type":      []string{"0"},
		"listType":  []string{"2"},
		"status":    []string{"1"},
	}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		result, _ := ret.Get("result").Map()
		count, _ := result["count"].(json.Number).Int64()
		rows, _ := result["rows"].([]interface{})
		if rows != nil && len(rows) > 0 {
			for _, row := range rows {
				if rowMap, ok := row.(map[string]interface{}); ok {
					_, _ = rowMap["businessId"].(json.Number).Int64()
					_, _ = rowMap["productId"].(json.Number).Int64()
					_, _ = rowMap["uuid"].(string)
					_, _ = rowMap["url"].(string)
					_, _ = rowMap["hitCount"].(json.Number).Int64()
					_, _ = rowMap["imageLabel"].(json.Number).Int64()
					_, _ = rowMap["status"].(json.Number).Int64()
					_, _ = rowMap["listType"].(json.Number).Int64()
					_, _ = rowMap["nosPath"].(string)
				}
			}
		}
		fmt.Printf("count:%d, rows:%s", count, rows)
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
