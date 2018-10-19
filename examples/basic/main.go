package main

import "github.com/gladiusio/legion"

func main() {
	conf := legion.SimpleConfig("localhost", 7946)
	l := legion.New(conf)

	go l.Listen()
}
