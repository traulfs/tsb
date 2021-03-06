package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/traulfs/tsb"
)

func main() {
	serv, err := tsb.NewSerialServer("/dev/ttyUSB0")
	//serv, err := tsb.NewTcpServer("loaclhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	go uartExample(serv, 1)
	go uartExample(serv, 2)
	go portExample(serv, 3)
	go portExample(serv, 4)
	go i2cExample(serv, 5)
}

func uartExample(s tsb.Server, jack int) {
	GetChan, PutChan, err := s.UartInit(1, tsb.UartBaud115200, tsb.UartData8&tsb.UartParityNone&tsb.UartStopbits1)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		PutChan <- "Hello Chan" + strconv.Itoa(jack)
		time.Sleep(time.Duration(jack) * time.Second)
	}()
	for {
		fmt.Printf("Received from Jack%d: %s\n\r", jack, <-GetChan)
	}
}

func portExample(s tsb.Server, jack int) {
}

func i2cExample(s tsb.Server, jack int) {
}
