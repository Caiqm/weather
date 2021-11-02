# 心知天气通过微信模版消息推送
> 开发语言golang
### 主要配置的信息
```
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
```
### 发送用户配置、查看的天气区域
```
func main() {
	...
	user["user"] = "用户openid"
	user["local"] = []string{"广东天河"}
	...
}
```
