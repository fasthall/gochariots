package main

import (
	"fmt"
	"time"
)

func send(c chan int) {
	for i := 0; i < 1000; i++ {
		if i%10 == 0 {
			time.Sleep(time.Nanosecond * 1)
		}
		c <- i
	}
}

func sweep(c chan int) {
	buf := make([]int, 0, 10)
	for {
		select {
		case i := <-c:
			buf = append(buf, i)
			if cap(buf) == len(buf) {
				fmt.Println(buf)
				buf = buf[:0]
			}
		default:
			if len(buf) > 0 {
				fmt.Printf("def ")
				fmt.Println(buf)
				buf = buf[:0]
			}
		}
	}
}

func main() {
	c := make(chan int, 3)

	go send(c)
	go sweep(c)
	time.Sleep(time.Second * 300)
}
