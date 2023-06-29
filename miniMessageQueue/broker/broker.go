package broker

import (
	"bufio"
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

type Msg struct {
	Id       int64
	TopicLen int64
	Topic    string
	// 1-consumer 2-producer 3-comsumer-ack 4-error
	MsgType int64  // 消息类型
	Len     int64  // 消息长度
	Payload []byte // 消息
}

func BytesToMsg(reader io.Reader) Msg {

	m := Msg{}
	var buf [128]byte
	n, err := reader.Read(buf[:])
	if err != nil {
		fmt.Println("read failed, err:", err)
	}
	fmt.Println("read bytes:", n)
	// id
	buff := bytes.NewBuffer(buf[0:8])
	binary.Read(buff, binary.LittleEndian, &m.Id)
	// topiclen
	buff = bytes.NewBuffer(buf[8:16])
	binary.Read(buff, binary.LittleEndian, &m.TopicLen)
	// topic
	msgLastIndex := 16 + m.TopicLen
	m.Topic = string(buf[16:msgLastIndex])
	// msgtype
	buff = bytes.NewBuffer(buf[msgLastIndex : msgLastIndex+8])
	binary.Read(buff, binary.LittleEndian, &m.MsgType)

	buff = bytes.NewBuffer(buf[msgLastIndex : msgLastIndex+16])
	binary.Read(buff, binary.LittleEndian, &m.Len)

	if m.Len <= 0 {
		return m
	}

	m.Payload = buf[msgLastIndex+16:]
	return m
}

func MsgToBytes(msg Msg) []byte {
	msg.TopicLen = int64(len([]byte(msg.Topic)))
	msg.Len = int64(len([]byte(msg.Payload)))

	var data []byte
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, msg.Id)
	data = append(data, buf.Bytes()...)

	buf = bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, msg.TopicLen)
	data = append(data, buf.Bytes()...)

	data = append(data, []byte(msg.Topic)...)

	buf = bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, msg.MsgType)
	data = append(data, buf.Bytes()...)

	buf = bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, msg.Len)
	data = append(data, buf.Bytes()...)
	data = append(data, []byte(msg.Payload)...)

	return data
}

type Queue struct {
	len  int
	data list.List
}

var lock sync.Mutex

func (queue *Queue) offer(msg Msg) {
	queue.data.PushBack(msg)
	queue.len = queue.data.Len()
}

func (queue *Queue) poll() Msg {
	if queue.len == 0 {
		return Msg{}
	}
	msg := queue.data.Front()
	return msg.Value.(Msg)
}

func (queue *Queue) delete(id int64) {
	lock.Lock()
	for msg := queue.data.Front(); msg != nil; msg = msg.Next() {
		if msg.Value.(Msg).Id == id {
			queue.data.Remove(msg)
			queue.len = queue.data.Len()
			break
		}
	}
	lock.Unlock()
}

var topics = sync.Map{}

func handleErr(conn net.Conn) {
	if err := recover(); err != nil {
		println(err.(string))
		conn.Write(MsgToBytes(Msg{MsgType: 4}))
	}
}

func Process(conn net.Conn) {
	defer handleErr(conn)
	reader := bufio.NewReader(conn)
	msg := BytesToMsg(reader)
	queue, ok := topics.Load(msg.Topic)
	var res Msg
	if msg.MsgType == 1 {
		// comsumer
		if queue == nil || queue.(*Queue).len == 0 {
			return
		}
		msg = queue.(*Queue).poll()
		msg.MsgType = 1
		res = msg
	} else if msg.MsgType == 2 {
		// producer
		if !ok {
			queue = &Queue{}
			queue.(*Queue).data.Init()
			topics.Store(msg.Topic, queue)
		}
		queue.(*Queue).offer(msg)
		res = Msg{Id: msg.Id, MsgType: 2}
	} else if msg.MsgType == 3 {
		// consumer ack
		if queue == nil {
			return
		}
		queue.(*Queue).delete(msg.Id)

	}
	conn.Write(MsgToBytes(res))

}

func Save() {
	ticker := time.NewTicker(60)
	for {
		select {
		case <-ticker.C:
			topics.Range(func(key, value interface{}) bool {
				if value == nil {
					return false
				}
				file, _ := os.Open(key.(string))
				if file == nil {
					file, _ = os.Create(key.(string))
				}
				for msg := value.(*Queue).data.Front(); msg != nil; msg = msg.Next() {
					file.Write(MsgToBytes(msg.Value.(Msg)))
				}
				file.Close()
				return false
			})
		default:
			time.Sleep(1)
		}
	}
}
