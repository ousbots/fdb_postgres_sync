package main

import ()

func main() {
	state := newState()

	go state.runWriter()

	for id := int64(0); id < 10; id++ {
		go state.dataHammer(id)
	}

	select {}
}
