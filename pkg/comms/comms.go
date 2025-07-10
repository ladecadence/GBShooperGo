package comms

import (
	"errors"
	"fmt"
	"time"

	"github.com/ziutek/ftdi"
)

const (
	ID_MANUFACTURER = "ladecadence.net"
	ID_PRODUCT      = "GB Flasher"
	SEND_DELAY      = 50
	BAUDRATE_115_2K = 115200
	BAUDRATE_230_4K = 230400
	BAUDRATE_1M     = 1000000

	// Packet types
	TYPE_COMMAND = 0x11
	TYPE_DATA    = 0x22
	TYPE_STAT    = 0x33
	TYPE_INFO    = 0x44
)

type Packet struct {
	Type uint8
	Data uint8
}

type GBSDevice struct {
	Dev *ftdi.Device
}

func (gbs *GBSDevice) Open() error {
	list, err := ftdi.FindAll(0x0403, 0x6001)
	if err != nil {
		return err
	}
	found := false
	for _, d := range list {
		if d.Manufacturer == ID_MANUFACTURER && d.Description == ID_PRODUCT {
			gbs.Dev, err = ftdi.OpenUSBDev(d, ftdi.ChannelAny)
			if err != nil {
				return err
			}
			found = true
		}
	}
	if found {
		gbs.Dev.SetBaudrate(BAUDRATE_230_4K)
		gbs.Dev.SetFlowControl(ftdi.FlowCtrlDisable)
		gbs.Dev.SetLineProperties(8, 1, ftdi.ParityNone)
		return nil
	} else {
		return errors.New("No device found")
	}
}

func (gbs *GBSDevice) Close() error {
	return gbs.Dev.Close()
}

func (gbs *GBSDevice) SendByte(data uint8) error {
	err := gbs.Dev.WriteByte(data)
	time.Sleep(time.Microsecond * SEND_DELAY)
	return err
}

func (gbs *GBSDevice) SendPacket(packet Packet) error {
	err := gbs.SendByte(packet.Type)
	time.Sleep(time.Microsecond * SEND_DELAY)
	err = gbs.SendByte(packet.Data)
	return err
}

func (gbs *GBSDevice) ReceiveByte(timeout time.Duration) (uint8, error) {
	var data uint8
	var err error

	// timeout
	received := false
	for start := time.Now(); time.Since(start) < (timeout * time.Second); {
		data, err = gbs.Dev.ReadByte()
		if err == nil {
			received = true
			break
		}
	}
	if received {
		fmt.Printf("Byte: %x\n", data)
		return data, nil
	} else {
		return 0, errors.New("Timeout")
	}
}

func (gbs *GBSDevice) ReceivePacket(timeout time.Duration) (Packet, error) {
	var packet Packet
	var data []uint8 = make([]uint8, 2)

	remaining := 2
	for start := time.Now(); time.Since(start) < (timeout * time.Second); {
		num, _ := gbs.Dev.Read(data)
		remaining -= num
		if remaining == 0 {
			break
		}
	}

	if remaining > 0 {
		return Packet{}, errors.New("Timeout")
	}
	packet.Type = data[0]
	packet.Data = data[1]

	return packet, nil
}

func (gbs *GBSDevice) SendBuffer(buffer []uint8) error {
	_, err := gbs.Dev.Write(buffer)
	return err
}
