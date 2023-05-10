package main

import (
	"fmt"
	"log"
	"time"

	"github.com/traulfs/tsb"
)

const LED_Green = 1
const LED_Red = 2

func main() {
	server, err := tsb.NewTcpServer("10.1.108.57:3000")
	//server, err := tsb.NewSerialServer("/dev/tty.usbmodem11201")
	if err != nil {
		log.Fatal(err)
	}
	input, err := tsb.NewPort(1, server)
	if err != nil {
		log.Fatal(err)
	}
	output, err := tsb.NewPort(3, server)
	if err != nil {
		log.Fatal(err)
	}
	input.Write(tsb.PortCharNibble(tsb.PortcharNotification, 0x0f))
	output.Write(tsb.PortCharNibble(tsb.PortcharNotification, 0x0f))
	b := make([]byte, 4096)
	for i := 0; i < 100; i++ {
		n, err := input.Read(b)
		if err != nil {
			log.Fatal(err)
		}
		if n > 0 {
			fmt.Printf("Input Read %d bytes: %x\n", n, b[:n])
		}
		n, err = output.Read(b)
		if err != nil {
			log.Fatal(err)
		}
		if n > 0 {
			fmt.Printf("Output Read %d bytes: %x\n", n, b[:n])
		}
		input.Write(tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Red))
		output.Write(tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Green))
		output.Write(tsb.PortCharNibble(tsb.PortcharSetDirection, 0x0f))
		output.Write(tsb.PortCharNibble(tsb.PortcharSetOutput, 0x0f))
		fmt.Printf("Toggle Red LED:   %x\n", tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Red))
		time.Sleep(1000 * time.Millisecond)
		input.Write(tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Green))
		output.Write(tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Red))
		output.Write(tsb.PortCharNibble(tsb.PortcharClearOutput, 0x0f))
		fmt.Printf("Toggle Green LED: %x\n", tsb.PortCharNibble(tsb.PortcharToggleLED, LED_Green))
		time.Sleep(1000 * time.Millisecond)
	}
}
