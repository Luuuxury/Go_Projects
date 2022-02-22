package main

import (
	"fmt"
	"time"
)

func main() {
	c := make(chan int, 2)
	fmt.Println("len c is ", len(c), "cap c is ", cap(c))

	go func() {
		defer fmt.Println("sub go is dnoe!")

		for i := 0; i < 3; i++ {
			c <- i
			fmt.Println("sub-go is sending element", i, "Len c is", len(c), "cap c is ", cap(c))
		}
	}()

	time.Sleep(2 * time.Second)
	for i := 0; i < 3; i++ {
		num := <-c
		fmt.Println("num is ", num)
	}
	fmt.Println("main done!")

}
