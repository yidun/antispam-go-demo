/*
@Author : yidun_dev
@Date : 2020-01-20
@File : mediasolution_callback.go
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
	apiUrl    = "http://as.dun.163.com/v2/mediasolution/callback/results"
	version   = "v2"
	secretId  = "your_secret_id"  //产品密钥ID，产品标识
	secretKey = "your_secret_key" //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
)

//请求易盾接口
func check() *simplejson.Json {
	params := url.Values{}
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
	ret := check()

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		resultArray, _ := ret.Get("result").Array()
		if resultArray == nil || len(resultArray) == 0 {
			fmt.Printf("暂时没有结果需要获取，请稍后重试！")
		} else {
			for _, resultItem := range resultArray {
				if resultMap, ok := resultItem.(map[string]interface{}); ok {
					if antispam, ok := resultMap["antispam"].(map[string]interface{}); ok {
						_ = antispam["taskId"].(string)
						_ = antispam["dataId"].(string)
						_ = antispam["callback"].(string)
						_, _ = antispam["checkStatus"].(json.Number).Int64()
						_, _ = antispam["result"].(json.Number).Int64()
						if resultMap["evidences"] != nil {
							evidences, _ := antispam["evidences"].(map[string]interface{})
							if evidences["texts"] != nil {
								texts := evidences["texts"].([]interface{})
								for _, textItem := range texts {
									if textMap, ok := textItem.(map[string]interface{}); ok {
										dataId, _ := textMap["dataId"].(string)
										suggestion, _ := textMap["suggestion"].(json.Number).Int64()
										fmt.Printf("文本信息, dataId:%s, 建议动作:%d", dataId, suggestion)
									}
								}
							} else if evidences["images"] != nil {
								images := evidences["images"].([]interface{})
								for _, imageItem := range images {
									if imageMap, ok := imageItem.(map[string]interface{}); ok {
										dataId, _ := imageMap["dataId"].(string)
										status, _ := imageMap["status"].(json.Number).Int64()
										suggestion, _ := imageMap["suggestion"].(json.Number).Int64()
										fmt.Printf("图片信息, dataId:%s, 检测状态:%d, 建议动作:%d", dataId, status, suggestion)
									}
								}
							} else if evidences["audios"] != nil {
								audios := evidences["audios"].([]interface{})
								for _, audioItem := range audios {
									if audioMap, ok := audioItem.(map[string]interface{}); ok {
										dataId, _ := audioMap["dataId"].(string)
										status, _ := audioMap["asrStatus"].(json.Number).Int64()
										suggestion, _ := audioMap["suggestion"].(json.Number).Int64()
										fmt.Printf("语音信息, dataId:%s, 检测状态:%d, 建议动作:%d", dataId, status, suggestion)
									}
								}
							} else if evidences["audiovideos"] != nil {
								audiovideos := evidences["audiovideos"].([]interface{})
								for _, audiovideoItem := range audiovideos {
									if audiovideoMap, ok := audiovideoItem.(map[string]interface{}); ok {
										dataId, _ := audiovideoMap["dataId"].(string)
										suggestion, _ := audiovideoMap["suggestion"].(json.Number).Int64()
										fmt.Printf("音视频信息, dataId:%s, 建议动作:%d", dataId, suggestion)
									}
								}
							} else if evidences["files"] != nil {
								files := evidences["files"].([]interface{})
								for _, fileItem := range files {
									if fileMap, ok := fileItem.(map[string]interface{}); ok {
										dataId, _ := fileMap["dataId"].(string)
										suggestion, _ := fileMap["suggestion"].(json.Number).Int64()
										fmt.Printf("文档信息, dataId:%s, 建议动作:%d", dataId, suggestion)
									}
								}
							}
						}
					}

				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
