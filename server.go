package tsb

import (
	"fmt"
	"log"
	"net"

	"github.com/tarm/serial"
)

const (
	MaxJacks    byte = 8
	JackModeReg byte = 0x80
	JackUartReg byte = 0x82
	JackPortReg byte = 0x86
	JackI2cReg  byte = 0x88
)

const (
	JackPort byte = 1
	JackI2c
	JackUart
	JackSpi
)

type jack struct {
	ReadChan [MaxTyp + 1]chan byte
}

type Server struct {
	Adr      string
	Typ      string
	callback [MaxJacks + 1]map[byte]func(data []byte)
	Jack     [MaxJacks + 1]jack
	conn     net.Conn
	sport    *serial.Port
	tdPutCh  chan TsbData
	tdGetCh  chan TsbData
	done     chan struct{}
}

func NewSerialServer(adr string) (Server, error) {
	var err error
	s := Server{Adr: adr}
	s.Typ = "Serial"
	s.sport, err = serial.OpenPort(&serial.Config{Name: adr, Baud: 115200})
	if err != nil {
		log.Fatal(err)
	}
	s.tdPutCh = PutData(s.sport, 200)
	s.tdGetCh, s.done = GetData(s.sport, 200)
	s.serv()
	return s, nil
}

func NewTcpServer(adr string) (Server, error) {
	var err error
	s := Server{Adr: adr}
	s.Typ = "TCP"
	s.conn, err = net.Dial("tcp", adr)
	if err != nil {
		log.Fatal(err)
	}
	s.tdPutCh = PutData(s.conn, 200)
	s.tdGetCh, s.done = GetData(s.conn, 200)
	s.serv()
	return s, nil
}

func (s Server) Close() {
	close(s.tdPutCh)
	close(s.done)
}

func (s Server) SetCallback(jack byte, typ byte, f func(payload []byte)) {
	CheckJack(jack)
	if s.callback[jack] == nil {
		s.callback[jack] = make(map[byte]func(data []byte))
	}
	s.callback[jack][typ] = f
	s.callback[jack][typ]([]byte{0x65, 0x65, 0x65, 0x65, 0x65, 0x65})
	fmt.Printf("Callback set for Jack: %d, Typ: %d\n", jack, typ)
}

func (s Server) serv() {
	for i := 0; i <= int(MaxJacks); i++ {
		s.Jack[i].ReadChan[TypI2c] = make(chan byte, 1024)
		s.Jack[i].ReadChan[TypPort] = make(chan byte, 1024)
		s.Jack[i].ReadChan[TypRaw] = make(chan byte, 1024)
		s.Jack[i].ReadChan[TypError] = make(chan byte, 1024)
		s.Jack[i].ReadChan[TypModbus] = make(chan byte, 1024)
	}
	fmt.Printf("TSB client connected to tsb server: %s\n", s.Adr)
	go func(s Server) {
		for {
			select {
			case td := <-s.tdGetCh:
				{
					fmt.Printf("td: ch: %d, typ: %s, %x\n", td.Ch[0], TypLabel[td.Typ[0]], td.Payload)
					if td.Typ[0] > MaxTyp {
						//log.Printf("Invalid Typ %d!\n\r", td.Typ[0])
						break
					}
					if td.Ch[0] > MaxJacks {
						//log.Printf("Invalid Jacknr %d!\n\r", td.Ch[0])
						break
					}
					if s.Jack[td.Ch[0]].ReadChan[td.Typ[0]] == nil {
						log.Printf("Channel: %d, Type: %d is not initialized!\n\r", td.Ch[0], td.Typ[0])
						break
					}
					if len(s.Jack[td.Ch[0]].ReadChan[td.Typ[0]]) > 800 {
						log.Printf("Read Channel Overflow! Jack: %d, Typ: %d, cap: %d, len: %d", td.Ch[0], td.Typ[0],
							cap(s.Jack[td.Ch[0]].ReadChan[td.Typ[0]]), len(s.Jack[td.Ch[0]].ReadChan[td.Typ[0]]))
					}
					if s.callback[td.Ch[0]][td.Typ[0]] != nil {
						fmt.Printf("Callback called for Jack: %d, Typ: %d\n", td.Ch[0], td.Typ[0])
						s.callback[td.Ch[0]][td.Typ[0]](td.Payload)
					} else {
						fmt.Printf("No callback for Jack: %d, Typ: %d callback: %v\n", td.Ch[0], td.Typ[0], s.callback[td.Ch[0]])
					}
					for i := range td.Payload {
						s.Jack[td.Ch[0]].ReadChan[td.Typ[0]] <- td.Payload[i]
						//fmt.Printf("Ch: %d, Typ: %d, Data: %02x\n", td.Ch[0], td.Typ[0], td.Payload[i])
					}
				}
			case <-s.done:
				{
					fmt.Printf("TSB client connection closed!\n")
					return
				}
			}
		}
	}(s)
}

func (s Server) SpiInit(jack byte) (err error) {
	CheckJack(jack)
	return nil
}

func CheckJack(jack byte) {
	if jack > MaxJacks {
		log.Fatalf("Illegal Jack nr: %d", jack)
	}
}
