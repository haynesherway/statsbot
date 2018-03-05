package main

import (
	"fmt"
	"github.com/haynesherway/statsbot"
)

func main() {
	err := statsbot.ReadConfig()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	statsbot.Start()

	<-make(chan struct{})

	return
}
