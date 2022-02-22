package main

import "fmt"

func main() {
	c := make(chan int, 5)
	fmt.Println("len c is", len(c), "cap c is ", cap(c))
	go func() {
		for i := 0; i < 5; i++ {
			c <- i
		}
		//close(c)
	}()
	for {
		if data, ok := <-c; ok {
			fmt.Println(data)
		} else {
			break
		}
	}
}
