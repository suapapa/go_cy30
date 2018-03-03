package main

import (
	"log"
	"time"
)

func main() {
	c := newCy30("/dev/ttyUSB0")
	defer c.Close()

	lastTime := time.Now()

	for i := 0; i < 10; i++ {
		r, err := c.SingleDistance()
		if err != nil {
			panic(err)
		}

		log.Println(time.Since(lastTime), r)
		lastTime = time.Now()
	}
}
