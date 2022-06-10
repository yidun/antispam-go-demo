/*
@Author : yidun_dev
@Date : 2020-01-20
@File : mediasolution_submit.go
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
	apiUrl    = "http://as.dun.163.com/v2/mediasolution/submit"
	version   = "v2.0"
	secretId  = "your_secret_id"  //产品密钥ID，产品标识
	secretKey = "your_secret_key" //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
)

//请求易盾接口
func check(params url.Values) *simplejson.Json {
	params["secretId"] = []string{secretId}
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
	var arr []map[string]string
	text := map[string]string{
		"type": "text",
		"data": "融媒体文本段落",
	}
	image := map[string]string{
		"type": "image",
		"data": "http://xxx",
	}
	audio := map[string]string{
		"type": "audio",
		"data": "http://xxx",
	}
	video := map[string]string{
		"type": "video",
		"data": "http://xxx",
	}
	audiovideo := map[string]string{
		"type": "audiovideo",
		"data": "http://xxx",
	}
	file := map[string]string{
		"type": "file",
		"data": "http://xxx",
	}

	arr = append(arr, text, image, audio, video, audiovideo, file)
	jsonString, _ := json.Marshal(arr)
	params := url.Values{
		"title":   []string{"融媒体解决方案的标题"},
		"content": []string{string(jsonString)},
	}

	ret := check(params)
	// 展示了异步结果返回示例，同步结果返回示例请参考官网开发文档
	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		result, _ := ret.Get("result").Map()
		if antispam, ok := result["antispam"].(map[string]interface{}); ok {
			taskId := antispam["taskId"].(string)
			dataId := antispam["dataId"].(string)
			fmt.Printf("SUBMIT SUCCESS: taskId=%s, dataId=%s", taskId, dataId)
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
