package websocket

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/util"
	log "github.com/sirupsen/logrus"
)

// safe websocket
type SafeWebSocket struct {
	Conn          *websocket.Conn
	SendChannel   chan *WebSocketSendingMessage
	OnRecvMessage func(messageType int, data []byte)
	OnClose       func(int, string)
}

type WebSocketSendingMessage struct {
	MessageType int
	Data        []byte
}

func (ws *SafeWebSocket) Send(messageType int, data []byte) {
	ws.SendChannel <- &WebSocketSendingMessage{
		MessageType: messageType,
		Data:        data,
	}
}

func NewSafeWebSocket(appid string, conn *websocket.Conn, OnRecvMessage func(messageType int, data []byte), onClose func(int, string)) *SafeWebSocket {
	ws := &SafeWebSocket{
		Conn:          conn,
		SendChannel:   make(chan *WebSocketSendingMessage, 100),
		OnRecvMessage: OnRecvMessage,
		OnClose:       onClose,
	}

	conn.SetCloseHandler(func(code int, text string) error {
		ws.OnClose(code, text)
		return nil
	})

	// 接受消息
	util.SafeGo(func() {
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				addr := conn.RemoteAddr()
				fmt.Printf("机器人 %s 地址 %s 已断开连接\n", appid, addr.String())
				delete(Bots[appid], addr.String())
				log.Errorf("failed to read message, err: %+v", err)
				_ = conn.Close()
				return
			}
			if messageType == websocket.PingMessage {
				ws.Send(websocket.PongMessage, []byte("pong"))
				continue
			}
			ws.OnRecvMessage(messageType, data)
		}
	})

	// 发送消息
	util.SafeGo(func() {
		for sendingMessage := range ws.SendChannel {
			if ws.Conn == nil {
				log.Errorf("failed to send websocket message, conn is nil")
				return
			}
			err := ws.Conn.WriteMessage(sendingMessage.MessageType, sendingMessage.Data)
			if err != nil {
				addr := conn.RemoteAddr()
				fmt.Printf("机器人 %s 地址 %s 已断开连接\n", appid, addr.String())
				delete(Bots[appid], addr.String())
				log.Errorf("failed to send websocket message, %+v", err)
				_ = conn.Close()
				return
			}
		}
	})
	return ws
}
