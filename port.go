package tsb

const (
	PortPad0    byte = 1
	PortPad1    byte = 2
	PortPad2    byte = 4
	PortPad3    byte = 8
	PortAllPads byte = 15
)

const (
	PortcharReadWrite         = 0x00
	PortcharRead              = 0x01
	PortcharSetOutput         = 0x02
	PortcharClearOutput       = 0x03
	PortcharToggleOutput      = 0x04
	PortcharNotification      = 0x05
	PortcharDelay             = 0x06
	PortcharSetDirection      = 0x08
	PortcharClearDirection    = 0x09
	PortcharSetPullEnable     = 0x0A
	PortcharClearPullEnable   = 0x0B
	PortcharSetNotification   = 0x0C
	PortcharClearNotification = 0x0D
	PortcharSetLED            = 0x10
	PortcharClearLED          = 0x11
	PortcharToggleLED         = 0x12
)

// I2C represents a connection to I2C-device.
type Port struct {
	Jack   byte
	Server *Server
}

// NewI2C opens a connection for I2C-device.
func NewPort(jack byte, server *Server) (*Port, error) {
	CheckJack(jack)
	port := &Port{Server: server, Jack: jack}
	/*
		err := ModbusWriteSingleRegister(ModeRegisterAdr, jack, server, RegModeValuePort)
		if err != nil {
			return nil, err
		}
	*/
	return port, nil
}

// Write writes a buffer
func (p *Port) Write(b []byte) (n int, err error) {
	td := TsbData{Ch: []byte{byte(p.Jack)}, Typ: []byte{TypPort}, Payload: b}
	p.Server.tdPutCh <- td
	return len(b), nil
}

// Read reads a buffer
func (p *Port) Read(b []byte) (n int, err error) {
	n = len(p.Server.Jack[p.Jack].ReadChan[TypPort])
	if n > len(b) {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		b[i] = <-p.Server.Jack[p.Jack].ReadChan[TypPort]
	}
	return n, nil
}

func PortCharNibble(code byte, value int) []byte {
	switch code {
	case PortcharReadWrite:
		fallthrough
	case PortcharRead:
		fallthrough
	case PortcharSetOutput:
		fallthrough
	case PortcharClearOutput:
		fallthrough
	case PortcharToggleOutput:
		fallthrough
	case PortcharNotification:
		fallthrough
	case PortcharDelay:
		return []byte{code<<4 | (byte(value) & 0x0f)}
	case PortcharSetDirection:
		fallthrough
	case PortcharClearDirection:
		fallthrough
	case PortcharSetPullEnable:
		fallthrough
	case PortcharClearPullEnable:
		fallthrough
	case PortcharSetNotification:
		fallthrough
	case PortcharClearNotification:
		fallthrough
	case PortcharSetLED:
		fallthrough
	case PortcharClearLED:
		fallthrough
	case PortcharToggleLED:
		return []byte{0xe0 | ((code >> 2) & 0x0f), 0x80 | ((code & 0x03) << 4), 0x80, 0x80 | (byte(value) & 0x0f)}
	}
	return nil
}
