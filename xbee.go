package main

import (
	"fmt"
	"github.com/tarm/serial"
)

var s *serial.Port

func rroutine() {
	for {
		s.Write([]byte("test"))
	}
}

func wroutine() {
	buf := make([]byte, 1)
	for {
		s.Read(buf)
		fmt.Print(string(buf))
	}
}

func main() {
	var err error
	c := &serial.Config{Name: "/dev/ttyS0", Baud: 9600}
	s, err = serial.OpenPort(c)
	if err != nil {
		panic(err)
	}

	go rroutine()
	wroutine()
}
