package main

import (
	"github.com/gladiusio/legion"
	"github.com/gladiusio/legion/utils"
)

func main() {
	conf := legion.DefaultConfig("localhost", 7947)
	l := legion.New(conf)
	go l.Listen()
	l.Started()
	err := l.AddPeer(utils.LegionAddressFromString("localhost:7946"))
	if err != nil {
		panic(err)
	}
	select {}
}
