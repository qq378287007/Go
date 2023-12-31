package main

import (
	"context"
	"fmt"
	"time"
)

func reqTask(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("stop", name, ctx.Err())
			return
		default:
			fmt.Println(name, "send request")
			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(1*time.Second))
	n := 3
	for i := 0; i < n; i++ {
		go reqTask(ctx, fmt.Sprintf("worker%d", i+1))
	}

	time.Sleep(3 * time.Second)
	fmt.Println("before cancel")
	cancel()
	time.Sleep(3 * time.Second)
}
