package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var numberChan chan (int)
var prodChan chan (int)
var consChan chan (int)
var reachedCapacity int64
var bufferPos int64

func main() {
	rand.Seed(time.Now().UnixNano())
	numberChan = make(chan int, 100)
	prodChan = make(chan int)
	consChan = make(chan int)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)
	go consume(&wg, 300*time.Millisecond)
	go illustrate(&wg)
	go produce(ctx, 20*time.Millisecond, 1*time.Second)

	time.Sleep(30 * time.Second)
	cancel()
	wg.Wait()

	fmt.Printf("Buffer reached capacity %d times\r\n", reachedCapacity)
}

func produce(ctx context.Context, d1 time.Duration, d2 time.Duration) {
	x := 0
	for {
		select {
		case <-ctx.Done():
			close(numberChan)
			close(prodChan)
			return
		default:
			select {
			case numberChan <- x:
			default:
				atomic.AddInt64(&reachedCapacity, 1)
				numberChan <- x
			}
		}
		prodChan <- x
		x++
		atomic.AddInt64(&bufferPos, 1)
		time.Sleep(d1)
		if rand.Intn(10) == 0 {
			time.Sleep(d2)
		}
	}
}

func consume(wg *sync.WaitGroup, d time.Duration) {
	defer wg.Done()

	for {
		x, ok := <-numberChan
		if !ok {
			break
		}
		consChan <- x
		atomic.AddInt64(&bufferPos, -1)
		time.Sleep(d)
	}
	close(consChan)
}

func illustrate(wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		prod, cons int
		p, c       int
		okp, okc   bool
	)

	for {
		select {
		case p, okp = <-prodChan:
			if !okp {
				break
			}
			if p > prod {
				prod = p
			}
			draw(prod, cons, int(bufferPos))
		case c, okc = <-consChan:
			if !okc {
				break
			}
			if c > cons {
				cons = c
			}
			draw(prod, cons, int(bufferPos))
		}
		if !okp && !okc {
			break
		}
	}
	draw(prod, cons, int(bufferPos))
}

func draw(prod int, cons int, buf int) {
	fmt.Print("\033[H\033[2J") //clear screen
	fmt.Println(fmt.Sprintf("%s%s%d", strings.Repeat(" ", prod%100), "P:", prod))
	fmt.Println(strings.Repeat(" ", prod%100) + "|")
	fmt.Println(strings.Repeat("0123456789", 10))
	fmt.Println(strings.Repeat(" ", cons%100) + "|")
	fmt.Println(fmt.Sprintf("%s%s%d", strings.Repeat(" ", cons%100), "C:", cons))
	fmt.Println(strings.Repeat(" ", buf) + "|")
	fmt.Println(fmt.Sprintf("%s%s%d", strings.Repeat(" ", buf), "B:", buf))
}
