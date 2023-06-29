package main

import (
	"fmt"
	"miniMessageQueue/broker"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:12345")
	if err != nil {
		fmt.Print("connect failed, err:", err)
		os.Exit(1)
	}
	defer conn.Close()

	msg := broker.Msg{Id: 1102, Topic: "topic-test", MsgType: 2, Payload: []byte("我")}
	n, err := conn.Write(broker.MsgToBytes(msg))
	if err != nil {
		fmt.Print("write failed, err:", err)
		os.Exit(1)
	}
	fmt.Print(n)
}
