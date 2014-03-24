package main

import (
	"flag"
	"github.com/creack/multio/logger"
)

var hello = logger.New(nil, "hello", 2)
var toto = logger.New(nil, "hello", 3)

func main() {
	flag.Parse()
	hello.Infof("hello world!\n")
	toto.Infof("toto\n")
}
