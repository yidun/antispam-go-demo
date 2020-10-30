/*
@Author : yidun_dev
@Date : 2020-10-30
@File : image_asyncresult.go
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
	apiUrl     = "http://as.dun.163.com/v4/image/asyncResult"
	version    = "v4"
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
	var taskIds = []string{"202b1d65f5854cecadcb24382b681c1a", "0f0345933b05489c9b60635b0c8cc721"}
	jsonString, _ := json.Marshal(taskIds)
	params := url.Values{
		"taskIds": []string{string(jsonString)},
	}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		antispamArray, _ := ret.Get("antispam").Array()
		for _, antispamResult := range antispamArray {
			if antispamMap, ok := antispamResult.(map[string]interface{}); ok {
				name := antispamMap["name"].(string)
				taskId := antispamMap["taskId"].(string)
				status, _ := antispamMap["status"].(json.Number).Int64()
				//图片检测状态码，定义为：0：检测成功，610：图片下载失败，620：图片格式错误，630：其它
				if status == 0 {
					//图片维度结果
					action, _ := antispamMap["action"].(json.Number).Int64()
					labelArray := antispamMap["labels"].([]interface{})
					fmt.Printf("taskId: %s, status: %d, name: %s, action: %d", taskId, status, name, action)
					//产品需根据自身需求，自行解析处理，本示例只是简单判断分类级别
					for _, labelItem := range labelArray {
						if labelItemMap, ok := labelItem.(map[string]interface{}); ok {
							label, _ := labelItemMap["label"].(json.Number).Int64()
							level, _ := labelItemMap["level"].(json.Number).Int64()
							rate, _ := labelItemMap["rate"].(json.Number).Float64()
							subLabels := labelItemMap["subLabels"].([]interface{})
							fmt.Printf("label: %d, level: %d, rate: %f, subLabels: %s", label, level, rate, subLabels)
						}
					}
					if action == 0 {
						fmt.Printf("#图片机器检测结果: 最高等级为\"正常\"\n")
					} else if action == 1 {
						fmt.Printf("#图片机器检测结果: 最高等级为\"嫌疑\"\n")
					} else if action == 2 {
						fmt.Printf("#图片机器检测结果: 最高等级为\"确定\"\n")
					}
				} else {
					//status对应失败状态码：610：图片下载失败，620：图片格式错误，630：其它
					fmt.Printf("图片检测失败, taskId: %s, status: %d, name: %s", taskId, status, name)
				}
			}
		}
		ocrArray, _ := ret.Get("ocr").Array()
		for _, ocrResult := range ocrArray {
			if ocrMap, ok := ocrResult.(map[string]interface{}); ok {
				name := ocrMap["name"].(string)
				taskId := ocrMap["taskId"].(string)
				details := ocrMap["details"].([]interface{})
				fmt.Printf("taskId: %s, name: %s", taskId, name)
				//产品需根据自身需求，自行解析处理，本示例只是简单输出ocr结果信息
				for _, detail := range details {
					if detailMap, ok := detail.(map[string]interface{}); ok {
						content := detailMap["content"].(string)
						lineContents := detailMap["lineContents"]
						fmt.Printf("识别ocr文本内容: %s, ocr片段及坐标信息: %s", content, lineContents)
					}
				}
			}
		}
		faceArray, _ := ret.Get("face").Array()
		for _, faceResult := range faceArray {
			if faceMap, ok := faceResult.(map[string]interface{}); ok {
				name := faceMap["name"].(string)
				taskId := faceMap["taskId"].(string)
				details := faceMap["details"].([]interface{})
				fmt.Printf("taskId: %s, name: %s", taskId, name)
				//产品需根据自身需求，自行解析处理，本示例只是简单输出人脸结果信息
				for _, detail := range details {
					if detailMap, ok := detail.(map[string]interface{}); ok {
						faceNumber := detailMap["faceNumber"].(string)
						faceContents := detailMap["faceContents"]
						fmt.Printf("识别人脸数量: %s, 人物信息及坐标信息: %s", faceNumber, faceContents)
					}
				}
			}
		}
		qualityArray, _ := ret.Get("quality").Array()
		for _, qualityResult := range qualityArray {
			if qualityMap, ok := qualityResult.(map[string]interface{}); ok {
				name := qualityMap["name"].(string)
				taskId := qualityMap["taskId"].(string)
				details := qualityMap["details"].([]interface{})
				fmt.Printf("taskId: %s, name: %s", taskId, name)
				//产品需根据自身需求，自行解析处理，本示例只是简单输出质量结果信息
				for _, detail := range details {
					if detailMap, ok := detail.(map[string]interface{}); ok {
						aestheticsRate, _ := detailMap["aestheticsRate"].(json.Number).Float64()
						metaInfo := detailMap["metaInfo"]
						boarderInfo := detailMap["boarderInfo"]
						fmt.Printf("图片美观度分数:%f, 图片基本信息:%s, 图片边框信息:%s", aestheticsRate, metaInfo, boarderInfo)
					}
				}
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
