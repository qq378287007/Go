package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ch := make(chan string, 1)
	consol := bufio.NewScanner(os.Stdin)
	for consol.Scan() {
		action := consol.Text()
		item := strings.Split(action, " ")
		switch item[0] {
		case "push":
			if len(item) != 2 {
				fmt.Println("must be set value")
				continue
			}
			select {
			case ch <- item[1]:
				continue
			case <-time.After(10 * time.Second):
				fmt.Println("time out ")
				return
			}
		case "pop":
			fmt.Println("执行pop....")
			select {
			case v := <-ch:
				fmt.Println(v)
			case <-time.After(10 * time.Second):
				fmt.Println("time out ")
				return
			}
		case "quit", "exit":
			return
		default:
			fmt.Println("unknown action")
		}
	}
}
