package main

import (
	"fmt"
	"runtime"
)

func numberServer(addNumber <-chan int, giveNumber chan<- int, exit <-chan int) {
	var i = 0

	for {
		select {
		case j := <-addNumber:
			i += j
		case giveNumber <- i:
			//
		case <-exit:
			return
		}
	}
}

func incrementing(addNumber chan<- int, finished chan<- bool) {
	for j := 0; j < 1000000; j++ {
		addNumber <- 1
	}
	close(finished)
}

func decrementing(addNumber chan<- int, finished chan<- bool) {
	for j := 0; j < 999999; j++ {
		addNumber <- -1
	}
	close(finished)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	channelAddNumber := make(chan int)
	channelGiveNumber := make(chan int)
	channelFinishedIncr := make(chan bool)
	channelFinishedDecr := make(chan bool)
	channelExit := make(chan int)

	go numberServer(channelAddNumber, channelGiveNumber, channelExit)
	go incrementing(channelAddNumber, channelFinishedIncr)
	go decrementing(channelAddNumber, channelFinishedDecr)

	<-channelFinishedIncr
	<-channelFinishedDecr

	fmt.Println("The magic number is:", <-channelGiveNumber)
	channelExit <- 0
}
