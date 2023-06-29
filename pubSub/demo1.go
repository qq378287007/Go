package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// 订阅/发布模型
type (
	// 订阅者为一个通道
	subscriber chan interface{}
	//	主题为一个过滤器
	topicFunc func(v interface{}) bool
)

// Publisher 发布者对象
type Publisher struct {
	m           sync.RWMutex
	buffer      int
	timeout     time.Duration
	subscribers map[subscriber]topicFunc
}

// 构建一个发布者对象，可以蛇者超时时间和缓存队列长度
func NewPublisher(publishTimeout time.Duration, buffer int) *Publisher {
	return &Publisher{
		buffer:      buffer,
		timeout:     publishTimeout,
		subscribers: make(map[subscriber]topicFunc),
	}
}

// 添加一个新的订阅者，订阅过滤器筛选后的主题
func (p *Publisher) SubscribeTopic(topic topicFunc) chan interface{} {
	ch := make(chan interface{}, p.buffer)
	p.m.Lock()
	// 给指定的订阅者，加上主题过滤器
	p.subscribers[ch] = topic
	p.m.Unlock()
	return ch
}

// 添加一个新的订阅者，订阅所有主题
func (p *Publisher) Subscribe() chan interface{} {
	return p.SubscribeTopic(nil)
}

// 退出订阅
func (p *Publisher) Evict(sub chan interface{}) {
	p.m.Lock()
	defer p.m.Unlock()

	// 将该订阅者从发布者的信息中删除
	delete(p.subscribers, sub)
	close(sub)
}

// 发布一个主题
func (p *Publisher) Publish(v interface{}) {
	p.m.RLock()
	defer p.m.RUnlock()

	var wg sync.WaitGroup
	for sub, topic := range p.subscribers {
		wg.Add(1)
		go p.sendTopic(sub, topic, v, &wg)
	}
	// 等待主题发送完成
	wg.Wait()
}

// 发送主题，可以容忍一定的超时
func (p *Publisher) sendTopic(sub subscriber, topic topicFunc, v interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	// 如果该订阅者没有订阅全部，并且发布的主题又不符合主题过滤器
	// 那么直接返回
	if topic != nil && !topic(v) {
		return
	}

	// 一般 time.After 与 select case一同使用
	// 如果在指定时间内我们定义的通道中没有接受到值，
	// 那么将会执行<-time.After(p.timeout)
	// 是用于判断超时的操作
	select {
	case sub <- v:
	case <-time.After(p.timeout):
	}
}

func (p *Publisher) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	// 循环关闭所有的订阅者通道
	for sub := range p.subscribers {
		delete(p.subscribers, sub)
		close(sub)
	}
}

func main() {
	// 一个过期时间为 0.1秒，缓冲区大小为10的发布者
	// 发布者的缓冲区大小决定，订阅者的缓冲区大下
	// 如果发布的主题订阅者没有接受将会阻塞这个订阅者
	// 新发布的主题该订阅者无法在进行接收
	p := NewPublisher(100*time.Millisecond, 10)
	defer p.Close()

	// 添加两个订阅者，一个订阅全部，一个订阅"golang"
	all := p.Subscribe()
	golang := p.SubscribeTopic(
		// 主题过滤规则
		func(v interface{}) bool {
			//	是字符串类型吗？
			//	如果是，那么这个里面包含golang吗？
			//	满足上面的条件才是我这个订阅者想要的
			//	不然返回 false
			//	如果是 nil，就代表只要发送我就要，相当于全部订阅
			if s, ok := v.(string); ok {
				return strings.Contains(s, "golang")
			}
			return false
		})

	p.Publish("hello, world!")
	p.Publish("hello, golang!")
	p.Publish("golang!")

	go func() {
		for msg := range all {
			fmt.Println("all:", msg)
		}
	}()
	go func() {
		for msg := range golang {
			fmt.Println("golang:", msg)
		}
	}()

	//	运行一段时间后退出
	time.Sleep(3 * time.Second)
}
