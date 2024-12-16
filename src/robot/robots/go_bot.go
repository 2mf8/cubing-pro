package robots

import (
	"context"
	"encoding/json"
	"fmt"
	bot "github.com/2mf8/Go-QQ-SDK"
	"github.com/2mf8/Go-QQ-SDK/dto"
	"github.com/2mf8/Go-QQ-SDK/log"
	"github.com/2mf8/Go-QQ-SDK/openapi"
	"github.com/2mf8/Go-QQ-SDK/token"
	"github.com/2mf8/Go-QQ-SDK/webhook"
	"strings"
	"time"
)

type QQBot struct {
}

func (q *QQBot) Run() {
	var Apis = make(map[string]openapi.OpenAPI, 0)

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
