/*
@Author : yidun_dev
@Date : 2024-06-12
@File : aigc_callback.go
@Version : 1.0
@Golang : 1.13.5
@Doc : http://dun.163.com/api.html
*/
package callback

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

const (
	API_URL    = "https://as.dun.163.com/v1/stream/callback/results"
	SECRETID   = "your_secret_id"
	SECRETKEY  = "your_secret_key"
)

func main() {
	params := url.Values{
		"secretId":        []string{SECRETID},
		"version":         []string{"v1"},
		"timestamp":       []string{strconv.FormatInt(time.Now().UnixNano()/1000000, 10)},
		"nonce":           []string{strconv.FormatInt(rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(10000000000), 10)},
		"signatureMethod": []string{"MD5"},
	}

	params["signature"] = []string{genSignature(params)}

	resp, err := http.PostForm(API_URL, params)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	result, _ := simplejson.NewJson(body)

	code, _ := result.Get("code").Int()
	msg, _ := result.Get("msg").String()
	if code == 200 {
		resultArray, _ := result.Get("result").Array()
		if len(resultArray) == 0 {
			fmt.Println("暂时没有结果需要获取，请稍后重试！")
		} else {
			for _, streamCheckResult := range resultArray {
				streamCheckResultMap := streamCheckResult.(map[string]interface{})
				sessionTaskId := streamCheckResultMap["sessionTaskId"].(string)
				sessionId := streamCheckResultMap["sessionId"].(string)
				antispam := streamCheckResultMap["antispam"].(map[string]interface{})
				suggestion := antispam["suggestion"].(string)
				label := antispam["label"].(string)
				fmt.Printf("sessionTaskId=%s, sessionId=%s, suggestion=%s, label=%s\n",
					sessionTaskId, sessionId, suggestion, label)
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s\n", code, msg)
	}
}

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
	paramStr += SECRETKEY

	md5Reader := md5.New()
	md5Reader.Write([]byte(paramStr))
	return hex.EncodeToString(md5Reader.Sum(nil))
}