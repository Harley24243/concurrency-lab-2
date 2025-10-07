package main

import (
	"fmt"
	"sync"
)

func main_chan() {
	c := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(chan int) {
			sum := <-c
			wg.Done()
			c <- sum + 1
		}(c)
	}

	c <- 0
	wg.Wait()
	fmt.Println(<-c)
}

func main() {
	sum := 0
	var wg sync.WaitGroup
	var m sync.Mutex
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			m.Lock()
			sum = sum + 1
			wg.Done()
			m.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println(sum)
}
