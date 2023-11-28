/*
@Author : yidun_dev
@Date : 2023-11-24
@File : crawler_resource_query.go
@Version : 1.0
@Doc : https://support.dun.163.com/documents/606191408732381184?docId=716461276684730368  url检测详情批量查询接口
*/
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/tjfoc/gmsm/sm3"
)

const (
	apiUrl    = "https://as.dun.163.com/v3/crawler/callback/query"
	version   = "v3.0"
	secretId  = "your_secret_id"  //产品密钥ID，产品标识
	secretKey = "your_secret_key" //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
)

// 请求易盾接口
func query(params url.Values) *simplejson.Json {
	params["secretId"] = []string{secretId}
	params["version"] = []string{version}
	params["timestamp"] = []string{strconv.FormatInt(time.Now().UnixNano()/1000000, 10)}
	params["nonce"] = []string{strconv.FormatInt(rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(10000000000), 10)}
	params["signatureMethod"] = []string{"SM3"} // 签名方法支持国密SM3，默认MD5
	params["taskIdList"] = []string{"df9419b1b74f4fa8a63aa9faebdbc2b6, 87bff439c76c496e803c787b7257c94e"}
	params["signature"] = []string{genSignature(params)}

	resp, err := http.Post(apiUrl, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))

	if err != nil {
		fmt.Println("调用API接口失败:", err)
		return nil
	}

	defer resp.Body.Close()

	contents, _ := io.ReadAll(resp.Body)
	result, _ := simplejson.NewJson(contents)
	return result
}

// 生成签名信息
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
	params := url.Values{}

	ret := query(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		if len(resultArray) == 0 {
			fmt.Printf("Can't find Callback Data")
		} else {
			for _, result := range resultArray {
				if resultMap, ok := result.(map[string]interface{}); ok {
					if antispam, ok := resultMap["antispam"].(map[string]interface{}); ok {
						taskId := antispam["taskId"].(string)
						dataId, _ := antispam["dataId"].(string)
						suggestion, _ := antispam["suggestion"].(json.Number).Int64()
						fmt.Printf("SUCCESS: taskId=%s dataId=%s suggestion=%d \n", taskId, dataId, suggestion) // Fixed typo in dataId
					}
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
