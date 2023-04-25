package main

import (
	"fmt"
	"log"
	"time"

	"github.com/traulfs/tsb"
)

const MyJack byte = 3 // select Jack 1-8

const LED_Red = 1
const LED_Green = 2

func main() {
	server, err := tsb.NewTcpServer("localhost:3010")
	//server, err := tsb.NewSerialServer("/dev/tty.usbmodem11201")
	if err != nil {
		log.Fatal(err)
	}
	port, err := tsb.NewPort(MyJack, server)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		port.Write(tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Red))
		fmt.Printf("Toggle Red LED:   %x\n", tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Red))
		time.Sleep(1000 * time.Millisecond)
		port.Write(tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Green))
		fmt.Printf("Toggle Green LDE: %x\n", tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Green))
		time.Sleep(1000 * time.Millisecond)
	}
}
