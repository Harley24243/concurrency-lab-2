package main

import (
	"fmt"
	"sync"

	//"github.com/ChrisGora/semaphore"
	"math/rand"
	"time"

	"github.com/ChrisGora/semaphore"
)

type buffer struct {
	b                 []int
	size, read, write int
}

func newBuffer(size int) buffer {
	return buffer{
		b:     make([]int, size),
		size:  size,
		read:  0,
		write: 0,
	}
}

func (buffer *buffer) get() int {
	x := buffer.b[buffer.read]
	fmt.Println("Get\t", x, "\t", buffer)
	buffer.read = (buffer.read + 1) % len(buffer.b)
	return x
}

func (buffer *buffer) put(x int) {
	buffer.b[buffer.write] = x
	fmt.Println("Put\t", x, "\t", buffer)
	buffer.write = (buffer.write + 1) % len(buffer.b)
}

func producer(buffer *buffer, start, delta int, space semaphore.Semaphore, work semaphore.Semaphore, m *sync.Mutex) {
	x := start
	for {
		space.Wait()
		m.Lock()
		buffer.put(x)
		x = x + delta
		m.Unlock()
		work.Post()
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	}
}

func consumer(buffer *buffer, space semaphore.Semaphore, work semaphore.Semaphore, m *sync.Mutex) {
	for {
		work.Wait()
		m.Lock()
		_ = buffer.get()
		m.Unlock()
		space.Post()
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
	}
}

func main() {
	buffer := newBuffer(5)

	space := semaphore.Init(5, 5)
	work := semaphore.Init(5, 0)

	var m sync.Mutex

	go producer(&buffer, 1, 1, space, work, &m)
	go producer(&buffer, 1000, -1, space, work, &m)

	consumer(&buffer, space, work, &m)
}
