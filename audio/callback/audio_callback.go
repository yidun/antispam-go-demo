/*
@Author : yidun_dev
@Date : 2020-01-20
@File : audio_callback.go
@Version : 1.0
@Golang : 1.13.5
@Doc : http://dun.163.com/api.html
*/
package main

import (
	"crypto/md5"
	"encoding/hex"
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
	apiUrl     = "http://as.dun.163.com/v4/audio/callback/results"
	version    = "v4"               //点播语音版本v3.2及以上二级细分类结构进行调整
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
	ret := check()

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		if resultArray == nil || len(resultArray) == 0 {
			fmt.Printf("暂时没有结果需要获取, 请稍后重试!")
		} else {
			for _, result := range resultArray {
				if resultMap, ok := result.(map[string]interface{}); ok {
					if resultMap["antispam"] != nil {
						/*antispam, _ := resultMap["antispam"].(map[string]interface{})
						taskId := antispam["taskId"].(string)
						status, _ := antispam["status"].(json.Number).Int64()
						if status == 2 {
							fmt.Printf("CHECK SUCCESS: taskId=%s", taskId)
							suggestion, _ := antispam["suggestion"].(json.Number).Int64()
							resultType, _ := antispam["resultType"].(json.Number).Int64()
							segmentArray := resultMap["segments"].([]interface{})
							if segmentArray == nil && len(segmentArray) > 0 {
								fmt.Printf("暂无反垃圾检测数据")
							} else {
								for _, segment := range segmentArray {
									if segmentMap, ok := segment.(map[string]interface{}); ok {
										startTime, _ := segmentMap["startTime"].(json.Number).Int64()
										endTime, _ := segmentMap["endTime"].(json.Number).Int64()
										content, _ := segmentMap["content"].(string)
									}
								}
							}
						}*/
					}
					if resultMap["language"] != nil {
						//language, _ := resultMap["language"].(map[string]interface{})
						//taskId, _ := language["taskId"].(string)
						//dataId, _ := language["dataId"].(string)
						//callback, _ := language["callback"].(string)
					}
					if resultMap["asr"] != nil {
						//asr, _ := resultMap["asr"].(map[string]interface{})
						//taskId, _ := asr["taskId"].(string)
						//dataId, _ := asr["dataId"].(string)
						//callback, _ := asr["callback"].(string)
					}
					if resultMap["voice"] != nil {
						//voice, _ := resultMap["voice"].(map[string]interface{})
						//taskId, _ := voice["taskId"].(string)
						//dataId, _ := voice["dataId"].(string)
						//callback, _ := voice["callback"].(string)
					}
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
