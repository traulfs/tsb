/*
Package tsb implements the encoding and decoding of tsb protocol.
It defines the types of tsb.
*/
package tsb

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// TsbData implements the tsb data structure
type TsbData struct {
	Ch      []byte
	Typ     []byte
	Payload []byte
}

// Channel defintions
const (
	Buflen     int  = 1000
	MaxTyp     byte = 127
	MaxPayload int  = 250

	TypUnused     byte = 0x00
	TypRaw        byte = 0x01
	TypText       byte = 0x02
	TypPort       byte = 0x03
	TypI2c        byte = 0x04
	TypSpi        byte = 0x05
	TypModbus     byte = 0x07
	TypM2go       byte = 0x08
	TypAtCmd      byte = 0x09
	TypBline      byte = 0x11
	TypBline2     byte = 0x12
	TypHci        byte = 0x15
	TypCWB        byte = 0x16
	TypCoap       byte = 0x21
	TypCbor       byte = 0x31
	TypCan        byte = 0x41
	TypSenmlJSON  byte = 0x6E
	TypSensmlJSON byte = 0x6F
	TypSenmlCbor  byte = 0x70
	TypSensmlCbor byte = 0x71
	TypSenmlExi   byte = 0x72
	TypSensmlExi  byte = 0x73
	TypInflux     byte = 0x75
	TypLog        byte = 0x7d
	TypWarning    byte = 0x7e
	TypError      byte = 0x7f
)

// TypLabel maps a string to each type
var TypLabel = map[byte]string{
	TypUnused:     "unused",
	TypRaw:        "raw",
	TypText:       "text",
	TypPort:       "port",
	TypI2c:        "i2c",
	TypSpi:        "spi",
	TypModbus:     "modbus",
	TypM2go:       "m2go",
	TypAtCmd:      "atcmd",
	TypBline:      "bline",
	TypBline2:     "bline2",
	TypHci:        "hci",
	TypCWB:        "CWB",
	TypCoap:       "coap",
	TypCbor:       "cbor",
	TypCan:        "can",
	TypSenmlJSON:  "senml_json",
	TypSensmlJSON: "sensml_json",
	TypSenmlCbor:  "senml_cbor",
	TypSensmlCbor: "sensml_cbor",
	TypSenmlExi:   "senml_exi",
	TypSensmlExi:  "sensml_exi",
	TypInflux:     "influx",
	TypLog:        "log",
	TypWarning:    "warning",
	TypError:      "error",
}

// Verbose is a switch for more debug outputs
var Verbose bool = false

var ErrorVerbose bool = false

// Channel2Bytes converts a channel string in a []byte
// Example: "3.4.5" -> 0x83,0x84,0x05
func Channel2Bytes(ch string) []byte {
	buf := new(bytes.Buffer)
	if len(ch) < 1 {
		buf.WriteByte(0)
		return buf.Bytes()
	}
	routes := strings.Split(ch, ".")
	for i := 0; i < len(routes); i++ {
		channel, err := strconv.Atoi(routes[i])
		if err != nil {
			log.Printf("Invalid channel string: %s", ch)
		}
		if i < len(routes)-1 {
			buf.WriteByte(byte(channel + 128))
		} else {
			buf.WriteByte(byte(channel))
		}
	}
	return buf.Bytes()
}

// TEncode encodes tsb
func Encode(td TsbData) []byte {
	buf := new(bytes.Buffer)
	buf.Write(td.Ch)
	buf.Write(td.Typ)
	buf.Write(td.Payload)
	crc := checkSum(buf.Bytes())
	buf.WriteByte(byte(crc & 0xff))
	buf.WriteByte(byte(crc >> 8))
	return buf.Bytes()
}

// Decode encodes tsb
func Decode(packet []byte) (TsbData, error) {
	l := len(packet)
	td := TsbData{}
	c := 0
	t := 0
	if l < 4 {
		return td, fmt.Errorf("invalid tsb packet length (%d)", l)
	}
	for packet[c] > 127 {
		c++
		if c+3 > l {
			return td, fmt.Errorf("invalid tsb packet length (%d)", l)
		}
	}
	c++
	td.Ch = packet[0:c]
	t = c
	for packet[t] > 127 {
		t++
		if t+2 > l {
			return td, fmt.Errorf("invalid tsb packet length (%d)", l)
		}
	}
	t++
	td.Typ = packet[c:t]
	if len(packet)-2 < t {
		return td, fmt.Errorf("tsb invalid packet error")
	}
	td.Payload = packet[t : len(packet)-2]
	crc := checkSum(packet[0 : len(packet)-2])
	if byte(crc>>8) != packet[len(packet)-1] || byte(crc&0xff) != packet[len(packet)-2] {
		//fmt.Printf("TSB-Read:\tCrc error! packet= % X, crc=% X\n", packet, crc)
		//return td, fmt.Errorf("tsb crc error, packet= % X, crc=% X", packet, crc)
		return td, fmt.Errorf("tsb crc error")
	} else {
		if Verbose {
			fmt.Printf("TSB-Read:  Ch: 0x%X Typ: 0x%X Payload 0x% X\n", td.Ch, td.Typ, td.Payload)
		}
	}
	return td, nil
}

// CobsEncode implements the cobs algorithmus
func CobsEncode(p []byte) []byte {
	buf := new(bytes.Buffer)
	writeBlock := func(p []byte) {
		buf.WriteByte(byte(len(p) + 1))
		buf.Write(p)
	}
	for _, ch := range bytes.Split(p, []byte{0}) {
		for len(ch) > 0xfe {
			writeBlock(ch[:0xfe])
			ch = ch[0xfe:]
		}
		writeBlock(ch)
	}
	buf.WriteByte(0)
	return buf.Bytes()
}

// CobsDecode implements the cobs algorithmus
func CobsDecode(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty packet")
	}
	buf := new(bytes.Buffer)
	for n := b[0]; n > 0; n = b[0] {
		if int(n) >= len(b) {
			return nil, fmt.Errorf("cobs length byte (%d) is bigger or equal than remaining packet length (%d)", n, len(b))
		}
		buf.Write(b[1:n])
		b = b[n:]
		if n < 0xff && b[0] > 0 {
			buf.WriteByte(0)
		}
	}
	return buf.Bytes(), nil
}

// GetData reads tsb data from io.Reader and puts it in a channel
func GetData(r io.Reader) (chan TsbData, chan struct{}) {
	c := make(chan TsbData, 100)
	done := make(chan struct{})

	go func() {
		var i, k int
		wbuf := []byte{}
		rbuf := make([]byte, 10000)
		for {
			n, err := r.Read(rbuf)
			if err != nil {
				/*
					if err != io.EOF {
						// log.Fatal(err) funktioniert nicht unter Windows
					}
				*/
				break
			}
			k = 0
			for i = 0; i < n; i++ {
				if rbuf[i] == 0x00 {
					wbuf = append(wbuf, rbuf[k:i+1]...)
					k = i + 1
					packet, err := CobsDecode(wbuf)
					if err != nil {
						if ErrorVerbose {
							log.Print(err)
							fmt.Printf("\tCobsDecode packet:\t%x\n", wbuf)
						}
					} else {
						td, err := Decode(packet)
						if err != nil {
							if ErrorVerbose {
								log.Print(err)
								fmt.Printf("\tCobsDecode packet:\t%x\n", wbuf)
								fmt.Printf("\tDecode packet:\t\t%x\n", packet)
							}
						} else {
							c <- td
						}
					}
					wbuf = []byte{}
				}
			}
			wbuf = append(wbuf, rbuf[k:n]...)
		}
		done <- struct{}{}
	}()
	return c, done
}

// PutData reads tsb data from a channel and writes it to the io.Writer
func PutData(w io.Writer) chan TsbData {
	c := make(chan TsbData, 10)
	go func() {
		for {
			td := <-c
			out := CobsEncode(Encode(td))
			_, err := w.Write(out)
			if Verbose {
				fmt.Printf("TSB-Write: Ch: 0x%X Typ: 0x%X Payload 0x% X\n", td.Ch, td.Typ, td.Payload)
			}
			if err != nil {
				log.Fatal(err)
			}

		}
	}()
	return c
}

// GetTypList makes a string of all available Typs
func GetTypList() string {
	var s string
	for i, name := range TypLabel {
		s += fmt.Sprintf("\n\t0x%2X: %s", int(i), name)
	}
	return s
}
