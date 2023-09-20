package main

import (
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	// 连接WebSocket服务器
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 发送消息
	err = conn.WriteMessage(websocket.TextMessage, []byte("Hello, world!"))
	if err != nil {
		log.Fatal(err)
	}

	// 读取消息
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Received Type:", messageType)
	log.Println("Received message:", string(p))
}
