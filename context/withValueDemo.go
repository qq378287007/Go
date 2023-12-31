package main

import (
	"context"
	"fmt"
	"time"
)

type Options struct{ Interval time.Duration }

func reqTask(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("stop", name)
			return
		default:
			fmt.Println(name, "send request")
			op := ctx.Value("options").(*Options)
			time.Sleep(op.Interval * time.Second)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	vCtx := context.WithValue(ctx, "options", &Options{1})

	n := 3
	for i := 0; i < n; i++ {
		go reqTask(vCtx, fmt.Sprintf("worker%d", i+1))
	}

	time.Sleep(3 * time.Second)
	cancel()
	time.Sleep(3 * time.Second)
}
