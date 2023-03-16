package tsb

const (
	UartBaudAuto uint16 = iota
	UartBaud1200
	UartBaud2400
	UartBaud4800
	UartBaud9600
	UartBaud19200
	UartBaud38400
	UartBaud57600
	UartBaud115200
	UartBaud230400
	UartBaud460800
	UartBaud921600
	UartBaud1000000
	UartBaud1500000
	UartBaud3000000
)

const (
	UartStopbits1 uint16 = iota << 8
	UartStopbits15
	UartStopbits2
)

const (
	UartParityNone uint16 = iota << 10
	UartParityEven
	UartParityOdd
)

const (
	UartData8 uint16 = iota << 12
	UartData9
	UartData7
	UartData6
	UartData5
)

// UART represents a connection to UART-device.
type UART struct {
	Jack   byte
	Server Server
}

// NewUart opens a connection.
func NewUart(jack byte, server Server) (*UART, error) {
	CheckJack(jack)
	err := ModbusWriteSingleRegister(ModeRegisterAdr, jack, server, RegModeValueUart)
	if err != nil {
		return nil, err
	}
	uart := &UART{Server: server, Jack: jack}
	return uart, nil
}

// Config ?
func (u *UART) Config(baud uint16, databits uint16, parity uint16, stopbits uint16) error {
	err := ModbusWriteSingleRegister(UartRegisterAdr, u.Jack, u.Server, baud|databits|parity|stopbits)
	if err != nil {
		return err
	}
	return nil
}

// Write writes a buffer to the uart
func (u *UART) Write(b []byte) (n int, err error) {
	td := TsbData{Ch: []byte{byte(u.Jack)}, Typ: []byte{TypRaw}, Payload: b}
	u.Server.tdPutCh <- td
	return len(b), nil
}

// Read reads a buffer from the uart
func (u *UART) Read(b []byte) (n int, err error) {
	b[0] = <-u.Server.Jack[u.Jack].ReadChan[TypRaw]
	n = len(u.Server.Jack[u.Jack].ReadChan[TypRaw]) + 1
	if n > len(b) {
		n = len(b)
	}
	for i := 1; i < n; i++ {
		b[i] = <-u.Server.Jack[u.Jack].ReadChan[TypRaw]
	}
	return n, nil
}
