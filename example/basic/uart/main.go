package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/traulfs/tsb"
)

const MyJack byte = 5 // select Jack 1-8

func main() {
	var buf []byte = make([]byte, 256)
	//serv, err := tsb.NewSerialServer("/dev/ttyUSB0")
	serv, err := tsb.NewTcpServer("localhost:3001")
	if err != nil {
		log.Fatal(err)
	}
	uart, err := tsb.NewUart(MyJack, serv)
	if err != nil {
		log.Fatal(err)
	}
	err = uart.Init(tsb.UartBaud115200, tsb.UartData8, tsb.UartParityNone, tsb.UartStopbits1)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			_, err := uart.Write([]byte("Hello Jack" + strconv.Itoa(int(uart.Jack)) + "\n"))
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(5 * time.Duration(time.Second))
		}
	}()
	go func() {
		for {
			//fmt.Printf("%d", jack)
			n, err := uart.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			if n > 0 {
				fmt.Printf("Received from Jack %d: %s\n", uart.Jack, buf)
			}
		}
	}()
}
