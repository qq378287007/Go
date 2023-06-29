package main

import (
	"fmt"
	"sync"
	"time"
)

//var wg sync.WaitGroup

func doTask(n int, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Duration(n))
	fmt.Printf("Task %d Done\n", n)
}

func main() {
	var wg sync.WaitGroup
	n := 3
	wg.Add(n)
	for i := 0; i < n; i++ {
		go doTask(i+1, &wg)
	}
	wg.Wait()

	fmt.Println("All Task Done")
}
