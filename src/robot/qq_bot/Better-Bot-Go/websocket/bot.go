package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/fanliao/go-promise"
	"github.com/gorilla/websocket"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/onebot"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/util"
)

var Bots = make(map[string]map[string]*Bot)
var bots = new(sync.RWMutex)
var echo = ""

type Bot struct {
	BotId         string
	BotSecret     string
	Session       *SafeWebSocket
	mux           sync.RWMutex
	WaitingFrames map[string]*promise.Promise
}

func NewBot(xSelfId string, addr string, conn *websocket.Conn) *Bot {
	messageHandler := func(messageType int, data []byte) {
		bots.RLock()
		if _, ok := Bots[xSelfId]; ok {
			_, ok = Bots[xSelfId][addr]
			bots.RUnlock()
			if !ok {
				fmt.Printf("机器人 %s 地址 %s 已断开连接\n", xSelfId, addr)
				bots.Lock()
				delete(Bots[xSelfId], addr)
				bots.Unlock()
				_ = conn.Close()
				return
			}
		}
		bots.RUnlock()
	}
	closeHandler := func(code int, message string) {
		fmt.Printf("机器人 %s 地址 %s 已断开连接\n", xSelfId, addr)
		bots.Lock()
		delete(Bots[xSelfId], addr)
		bots.Unlock()
	}
	safeWs := NewSafeWebSocket(xSelfId, conn, messageHandler, closeHandler)
	bot := &Bot{
		BotId:         xSelfId,
		Session:       safeWs,
		WaitingFrames: make(map[string]*promise.Promise),
	}
	bots.RLock()
	if _, ok := Bots[xSelfId]; ok {
		bots.RUnlock()
		bots.Lock()
		Bots[xSelfId][addr] = bot
		bots.Unlock()
	} else {
		bots.RUnlock()
		bots.Lock()
		Bots[xSelfId] = map[string]*Bot{}
		Bots[xSelfId][addr] = bot
		bots.Unlock()
	}
	fmt.Printf("\n新机器人及地址已连接：%s 地址 %s\n", xSelfId, addr)
	fmt.Println("所有机器人及地址列表：")
	bots.RLock()
	for xId, xbot := range Bots {
		for ad, _ := range xbot {
			println(xId, ad)
		}
	}
	bots.RUnlock()
	return bot
}

func NewSecretBot(xSelfId, xBotSecret string, addr string, conn *websocket.Conn) *Bot {
	messageHandler := func(messageType int, data []byte) {
		bots.RLock()
		if _, ok := Bots[xSelfId]; ok {
			_, ok = Bots[xSelfId][addr]
			bots.RUnlock()
			if !ok {
				fmt.Printf("机器人 %s 地址 %s 已断开连接\n", xSelfId, addr)
				bots.Lock()
				delete(Bots[xSelfId], addr)
				bots.Unlock()
				_ = conn.Close()
				return
			}
		}
		bots.RUnlock()
	}
	closeHandler := func(code int, message string) {
		fmt.Printf("机器人 %s 地址 %s 已断开连接\n", xSelfId, addr)
		bots.Lock()
		delete(Bots[xSelfId], addr)
		bots.Unlock()
	}
	safeWs := NewSafeWebSocket(xSelfId, conn, messageHandler, closeHandler)
	bot := &Bot{
		BotId:         xSelfId,
		BotSecret:     xBotSecret,
		Session:       safeWs,
		WaitingFrames: make(map[string]*promise.Promise),
	}
	bots.RLock()
	if _, ok := Bots[xSelfId]; ok {
		bots.RUnlock()
		bots.Lock()
		Bots[xSelfId][addr] = bot
		bots.Unlock()
	} else {
		bots.RUnlock()
		bots.Lock()
		Bots[xSelfId] = map[string]*Bot{}
		Bots[xSelfId][addr] = bot
		bots.Unlock()
	}
	fmt.Printf("\n新机器人及地址已连接：%s 地址 %s\n", xSelfId, addr)
	fmt.Println("所有机器人及地址列表：")
	bots.RLock()
	for xId, xbot := range Bots {
		for ad, _ := range xbot {
			println(xId, ad)
		}
	}
	bots.RUnlock()
	return bot
}

func sendFrameAndWait(bot *Bot, appid string, frame *onebot.Frame) (*onebot.Frame, error) {
	frame.BotId = appid
	frame.Echo = util.GenerateIdStr()
	frame.Ok = true
	data, err := json.Marshal(frame)
	if err != nil {
		return nil, err
	}
	bot.Session.Send(websocket.BinaryMessage, data)
	p := promise.NewPromise()
	bot.setWaitingFrame(frame.Echo, p)
	defer bot.delWaitingFrame(frame.Echo)
	resp, err, timeout := p.GetOrTimeout(120000)
	if err != nil || timeout {
		return nil, err
	}
	respFrame, ok := resp.(*onebot.Frame)
	if !ok {
		return nil, errors.New("failed to convert promise result to resp frame")
	}
	return respFrame, nil
}

func (bot *Bot) setWaitingFrame(key string, value *promise.Promise) {
	bot.mux.Lock()
	defer bot.mux.Unlock()
	bot.WaitingFrames[key] = value
}

func (bot *Bot) getWaitingFrame(key string) (*promise.Promise, bool) {
	bot.mux.RLock()
	defer bot.mux.RUnlock()
	value, ok := bot.WaitingFrames[key]
	return value, ok
}

func (bot *Bot) delWaitingFrame(key string) {
	bot.mux.Lock()
	defer bot.mux.Unlock()
	delete(bot.WaitingFrames, key)
}

func NewPush(appid string, frame *onebot.Frame) error {
	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	bots.RLock()
	if xbot, ok := Bots[appid]; ok {
		for _, bot := range xbot {
			bot.Session.Send(websocket.TextMessage, data)
		}
	}
	bots.RUnlock()
	return nil
}

func NewSecretPush(appid, secret string, frame *onebot.Frame) error {
	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	bots.RLock()
	if xbot, ok := Bots[appid]; ok {
		for _, bot := range xbot {
			if secret == bot.BotSecret {
				bot.Session.Send(websocket.TextMessage, data)
			}
		}
	}
	bots.RUnlock()
	return nil
}
