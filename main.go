package main

import (
	"time"
)

func main() {
	state := newState()
	go state.runWriter()

	for id := int64(0); id < 10; id++ {
		go state.dataHammer(id, 100*time.Millisecond)
	}

	select {}
}
