/*
@Author : yidun_dev
@Date : 2024-06-12
@File : aigc_check.go
@Version : 1.0
@Golang : 1.13.5
@Doc : http://dun.163.com/api.html
*/
package check

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	SECRETID  = "your_secret_id"
	SECRETKEY = "your_secret_key"
	API_URL   = "https://as.dun.163.com/v1/stream/push"
)

type AigcStreamPushAPIDemo struct {
	HttpClient *http.Client
}

func NewAigcStreamPushAPIDemo() *AigcStreamPushAPIDemo {
	return &AigcStreamPushAPIDemo{
		HttpClient: &http.Client{},
	}
}

func (a *AigcStreamPushAPIDemo) pushDemoForInputCheck(sessionId string) {
	params := a.prepareParams()
	params.Set("sessionId", sessionId)
	params.Set("type", "2")
	params.Set("dataId", "yourDataId")
	params.Set("content", "当前会话输入的内容")
	params.Set("publishTime", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))
	a.invokeAndParseResponse(params)
}

func (a *AigcStreamPushAPIDemo) prepareParams() url.Values {
	params := url.Values{}
	params.Set("secretId", SECRETID)
	params.Set("version", "v1")
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))
	params.Set("nonce", strconv.FormatInt(rand.Int63(), 10))
	params.Set("signatureMethod", "MD5")
	return params
}

func (a *AigcStreamPushAPIDemo) invokeAndParseResponse(params url.Values) {
	signature := a.genSignature(params)
	params.Set("signature", signature)
	response, err := a.HttpClient.Post(API_URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	code := result["code"].(float64)
	msg := result["msg"].(string)
	if code == 200 {
		streamCheckResult := result["result"].(map[string]interface{})
		if streamCheckResult != nil {
			sessionTaskId := streamCheckResult["sessionTaskId"].(string)
			sessionIdReturn := streamCheckResult["sessionId"].(string)
			antispam := streamCheckResult["antispam"].(map[string]interface{})
			fmt.Printf("sessionTaskId=%s, sessionId=%s, antispam=%s", sessionTaskId, sessionIdReturn, antispam)
		}
	} else {
		fmt.Printf("ERROR: code=%f, msg=%s\n", code, msg)
	}
}

func (a *AigcStreamPushAPIDemo) genSignature(params url.Values) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var srcStr string
	for _, k := range keys {
		srcStr += k + "=" + params.Get(k) + "&"
	}
	srcStr = strings.TrimRight(srcStr, "&")
	h := hmac.New(sha1.New, []byte(SECRETKEY))
	h.Write([]byte(srcStr))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature
}

func main() {
	sessionId := "yourSessionId" + strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	demo := NewAigcStreamPushAPIDemo()
	demo.pushDemoForInputCheck(sessionId)
}