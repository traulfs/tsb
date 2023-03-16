package tsb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	MbFcReadHoldingRegister   byte = 0x03
	MbFcWriteSingleRegister   byte = 0x06
	MbFcWriteMultipleRegister byte = 0x10
)

const (
	ModeRegisterAdr uint16 = 0x0002
	PortRegisterAdr uint16 = 0x0004
	UartRegisterAdr uint16 = 0x0006
	I2cRegisterAdr  uint16 = 0x0008
	SpiRegisterAdr  uint16 = 0x000a
)

const (
	RegModeValuePort uint16 = 1
	RegModeValueI2c  uint16 = 2
	RegModeValueUart uint16 = 3
	RegModeValueSpi  uint16 = 4
)

const (
	MbExceptionResponseFlag uint8 = 0x80
)

// ModbusWriteSingleRegister macht was?
func ModbusWriteSingleRegister(adr uint16, jack byte, server Server, value uint16) error {
	w := []byte{MbFcWriteSingleRegister, byte(adr >> 8), byte(adr), byte(value >> 8), byte(value)}
	td := TsbData{Ch: []byte{byte(jack)}, Typ: []byte{TypModbus}, Payload: w}
	server.tdPutCh <- td

	receiveCount := len(w)
	r := make([]byte, receiveCount)
	for i := 0; i < receiveCount; i++ {
		select {
		case r[i] = <-server.Jack[jack].ReadChan[TypModbus]:
			if i == 0 && (r[0]&MbExceptionResponseFlag) == MbExceptionResponseFlag {
				receiveCount = 2 // exception bit set, so only receive code
			}
		case <-time.After(1 * time.Second):
			return fmt.Errorf("timeout")
		}
	}

	if receiveCount == 2 {
		return fmt.Errorf("exception code: %d", r[1])
	} else {
		if !bytes.Equal(w, r) {
			return fmt.Errorf("invalid response: %s", hex.EncodeToString(r))
		}
		return nil
	}
}
