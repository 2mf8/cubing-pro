# Better-Bot-Go

## 由原来的 [Go-QQ-SDK](https://github.com/guojia99/cubing-pro/src/robot/qq_bot/Go-QQ-SDK) 迁移过来。

## 已支持正向 WebSocket ,正向 WebSocket 地址为 `wss://你的域名:端口/websocket` , 程序启动默认启用。

示例 `wss://fw1009zb5979.vicp.fun:443/websocket`

## 对应的 WebSocket 客户端地址 [Bot-Client-Go](https://github.com/guojia99/cubing-pro/src/robot/qq_bot/Bot-Client-Go) ,欢迎大家使用。

QQ频道机器人，官方 GOLANG SDK。

[![Go Reference](https://pkg.go.dev/badge/github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go.svg)](https://pkg.go.dev/github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go)

# [QQ交流群 677742758](https://qm.qq.com/q/okWktIaAqk)

<details>

<summary><font size="4">已完成功能/开发计划列表</font></summary>

### **登录**

- [x] 登录

### **消息类型**
- [x] 文本
- [x] 图片
- [x] 语音
- [x] MarkDown
- [ ] 表情
- [ ] At
- [ ] 回复
- [ ] 长消息(仅群聊/私聊)
- [ ] 链接分享
- [ ] 小程序(暂只支持RAW)
- [x] 短视频
- [ ] 合并转发
- [ ] 群文件(上传与接收信息)

### **群聊**

- [x] 收发群消息
- [x] 机器人加群通知
- [x] 机器人离群通知
- [x] 群接收机器人主动消息通知
- [x] 群拒绝机器人主动消息通知
- [x] 机器人撤回自己在2分钟内的消息
- [x] 机器人获取群成员列表【需要申请权限】

### **C2C**

- [x] 收发C2C消息
- [x] 机器人加好友通知
- [x] 机器人删好友通知
- [x] 接收机器人消息通知
- [x] 拒绝机器人消息通知
- [x] 机器人撤回自己在2分钟内的消息

</details>

## 一、如何使用

### 1.回调地址配置

#### is_open 为 true 时, 服务端只需要关注 port, cert_file 和 cert_key
https://你的域名:端口/qqbot/你的应用appid/你的应用app_secret

示例 `https://fw1009zb5979.vicp.fun:443/qqbot/101981675/hjksdfhi3jkslfjlksdfjksejkdjk`

#### is_open 为 false 时
https://你的域名:端口/qqbot/你的应用appid

示例 `https://fw1009zb5979.vicp.fun:443/qqbot/101981675`

### 2.配置文件填写（支持多账号）

默认配置文件为
```
{
	"apps": {
		"123456": {
			"qq": 123456,
			"app_id": 123456,
			"token": "你的AppToken",
			"app_secret": "你的AppSecret"
			"is_sandbox": true,
			"wss_addr": "你的wss地址，服务端不用管"
		}
	},
	"port": 8443,
	"cert_file": "ssl证书文件路径",
	"cert_key": "ssl证书密钥"
	"is_open": true
}
```
多账号

```
{
	"apps": {
		"5123456": {
			"qq": 123456,
			"app_id": 5123456,
			"token": "你的AppToken",
			"app_secret": "你的AppSecret"
			"is_sandbox": true,
			"wss_addr": "你的wss地址，服务端不用管"
		},
		"7234567": {
			"qq": 234567,
			"app_id": 7234567,
			"token": "你的AppToken",
			"app_secret": "你的AppSecret"
			"is_sandbox": true,
			"wss_addr": "你的wss地址，服务端不用管"
		}
	},
	"port": 8443,
	"cert_file": "ssl证书文件路径",
	"cert_key": "ssl证书密钥"
	"is_open": true
}
```

### 3.请求 openapi 接口，操作资源

```golang
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/dto"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/openapi"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/token"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/webhook"
	log "github.com/sirupsen/logrus"
)

var Apis = make(map[string]openapi.OpenAPI, 0)

func main() {
	webhook.InitLog()
	as := webhook.ReadSetting()
	var ctx context.Context
	for i, v := range as.Apps {
		token := token.BotToken(v.AppId, v.Token, string(token.TypeBot))
		api := bot.NewOpenAPI(token).WithTimeout(3 * time.Second)
		Apis[i] = api
	}
	b, _ := json.Marshal(as)
	fmt.Println("配置", string(b))
	webhook.GroupAtMessageEventHandler = func(bot *webhook.BotHeaderInfo, event *dto.WSPayload, data *dto.WSGroupATMessageData) error {
		fmt.Println(bot.XBotAppid, data.GroupId, data.Content)
		if len(data.Attachments) > 0 {
			log.Infof(`BotId(%s) GroupId(%s) UserId(%s) <- %s <image id="%s">`, bot.XBotAppid[0], data.GroupId, data.Author.UserId, data.Content, data.Attachments[0].URL)
		} else {
			log.Infof("BotId(%s) GroupId(%s) UserId(%s) <- %s", bot.XBotAppid[0], data.GroupId, data.Author.UserId, data.Content)
		}
		if strings.TrimSpace(data.Content) == "测试" {
			Apis[bot.XBotAppid[0]].PostGroupMessage(ctx, data.GroupId, &dto.GroupMessageToCreate{
				Content: "成功",
				MsgID:   data.MsgId,
				MsgType: 0,
			})
		}
		return nil
	}
	webhook.C2CMessageEventHandler = func(bot *webhook.BotHeaderInfo, event *dto.WSPayload, data *dto.WSC2CMessageData) error {
		b, _ := json.Marshal(event)
		fmt.Println(bot.XBotAppid, string(b), data.Content)
		return nil
	}
	webhook.MessageEventHandler = func(bot *webhook.BotHeaderInfo, event *dto.WSPayload, data *dto.WSMessageData) error {
		b, _ := json.Marshal(event)
		fmt.Println(bot.XBotAppid, string(b), data.Content)
		return nil
	}
	webhook.InitGin()
	select {}
}
```