package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// access_token结构体
type AccessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int    `json:"expires_in"`
}

// access_token文件结构体
type AccessTokenFile struct {
	Mp         AccessToken `json:"token"`
	ExpireTime int         `json:"expire_time"`
}

var (
	// 公众号appid
	appid = "xxx"
	// 公众号密钥
	secret = "xxx"
	// 获取access_token链接
	accessTokenUrl = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	// 发送模板消息链接
	sendTemplateUrl = "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s"
	// 模板ID
	templateId = "xx-cUm1sc"
	// 心知天气apiKey
	apiKey = "xxx"
)

// 判断目录是否存在
func fileExist(path, fileName string) string {
	// 路径拼接
	filePath := filepath.Join(path, fileName)
	_, err := os.Stat(filePath)
	if err != nil {
		os.Create(filePath)
	}
	return filePath
}

// 保存Access_token
func saveAccessToken(tokenStruct AccessToken) {
	nowPath, _ := os.Getwd()
	fileName := fileExist(nowPath, "token.txt")
	// 过期时间
	expireTime := int(time.Now().Unix()) + tokenStruct.ExpiresIn
	// token信息
	tStruct := AccessTokenFile{
		tokenStruct,
		expireTime,
	}
	// 转化为json
	b, err := json.Marshal(tStruct)
	if err != nil {
		fmt.Println("save access_token fail, err = ", err)
		return
	}
	err = os.WriteFile(fileName, b, 0777)
	if err != nil {
		fmt.Println("write file fail, err = ", err)
		return
	}
	return
}

// 获取Access_token
func requestAccessToken() (tokenStruct AccessToken, err error) {
	// 拼接token链接
	tokenUrl := fmt.Sprintf(accessTokenUrl, appid, secret)
	// 请求链接
	tokenRsp, err := http.Get(tokenUrl)
	if err != nil {
		err = errors.New("require fail")
		return
	}
	body, _ := ioutil.ReadAll(tokenRsp.Body)
	defer tokenRsp.Body.Close()
	json.Unmarshal([]byte(body), &tokenStruct)
	// 保存token信息
	saveAccessToken(tokenStruct)
	return
}

// 获取token字符串
func getAccessToken() (token string, err error) {
	nowPath, _ := os.Getwd()
	fileName := fileExist(nowPath, "token.txt")
	// 读取文件
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return 
	}
	// 读取文件内容
	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return 
	}
	var atFile AccessTokenFile
	// 读取文件信息
	json.Unmarshal([]byte(content), &atFile)
	// 判断token是否过期
	if atFile.ExpireTime <= int(time.Now().Unix()) {
		var tokenStruct AccessToken
		// 重新获取token
		tokenStruct, err = requestAccessToken()
		if err != nil {
			return 
		}
		token = tokenStruct.Token
		return
	}
	token = atFile.Mp.Token
	return
}

// 模版消息数据，根据自己申请的模版修改
func templateData(weather map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{}, 5)
	// 第一段
	first := make(map[string]interface{}, 1)
	first["value"] = weather["location"].(map[string]interface{})["path"]
	first["color"] = "#173177"
	// 关键词1
	keyword1 := make(map[string]interface{}, 1)
	keyword1["value"] = time.Now().Format("2006-01-02 15:04:05")
	keyword1["color"] = "#173177"
	// 关键词2
	keyword2 := make(map[string]interface{}, 1)
	keyword2["value"] = weather["now"].(map[string]interface{})["text"]
	keyword2["color"] = "#173177"
	// 关键词3
	keyword3 := make(map[string]interface{}, 1)
	keyword3["value"] = weather["last_update"]
	keyword3["color"] = "#173177"
	// 备注
	remark := make(map[string]interface{}, 1)
	remark["value"] = "当前温度：" + weather["now"].(map[string]interface{})["temperature"].(string) + "°C"
	remark["color"] = "#173177"
	// 数据拼接
	data["first"] = first
	data["keyword1"] = keyword1
	data["keyword2"] = keyword2
	data["keyword3"] = keyword3
	data["remark"] = remark
	return data
}

// 发送模版
func sendTemplate(openId, token string, data map[string]interface{}) (err error) {
	// 请求接口
	templateUrl := fmt.Sprintf(sendTemplateUrl, token)
	// 发送信息拼接
	postData := make(map[string]interface{}, 1)
	postData["touser"] = openId
	postData["template_id"] = templateId
	postData["url"] = ""
	postData["data"] = data
	// 请求转化
	jsonStr, _ := json.Marshal(postData)
	var templateRsp *http.Response
	templateRsp, err = http.Post(templateUrl, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		return 
	}
	body, _ := ioutil.ReadAll(templateRsp.Body)
	defer templateRsp.Body.Close()
	fmt.Println(string(body))
	return
}

// 获取天气
func getWeather(local string) (data map[string]interface{}, err error) {
	// 请求接口
	weatherUri := fmt.Sprintf("https://api.seniverse.com/v3/weather/now.json?key=%s&location=%s&language=zh-Hans&unit=c", apiKey, local)
	var rsp *http.Response
	rsp, err = http.Get(weatherUri)
	if err != nil {
		err = errors.New("require weather fail")
		return
	}
	body, _ := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	// 定义返回格式
	var weatherJson map[string]interface{}
	json.Unmarshal([]byte(body), &weatherJson)
	// 格式转换
	weatherJsonInterface := weatherJson["results"].([]interface{})
	weather := weatherJsonInterface[0].(map[string]interface{})
	// 模版消息拼接
	data = templateData(weather)
	return
}

func main() {
	toUser := make([]map[string]interface{}, 0)
	user := make(map[string]interface{}, 1)
	user["user"] = "用户openid"
	user["local"] = []string{"广东番禺", "广东南沙"}
	// 可以多个添加
	toUser = append(toUser, user)
	// 获取token
	token, err := getAccessToken()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 循环要发送的用户
	for _, v := range toUser {
		for _, local := range v["local"].([]string) {
			data, err := getWeather(local)
			if err != nil {
				fmt.Println(err)
				break
			}
			sendTemplate(v["user"].(string), token, data)
		}
	}
}
