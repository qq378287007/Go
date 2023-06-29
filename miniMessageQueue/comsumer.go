package main

import (
	"bytes"
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

	msg := broker.Msg{Topic: "topic-test", MsgType: 1}

	n, err := conn.Write(broker.MsgToBytes(msg))
	if err != nil {
		fmt.Println("write failed, err:", err)
		os.Exit(1)
	}
	fmt.Println("n", n)

	var res [128]byte
	conn.Read(res[:])
	buf := bytes.NewBuffer(res[:])
	receMsg := broker.BytesToMsg(buf)
	fmt.Print(receMsg)

	// ack
	conn, _ = net.Dial("tcp", "127.0.0.1:12345")
	l, e := conn.Write(broker.MsgToBytes(broker.Msg{Id: receMsg.Id, Topic: receMsg.Topic, MsgType: 3}))
	if e != nil {
		fmt.Println("write failed, err:", err)
		os.Exit(1)
	}
	fmt.Println("l:", l)
}
